package api

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
	sharedLog "github.com/sugerio/workflow-service-trial/shared/log"
)

// Format the error message. The original error message may contain sensitive information.
// 1. Filter out the pq DB error message.
func FormatErrorMessage(err error) string {
	msg := err.Error()
	if strings.HasPrefix(msg, "pq: ") {
		return "DB error occurred"
	}
	return msg
}

func HandleInternalServerErrorWithTrace(c *fiber.Ctx, err error) error {
	// Get the RunTime code file, line & function.
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	logger := sharedLog.GetLogger(c.UserContext())
	logger.Error(fmt.Sprintf("Failed to %s", frame.Function),
		"location", fmt.Sprintf("%s:%d", frame.File, frame.Line),
		"error", err,
		"request", c.Request())
	return c.Status(fiber.StatusInternalServerError).SendString(FormatErrorMessage(err))
}

func HandleBadRequestErrorWithTrace(c *fiber.Ctx, err error) error {
	// Get the RunTime code file, line & function.
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	logger := sharedLog.GetLogger(c.UserContext())
	logger.Error(fmt.Sprintf("Failed to %s", frame.Function),
		"location", fmt.Sprintf("%s:%d", frame.File, frame.Line),
		"error", err,
		"request", c.Request())
	return c.Status(fiber.StatusBadRequest).SendString(FormatErrorMessage(err))
}

func HandleUnauthorizedErrorWithTrace(c *fiber.Ctx, err error) error {
	// Get the RunTime code file, line & function.
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	logger := sharedLog.GetLogger(c.UserContext())
	logger.Error(fmt.Sprintf("Failed to %s", frame.Function),
		"location", fmt.Sprintf("%s:%d", frame.File, frame.Line),
		"error", err,
		"request", c.Request())
	return c.Status(fiber.StatusUnauthorized).SendString(FormatErrorMessage(err))
}

func HandleNotFoundErrorWithTrace(c *fiber.Ctx, err error) error {
	// Get the RunTime code file, line & function.
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	logger := sharedLog.GetLogger(c.UserContext())
	logger.Error(fmt.Sprintf("Failed to %s", frame.Function),
		"location", fmt.Sprintf("%s:%d", frame.File, frame.Line),
		"error", err,
		"request", c.Request())
	return c.Status(fiber.StatusNotFound).SendString(FormatErrorMessage(err))
}

// HandleConflictErrorWithTrace for duplicate or conflict error
func HandleConflictErrorWithTrace(c *fiber.Ctx, err error) error {
	// Get the RunTime code file, line & function.
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	logger := sharedLog.GetLogger(c.UserContext())
	logger.Error(fmt.Sprintf("Failed to %s", frame.Function),
		"location", fmt.Sprintf("%s:%d", frame.File, frame.Line),
		"error", err,
		"request", c.Request())
	return c.Status(fiber.StatusConflict).SendString(FormatErrorMessage(err))
}

func LogWithTrace(err error) {
	// Get the RunTime code file, line & function.
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	logger := sharedLog.GetLogger(context.Background())
	logger.Error(
		"location", fmt.Sprintf("%s:%d", frame.File, frame.Line),
		"msg", fmt.Sprintf("Failed to %s", frame.Function),
		"error", err)
}
