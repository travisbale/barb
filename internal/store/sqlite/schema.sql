CREATE TABLE IF NOT EXISTS smtp_profiles (
    id          TEXT    PRIMARY KEY,
    name        TEXT    NOT NULL,
    host        TEXT    NOT NULL,
    port        INTEGER NOT NULL DEFAULT 587,
    username    TEXT    NOT NULL DEFAULT '',
    password    TEXT    NOT NULL DEFAULT '',
    from_addr   TEXT    NOT NULL,
    from_name       TEXT    NOT NULL DEFAULT '',
    custom_headers  TEXT    NOT NULL DEFAULT '{}',
    created_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS email_templates (
    id          TEXT    PRIMARY KEY,
    name        TEXT    NOT NULL,
    subject     TEXT    NOT NULL,
    html_body   TEXT    NOT NULL DEFAULT '',
    text_body   TEXT    NOT NULL DEFAULT '',
    created_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS target_lists (
    id          TEXT    PRIMARY KEY,
    name        TEXT    NOT NULL,
    created_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS targets (
    id          TEXT    PRIMARY KEY,
    list_id     TEXT    NOT NULL REFERENCES target_lists(id) ON DELETE CASCADE,
    email       TEXT    NOT NULL,
    first_name  TEXT    NOT NULL DEFAULT '',
    last_name   TEXT    NOT NULL DEFAULT '',
    department  TEXT    NOT NULL DEFAULT '',
    position    TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_targets_list ON targets(list_id);

CREATE TABLE IF NOT EXISTS campaigns (
    id              TEXT    PRIMARY KEY,
    name            TEXT    NOT NULL,
    status          TEXT    NOT NULL DEFAULT 'draft',
    template_id     TEXT    NOT NULL REFERENCES email_templates(id),
    smtp_profile_id TEXT    NOT NULL REFERENCES smtp_profiles(id),
    target_list_id  TEXT    NOT NULL REFERENCES target_lists(id),
    miraged_id      TEXT    NOT NULL DEFAULT '',
    phishlet        TEXT    NOT NULL DEFAULT '',
    lure_url        TEXT    NOT NULL DEFAULT '',
    send_rate       INTEGER NOT NULL DEFAULT 10,
    created_at      INTEGER NOT NULL,
    started_at      INTEGER,
    completed_at    INTEGER
);

CREATE TABLE IF NOT EXISTS campaign_results (
    id          TEXT    PRIMARY KEY,
    campaign_id TEXT    NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    target_id   TEXT    NOT NULL REFERENCES targets(id),
    email       TEXT    NOT NULL,
    status      TEXT    NOT NULL DEFAULT 'pending',
    sent_at     INTEGER,
    clicked_at  INTEGER,
    captured_at INTEGER,
    session_id  TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_results_campaign ON campaign_results(campaign_id);
CREATE INDEX IF NOT EXISTS idx_results_campaign_email ON campaign_results(campaign_id, email);
CREATE INDEX IF NOT EXISTS idx_campaigns_miraged_status ON campaigns(miraged_id, status);

CREATE TABLE IF NOT EXISTS phishlets (
    id          TEXT    PRIMARY KEY,
    name        TEXT    NOT NULL UNIQUE,
    yaml        TEXT    NOT NULL,
    created_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS miraged_connections (
    id              TEXT    PRIMARY KEY,
    name            TEXT    NOT NULL,
    address         TEXT    NOT NULL,
    secret_hostname TEXT    NOT NULL,
    cert_pem        BLOB    NOT NULL,
    key_pem         BLOB    NOT NULL,
    ca_cert_pem     BLOB    NOT NULL,
    created_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id                       TEXT    PRIMARY KEY,
    username                 TEXT    NOT NULL UNIQUE,
    password_hash            TEXT    NOT NULL,
    password_change_required INTEGER NOT NULL DEFAULT 0,
    created_at               INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT    PRIMARY KEY,
    user_id    TEXT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at INTEGER NOT NULL
);
