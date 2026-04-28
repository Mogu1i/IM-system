package main

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(dsn string) (*MySQLStore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &MySQLStore{db: db}, nil
}

func (s *MySQLStore) SaveMessages(ctx context.Context, messages []ChatMessage) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO chat_messages (from_user, to_user, content, created_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, msg := range messages {
		var toUser sql.NullString
		if msg.ToUser != "" {
			toUser = sql.NullString{String: msg.ToUser, Valid: true}
		}
		if _, err := stmt.ExecContext(ctx, msg.FromUser, toUser, msg.Content, msg.CreatedAt); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
