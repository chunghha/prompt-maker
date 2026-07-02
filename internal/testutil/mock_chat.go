// Package testutil provides shared test doubles used across multiple packages.
package testutil

import (
	"context"
	"errors"

	"google.golang.org/genai"
)

// ErrSendMessageNotImplemented is returned when SendMessageFunc is nil.
var ErrSendMessageNotImplemented = errors.New("SendMessage not implemented")

// MockChatSession is a configurable test double for gemini.ChatSession.
type MockChatSession struct {
	SendMessageFunc func(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// SendMessage delegates to SendMessageFunc or returns ErrSendMessageNotImplemented.
func (m *MockChatSession) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	if m.SendMessageFunc != nil {
		return m.SendMessageFunc(ctx, parts...)
	}

	return nil, ErrSendMessageNotImplemented
}
