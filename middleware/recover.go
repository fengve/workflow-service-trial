package middleware

import (
	"crypto/md5"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sugerio/workflow-service-trial/shared/log"
)

const (
	maxLogBodySize = 1000
)

// PanicRecover recovers from panics and logs the panic message
func NewPanicRecover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				// Get the stack trace
				stack := debug.Stack()
				// Log the panic message
				handlePanic(c, r, stack)
				// Return a 500 Internal Server Error response
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal Server Error",
				})
			}
		}()
		// Call the next middleware/handler in the chain
		return c.Next()
	}
}

// Format the panic message
func formatMessage(c *fiber.Ctx, r any, stack []byte, stackHash string) string {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("Panic Error: %v\n", r))
	msg.WriteString(fmt.Sprintf("URL: %s\n", c.Path()))
	msg.WriteString(fmt.Sprintf("Method: %s\n", c.Method()))
	body := string(c.BodyRaw())
	if len(body) > maxLogBodySize {
		body = body[:maxLogBodySize] + "..." + fmt.Sprintf(" (truncated %d bytes)", len(body)-maxLogBodySize)
	}
	msg.WriteString(fmt.Sprintf("Body: %s\n", body))
	msg.WriteString(fmt.Sprintf("Stack Hash: %s\n", stackHash))
	msg.WriteString(fmt.Sprintf("Stack Trace:\n%s\n", string(stack)))
	return msg.String()
}

// Log the panic message
func handlePanic(c *fiber.Ctx, r any, stack []byte) {
	stackHash := fmt.Sprintf("%x", md5.Sum(stack))
	msg := formatMessage(c, r, stack, stackHash)
	log.GetLogger(c.UserContext()).Error(msg)
}
