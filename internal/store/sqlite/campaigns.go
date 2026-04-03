package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/travisbale/barb/internal/phishing"
)

type Campaigns struct{ db *DB }

func NewCampaignStore(db *DB) *Campaigns { return &Campaigns{db: db} }

func (s *Campaigns) CreateCampaign(c *phishing.Campaign) error {
	_, err := s.db.db.Exec(
		`INSERT INTO campaigns (id, name, status, template_id, smtp_profile_id, target_list_id, miraged_id, phishlet, redirect_url, lure_id, lure_url, send_rate, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.Name, c.Status, c.TemplateID, c.SMTPProfileID, c.TargetListID, c.MiragedID, c.Phishlet, c.RedirectURL, c.LureID, c.LureURL, c.SendRate, c.CreatedAt.Unix(),
	)
	if isConflict(err) {
		return phishing.ErrConflict
	}
	return err
}

func (s *Campaigns) GetCampaign(id string) (*phishing.Campaign, error) {
	row := s.db.db.QueryRow(
		`SELECT id, name, status, template_id, smtp_profile_id, target_list_id, miraged_id, phishlet, redirect_url, lure_id, lure_url, send_rate, created_at, started_at, completed_at
		 FROM campaigns WHERE id = ?`, id,
	)
	return scanCampaign(row)
}

func (s *Campaigns) UpdateCampaign(c *phishing.Campaign) error {
	res, err := s.db.db.Exec(
		`UPDATE campaigns SET name = ?, status = ?, template_id = ?, smtp_profile_id = ?, target_list_id = ?, miraged_id = ?, phishlet = ?, redirect_url = ?, lure_id = ?, lure_url = ?, send_rate = ?, started_at = ?, completed_at = ?
		 WHERE id = ?`,
		c.Name, c.Status, c.TemplateID, c.SMTPProfileID, c.TargetListID, c.MiragedID, c.Phishlet, c.RedirectURL, c.LureID, c.LureURL, c.SendRate, timeToUnix(c.StartedAt), timeToUnix(c.CompletedAt), c.ID,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *Campaigns) DeleteCampaign(id string) error {
	res, err := s.db.db.Exec(`DELETE FROM campaigns WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *Campaigns) ListCampaigns() ([]*phishing.Campaign, error) {
	rows, err := s.db.db.Query(
		`SELECT id, name, status, template_id, smtp_profile_id, target_list_id, miraged_id, phishlet, redirect_url, lure_id, lure_url, send_rate, created_at, started_at, completed_at
		 FROM campaigns ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*phishing.Campaign
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Campaigns) CreateResults(results []*phishing.CampaignResult) error {
	return s.db.WithTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(
			`INSERT INTO campaign_results (id, campaign_id, target_id, email, status)
			 VALUES (?, ?, ?, ?, ?)`,
		)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, r := range results {
			if _, err := stmt.Exec(r.ID, r.CampaignID, r.TargetID, r.Email, r.Status); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Campaigns) UpdateResult(r *phishing.CampaignResult) error {
	res, err := s.db.db.Exec(
		`UPDATE campaign_results SET status = ?, sent_at = ?, clicked_at = ?, captured_at = ?, session_id = ?
		 WHERE id = ?`,
		r.Status, timeToUnix(r.SentAt), timeToUnix(r.ClickedAt), timeToUnix(r.CapturedAt), r.SessionID, r.ID,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *Campaigns) ListResults(campaignID string) ([]*phishing.CampaignResult, error) {
	rows, err := s.db.db.Query(
		`SELECT id, campaign_id, target_id, email, status, sent_at, clicked_at, captured_at, session_id
		 FROM campaign_results WHERE campaign_id = ? ORDER BY email ASC`, campaignID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*phishing.CampaignResult
	for rows.Next() {
		var (
			r                             phishing.CampaignResult
			sentAt, clickedAt, capturedAt sql.NullInt64
		)
		if err := rows.Scan(&r.ID, &r.CampaignID, &r.TargetID, &r.Email, &r.Status, &sentAt, &clickedAt, &capturedAt, &r.SessionID); err != nil {
			return nil, err
		}
		r.SentAt = unixToTime(sentAt)
		r.ClickedAt = unixToTime(clickedAt)
		r.CapturedAt = unixToTime(capturedAt)
		out = append(out, &r)
	}
	return out, rows.Err()
}

func (s *Campaigns) GetResult(id string) (*phishing.CampaignResult, error) {
	row := s.db.db.QueryRow(
		`SELECT id, campaign_id, target_id, email, status, sent_at, clicked_at, captured_at, session_id
		 FROM campaign_results WHERE id = ?`, id,
	)
	var (
		result                        phishing.CampaignResult
		sentAt, clickedAt, capturedAt sql.NullInt64
	)
	err := row.Scan(&result.ID, &result.CampaignID, &result.TargetID, &result.Email, &result.Status, &sentAt, &clickedAt, &capturedAt, &result.SessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, phishing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	result.SentAt = unixToTime(sentAt)
	result.ClickedAt = unixToTime(clickedAt)
	result.CapturedAt = unixToTime(capturedAt)
	return &result, nil
}

func scanCampaign(row scanner) (*phishing.Campaign, error) {
	var (
		c                      phishing.Campaign
		createdAt              int64
		startedAt, completedAt sql.NullInt64
	)
	err := row.Scan(&c.ID, &c.Name, &c.Status, &c.TemplateID, &c.SMTPProfileID, &c.TargetListID, &c.MiragedID, &c.Phishlet, &c.RedirectURL, &c.LureID, &c.LureURL, &c.SendRate, &createdAt, &startedAt, &completedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, phishing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c.CreatedAt = time.Unix(createdAt, 0)
	c.StartedAt = unixToTime(startedAt)
	c.CompletedAt = unixToTime(completedAt)
	return &c, nil
}

func timeToUnix(t *time.Time) sql.NullInt64 {
	if t == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: t.Unix(), Valid: true}
}

func unixToTime(n sql.NullInt64) *time.Time {
	if !n.Valid {
		return nil
	}
	t := time.Unix(n.Int64, 0)
	return &t
}
