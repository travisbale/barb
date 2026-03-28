package phishing

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TargetList is a named collection of targets that can be assigned to campaigns.
type TargetList struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

// Target is an individual phishing target within a list.
type Target struct {
	ID         string
	ListID     string
	Email      string
	FirstName  string
	LastName   string
	Department string
	Position   string
}

type targetStore interface {
	CreateList(list *TargetList) error
	GetList(id string) (*TargetList, error)
	DeleteList(id string) error
	ListLists() ([]*TargetList, error)
	CreateTarget(target *Target) error
	ListTargets(listID string) ([]*Target, error)
	DeleteTarget(id string) error
}

// TargetService manages target lists and their members.
type TargetService struct {
	Store targetStore
}

func (s *TargetService) CreateList(name string) (*TargetList, error) {
	if name == "" {
		return nil, ErrNameRequired
	}
	list := &TargetList{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedAt: time.Now(),
	}
	if err := s.Store.CreateList(list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *TargetService) GetList(id string) (*TargetList, error) {
	return s.Store.GetList(id)
}

func (s *TargetService) DeleteList(id string) error {
	return s.Store.DeleteList(id)
}

func (s *TargetService) ListLists() ([]*TargetList, error) {
	return s.Store.ListLists()
}

func (s *TargetService) AddTarget(listID string, target *Target) error {
	if target.Email == "" {
		return ErrEmailRequired
	}
	target.ID = uuid.New().String()
	target.ListID = listID
	return s.Store.CreateTarget(target)
}

func (s *TargetService) ListTargets(listID string) ([]*Target, error) {
	return s.Store.ListTargets(listID)
}

func (s *TargetService) DeleteTarget(id string) error {
	return s.Store.DeleteTarget(id)
}

// ImportCSV reads a CSV from r and adds all rows as targets to the given list.
// The first row must be a header. Recognized columns (case-insensitive):
// email, first_name/first name/firstname, last_name/last name/lastname,
// department, position. Unrecognized columns are ignored.
// Returns the number of targets imported.
func (s *TargetService) ImportCSV(listID string, r io.Reader) (int, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("reading CSV header: %w", err)
	}

	colMap := mapColumns(header)
	if _, ok := colMap["email"]; !ok {
		return 0, fmt.Errorf("CSV must have an 'email' column")
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, fmt.Errorf("reading CSV row %d: %w", count+2, err)
		}

		target := &Target{
			Email:      getCol(record, colMap, "email"),
			FirstName:  getCol(record, colMap, "first_name"),
			LastName:   getCol(record, colMap, "last_name"),
			Department: getCol(record, colMap, "department"),
			Position:   getCol(record, colMap, "position"),
		}

		if target.Email == "" {
			continue
		}

		if err := s.AddTarget(listID, target); err != nil {
			return count, fmt.Errorf("importing row %d: %w", count+2, err)
		}
		count++
	}

	return count, nil
}

// mapColumns maps normalized header names to column indices.
func mapColumns(header []string) map[string]int {
	aliases := map[string]string{
		"email":      "email",
		"first_name": "first_name",
		"first name": "first_name",
		"firstname":  "first_name",
		"last_name":  "last_name",
		"last name":  "last_name",
		"lastname":   "last_name",
		"department": "department",
		"position":   "position",
		"title":      "position",
	}

	m := make(map[string]int)
	for i, col := range header {
		key := strings.ToLower(strings.TrimSpace(col))
		if canonical, ok := aliases[key]; ok {
			m[canonical] = i
		}
	}
	return m
}

func getCol(record []string, colMap map[string]int, name string) string {
	idx, ok := colMap[name]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}
