package oauth2

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type webSessionStorage struct {
	db          *sql.DB
	stopCleanup chan bool
}

func newWebSessionStorage(db *sql.DB) *webSessionStorage {
	p := &webSessionStorage{db: db}

	go p.startCleanup(5 * time.Minute)

	return p
}

// Find returns the data for a given session token from the WebSessionStorage instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (p *webSessionStorage) Find(token string) (b []byte, exists bool, err error) {
	row := p.db.QueryRow("SELECT data FROM web_sessions WHERE token = $1 AND DATETIME('now') < expiry", token)
	err = row.Scan(&b)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Commit adds a session token and data to the WebSessionStorage instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (p *webSessionStorage) Commit(token string, b []byte, expiry time.Time) error {
	_, err := p.db.Exec("REPLACE INTO web_sessions (token, data, expiry) VALUES ($1, $2, $3)", token, b, expiry)
	if err != nil {
		return fmt.Errorf("commit error: %w", err)
	}
	return nil
}

// Delete removes a session token and corresponding data from the WebSessionStorage
// instance.
func (p *webSessionStorage) Delete(token string) error {
	_, err := p.db.Exec("DELETE FROM web_sessions WHERE token = $1", token)
	return err
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the WebSessionStorage instance.
func (p *webSessionStorage) All() (map[string][]byte, error) {
	rows, err := p.db.Query("SELECT token, data FROM web_sessions WHERE DATETIME('now') < expiry")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make(map[string][]byte)

	for rows.Next() {
		var (
			token string
			data  []byte
		)

		err = rows.Scan(&token, &data)
		if err != nil {
			return nil, err
		}

		sessions[token] = data
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (p *webSessionStorage) startCleanup(interval time.Duration) {
	p.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := p.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-p.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the WebSessionStorage
// instance. It's rare to terminate this; generally WebSessionStorage instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the WebSessionStorage is transient.
// An example is creating a new WebSessionStorage instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// WebSessionStorage object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (p *webSessionStorage) StopCleanup() {
	if p.stopCleanup != nil {
		p.stopCleanup <- true
	}
}

func (p *webSessionStorage) deleteExpired() error {
	_, err := p.db.Exec("DELETE FROM web_sessions WHERE expiry < DATETIME('now')")
	return err
}
