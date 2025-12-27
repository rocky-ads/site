package handler

import (
	"fmt"
	"math/rand"
	"time"
)

// SimulateMessageCountUpdate simulates a message count update for testing
// In production, this would be called when a user receives a new message
func SimulateMessageCountUpdate(userID int) {
	// Initialize random seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate random message count between 0 and 14
	count := r.Intn(15)

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
