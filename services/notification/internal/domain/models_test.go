package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNotificationUnread(t *testing.T) {
	n := &Notification{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Type:     "info",
		Title:    "Neuer Auftrag",
		IsRead:   false,
	}

	if n.IsRead {
		t.Error("new notification should be unread")
	}
}

func TestNotificationMarkRead(t *testing.T) {
	n := &Notification{
		IsRead: false,
	}

	now := time.Now()
	n.IsRead = true
	n.ReadAt = &now

	if !n.IsRead {
		t.Error("notification should be read after marking")
	}
	if n.ReadAt == nil {
		t.Error("ReadAt should be set")
	}
}

func TestUnreadCount(t *testing.T) {
	uc := &UnreadCount{Count: 42}
	if uc.Count != 42 {
		t.Errorf("Count = %d, want 42", uc.Count)
	}
}

func TestNotificationPreference(t *testing.T) {
	pref := &NotificationPreference{
		Channel: "email",
		Type:    "project_update",
		Enabled: true,
	}
	if !pref.Enabled {
		t.Error("preference should be enabled")
	}
	if pref.Channel != "email" {
		t.Errorf("Channel = %q, want %q", pref.Channel, "email")
	}
}
