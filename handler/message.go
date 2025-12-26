package handler

import (
	"bufio"
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// MessageCountStreamHandler handles SSE stream for message count updates
func MessageCountStreamHandler(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set(fiber.HeaderConnection, "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// Initialize random seed
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		for {
			// Generate random message count between 0 and 10
			count := r.Intn(15)

			// Send the count as HTML
			// If count is 0, send empty string to hide the element
			// Otherwise, send the counter HTML
			if count == 0 {
				fmt.Fprintf(w, "data: \n\n")
			} else {
				countText := fmt.Sprintf("%d", count)
				fmt.Fprintf(w, "data: <span class=\"bg-green-500 text-white rounded-full h-6 min-w-6 px-1.5 flex items-center justify-center text-xs font-bold\">%s</span>\n\n", countText)
			}

			// Flush detects client disconnect
			if err := w.Flush(); err != nil {
				return
			}

			time.Sleep(1 * time.Second)
		}
	}))

	return nil
}
