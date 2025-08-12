package middleware

import "github.com/gofiber/fiber/v2"

func GetRequestID(c *fiber.Ctx) string {
	reqID, ok := c.Locals("requestid").(string)
	if ok {
		return reqID
	}
	return ""
}
