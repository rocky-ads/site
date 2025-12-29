package handler

import (
	"fmt"
	"math/rand"
	"time"
)

// GetMessageCount returns the current unread message count for a user
// In production, this would query the database for unread messages
// For now, it uses the same simulation logic as SimulateMessageCountUpdate
func GetMessageCount(userID int) int {
	// Initialize random seed based on userID for consistency
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(userID)))

	// Generate random message count between 0 and 14
	return r.Intn(15)
}

// SimulateMessageCountUpdate simulates a message count update for testing
// In production, this would be called when a user receives a new message
func SimulateMessageCountUpdate(userID int) {
	count := GetMessageCount(userID)

	// Create SSE event
	event := SSEEvent{Event: "message-count"}
	if count > 0 {
		event.Data = fmt.Sprintf("<span>%d</span>", count)
	}

	// Send event to user's browser
	SendSSEEvent(userID, event)
}

// StartMessageCountSimulator starts a background goroutine that simulates
// message count updates for testing purposes
func StartMessageCountSimulator() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Get all connected users and simulate updates
			channelMutex.RLock()
			userIDs := make([]int, 0, len(userChannels))
			for userID := range userChannels {
				userIDs = append(userIDs, userID)
			}
			channelMutex.RUnlock()

			// Send updates to all connected users
			for _, userID := range userIDs {
				SimulateMessageCountUpdate(userID)
			}
		}
	}()
}
