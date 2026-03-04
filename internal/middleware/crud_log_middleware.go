package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/logger"
)

type CRUDLogMiddleware struct {
	logger *logger.Logger
}

func NewCRUDLogMiddleware(logger *logger.Logger) *CRUDLogMiddleware {
	return &CRUDLogMiddleware{logger: logger}
}

func (m *CRUDLogMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := strings.ToUpper(c.Method())
		action := actionFromMethod(method)
		if action == "" {
			return c.Next()
		}

		table := tableFromPath(c.Path())
		if table == "" || table == "logs" || table == "health" || table == "auth" {
			return c.Next()
		}

		err := c.Next()

		userID := ""
		companyID := ""
		if user := GetUserFromContext(c); user != nil {
			userID = user.UserID
			companyID = user.CompanyID
		}

		recordID := c.Params("id")
		if recordID == "" {
			recordID = c.Params("user_id")
		}

		var payload interface{}
		if action == "CREATE" {
			payload = c.Body()
		}

		statusCode := c.Response().StatusCode()
		if err != nil || statusCode >= fiber.StatusBadRequest {
			if err == nil {
				err = fmt.Errorf("request failed with status %d", statusCode)
			}
			m.logger.LogError(action, table, userID, companyID, recordID, err)
			return err
		}

		if action == "CREATE" {
			m.logger.Log(action, table, userID, companyID, recordID, nil, payload)
		} else {
			m.logger.Log(action, table, userID, companyID, recordID, nil, nil)
		}
		return nil
	}
}

func actionFromMethod(method string) string {
	switch method {
	case fiber.MethodPost:
		return "CREATE"
	case fiber.MethodPut, fiber.MethodPatch:
		return "UPDATE"
	case fiber.MethodDelete:
		return "DELETE"
	default:
		return ""
	}
}

func tableFromPath(path string) string {
	clean := strings.Trim(path, "/")
	if clean == "" {
		return ""
	}
	parts := strings.Split(clean, "/")
	if len(parts) == 0 {
		return ""
	}
	if parts[0] == "api" {
		if len(parts) < 2 {
			return ""
		}
		return parts[1]
	}
	return parts[0]
}
