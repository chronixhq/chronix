package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func saveImapConfig() error {
	paths, err := resolvePaths(CLI.DataDir)
	if err != nil {
		return err
	}

	store, err := OpenStore(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	if _, err := store.Results.Exec(`DELETE FROM imap_settings`); err != nil {
		return err
	}
	_, err = store.Results.Exec(
		`INSERT INTO imap_settings (host, port, user, pass, ssl) VALUES (?, ?, ?, ?, ?)`,
		CLI.Config.Imap.Host,
		CLI.Config.Imap.Port,
		CLI.Config.Imap.User,
		CLI.Config.Imap.Pass,
		CLI.Config.Imap.SSL,
	)
	return err
}

func generateToken() error {
	paths, err := resolvePaths(CLI.DataDir)
	if err != nil {
		return err
	}

	token := fmt.Sprintf("token-%d", time.Now().UnixNano())
	payload := fmt.Sprintf("{\"token\":\"%s\"}", token)
	fmt.Println(payload)

	store, err := OpenStore(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	_, err = store.Results.Exec(`INSERT INTO shell_logs (args, output) VALUES (?, ?)`, fmt.Sprint(osArgsTail()), payload)
	return err
}

func runIMAPWorker(ctx context.Context, db *sql.DB, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	poll := func() {
		if err := syncIMAPInbox(db); err != nil {
			slog.Warn("imap poll failed", "error", err)
		}
	}

	poll()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			poll()
		}
	}
}

func syncIMAPInbox(db *sql.DB) error {
	var host, user, pass string
	var port int
	var sslEnabled bool
	if err := db.QueryRow(`SELECT host, port, user, pass, ssl FROM imap_settings LIMIT 1`).Scan(&host, &port, &user, &pass, &sslEnabled); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	var c *client.Client
	var err error
	if sslEnabled {
		c, err = client.DialTLS(addr, nil)
	} else {
		c, err = client.Dial(addr)
	}
	if err != nil {
		return fmt.Errorf("dial imap: %w", err)
	}
	defer c.Logout()

	if err := c.Login(user, pass); err != nil {
		return fmt.Errorf("imap login: %w", err)
	}

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return fmt.Errorf("select inbox: %w", err)
	}
	if mbox.Messages == 0 {
		return nil
	}

	fromSeq := uint32(1)
	if mbox.Messages > 10 {
		fromSeq = mbox.Messages - 9
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(fromSeq, mbox.Messages)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	for msg := range messages {
		if msg == nil || msg.Envelope == nil {
			continue
		}
		subject := msg.Envelope.Subject
		fromAddr := ""
		if len(msg.Envelope.From) > 0 {
			fromAddr = msg.Envelope.From[0].Address()
		}
		date := msg.Envelope.Date

		var exists int
		if err := db.QueryRow(`SELECT COUNT(*) FROM imap_logs WHERE subject = ? AND from_addr = ? AND received_at = ?`, subject, fromAddr, date).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		if _, err := db.Exec(`INSERT INTO imap_logs (subject, from_addr, received_at) VALUES (?, ?, ?)`, subject, fromAddr, date); err != nil {
			return err
		}

		slog.Info("logged imap message", "subject", subject, "from", fromAddr)
	}

	if err := <-done; err != nil {
		return fmt.Errorf("fetch imap messages: %w", err)
	}

	return nil
}

func osArgsTail() []string {
	if len(os.Args) <= 1 {
		return nil
	}
	return os.Args[1:]
}
