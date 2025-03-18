package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/sugerio/workflow-service-trial/shared/log"
)

func LoggerMiddleware(c *fiber.Ctx) error {
	logger := log.GetLogger(c.UserContext())
	ctx := context.WithValue(c.UserContext(), log.CTX_KEY_LOGGER, logger)
	c.SetUserContext(ctx)
	return c.Next()
}
