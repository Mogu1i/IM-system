package models

import (
	"context"
	"time"
)

type ChatMessage struct {
	FromUser  string
	ToUser    string
	Content   string
	CreatedAt time.Time
}

type MessageStore interface {
	SaveMessages(ctx context.Context, messages []ChatMessage) error
}
