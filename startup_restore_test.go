package main

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func openTestSQLiteDB(t *testing.T) *sqlx.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := sqlx.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_busy_timeout=3000")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := db.Ping(); err != nil {
		t.Fatalf("ping test db: %v", err)
	}

	return db
}

func TestInitializeSchemaAddsAutostartColumnAndBackfills(t *testing.T) {
	db := openTestSQLiteDB(t)

	_, err := db.Exec(`
		CREATE TABLE migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			token TEXT NOT NULL,
			webhook TEXT NOT NULL DEFAULT '',
			jid TEXT NOT NULL DEFAULT '',
			qrcode TEXT NOT NULL DEFAULT '',
			connected INTEGER,
			expiration INTEGER,
			events TEXT NOT NULL DEFAULT '',
			proxy_url TEXT DEFAULT ''
		);
	`)
	if err != nil {
		t.Fatalf("create legacy schema: %v", err)
	}

	for i, name := range []string{
		"initial_schema",
		"add_proxy_url",
		"change_id_to_string",
		"add_s3_support",
		"add_message_history",
		"add_quoted_message_id",
		"add_hmac_key",
		"add_data_json",
		"add_user_webhooks",
		"add_s3_failover_support",
	} {
		if _, err := db.Exec(`INSERT INTO migrations (id, name) VALUES (?, ?)`, i+1, name); err != nil {
			t.Fatalf("seed migrations: %v", err)
		}
	}

	_, err = db.Exec(`
		INSERT INTO users (id, name, token, webhook, jid, qrcode, connected, expiration, events, proxy_url)
		VALUES
			('paired-active', 'Paired Active', 'token-1', '', '111@s.whatsapp.net', '', 1, 0, '', ''),
			('paired-inactive', 'Paired Inactive', 'token-2', '', '222@s.whatsapp.net', '', 0, 0, '', ''),
			('unpaired-active', 'Unpaired Active', 'token-3', '', '', '', 1, 0, '', '')
	`)
	if err != nil {
		t.Fatalf("seed users: %v", err)
	}

	if err := initializeSchema(db); err != nil {
		t.Fatalf("initialize schema: %v", err)
	}

	type row struct {
		ID        string `db:"id"`
		Autostart int    `db:"autostart"`
	}

	var rows []row
	if err := db.Select(&rows, `SELECT id, autostart FROM users ORDER BY id`); err != nil {
		t.Fatalf("select autostart rows: %v", err)
	}

	got := map[string]int{}
	for _, row := range rows {
		got[row.ID] = row.Autostart
	}

	if got["paired-active"] != 1 {
		t.Fatalf("expected paired-active autostart=1, got %d", got["paired-active"])
	}
	if got["paired-inactive"] != 0 {
		t.Fatalf("expected paired-inactive autostart=0, got %d", got["paired-inactive"])
	}
	if got["unpaired-active"] != 0 {
		t.Fatalf("expected unpaired-active autostart=0, got %d", got["unpaired-active"])
	}
}

func TestLoadStartupUsersUsesAutostartAndRequiresJID(t *testing.T) {
	db := openTestSQLiteDB(t)

	if err := initializeSchema(db); err != nil {
		t.Fatalf("initialize schema: %v", err)
	}

	_, err := db.Exec(`
		INSERT INTO users (
			id, name, token, webhook, jid, qrcode, connected, expiration, events, proxy_url,
			autostart, history, s3_enabled, media_delivery, hmac_key
		)
		VALUES
			('restore-me', 'Restore Me', 'token-1', '', '111@s.whatsapp.net', '', 0, 0, 'Connected', '', 1, 0, 0, 'base64', NULL),
			('skip-no-autostart', 'Skip No Autostart', 'token-2', '', '222@s.whatsapp.net', '', 1, 0, 'Connected', '', 0, 0, 0, 'base64', NULL),
			('skip-no-jid', 'Skip No JID', 'token-3', '', '', '', 1, 0, 'Connected', '', 1, 0, 0, 'base64', NULL)
	`)
	if err != nil {
		t.Fatalf("seed users: %v", err)
	}

	users, err := loadStartupUsers(db)
	if err != nil {
		t.Fatalf("load startup users: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("expected exactly one startup user, got %d", len(users))
	}

	if users[0].ID != "restore-me" {
		t.Fatalf("expected restore-me to be loaded, got %s", users[0].ID)
	}
}
