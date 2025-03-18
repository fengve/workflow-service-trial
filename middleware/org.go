package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/sugerio/workflow-service-trial/shared/log"
)

// OrgIdMiddleware detects the orgId parameter in the url path and
// sets it in the user context to propagate in the request path.
func OrgIdMiddleware(c *fiber.Ctx) error {
	orgId := c.Params("orgId")
	ctx := context.WithValue(c.UserContext(), log.CTX_KEY_ORGID, orgId)
	c.SetUserContext(ctx)
	return c.Next()
}
