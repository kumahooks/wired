package notification

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/core/testutil"
)

func TestNew(t *testing.T) {
	t.Parallel()

	model := New()

	assert.Empty(t, model.notifications)
	assert.Equal(t, newStyle(testutil.DefaultTheme()), model.style)
}

func TestApplyThemeRebuildsStyle(t *testing.T) {
	t.Parallel()

	model := New()

	customTheme := testutil.CustomTheme()

	model.ApplyTheme(customTheme)

	assert.Equal(t, newStyle(customTheme), model.style)
}

func TestPushNotification(t *testing.T) {
	t.Parallel()

	model := New()

	before := time.Now().UTC()
	model.PushNotification("config reloaded")
	pushTime := time.Now().UTC()

	require.Len(t, model.notifications, 1)

	notification := model.notifications[0]
	assert.Equal(t, "config reloaded", notification.message)
	assert.WithinDuration(t, before.Add(Lifetime), notification.expiresAt, pushTime.Sub(before))
}

func TestPushNotificationPreservesFIFOOrder(t *testing.T) {
	t.Parallel()

	model := New()

	model.PushNotification("first")
	model.PushNotification("second")
	model.PushNotification("third")

	require.Len(t, model.notifications, 3)

	messages := make([]string, 0, len(model.notifications))
	for _, notification := range model.notifications {
		messages = append(messages, notification.message)
	}

	assert.Equal(t, []string{"first", "second", "third"}, messages)
}

func TestPruneExpired(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	expired := now.Add(-time.Minute)
	active := now.Add(time.Minute)

	tests := []struct {
		name          string
		notifications []Notification
		wantMessages  []string
	}{
		{
			name:          "empty queue stays empty",
			notifications: []Notification{},
			wantMessages:  []string{},
		},
		{
			name: "none expired keeps the queue",
			notifications: []Notification{
				{message: "a", expiresAt: active},
				{message: "b", expiresAt: active},
			},
			wantMessages: []string{"a", "b"},
		},
		{
			name: "expired prefix is dropped, active tail is kept",
			notifications: []Notification{
				{message: "gone-1", expiresAt: expired},
				{message: "gone-2", expiresAt: expired},
				{message: "kept", expiresAt: active},
			},
			wantMessages: []string{"kept"},
		},
		{
			name: "stops at first non-expired entry",
			notifications: []Notification{
				{message: "kept", expiresAt: active},
				{message: "stale-but-kept", expiresAt: expired},
			},
			wantMessages: []string{"kept", "stale-but-kept"},
		},
		{
			name: "all expired empties the queue",
			notifications: []Notification{
				{message: "gone-1", expiresAt: expired},
				{message: "gone-2", expiresAt: expired},
			},
			wantMessages: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := New()
			model.notifications = test.notifications

			model.PruneExpired()

			messages := make([]string, 0, len(model.notifications))
			for _, notification := range model.notifications {
				messages = append(messages, notification.message)
			}

			assert.Equal(t, test.wantMessages, messages)
		})
	}
}

func TestHasActiveNotifications(t *testing.T) {
	t.Parallel()

	model := New()
	assert.False(t, model.HasActiveNotifications())

	model.PushNotification("hello")
	assert.True(t, model.HasActiveNotifications())
}
