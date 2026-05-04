package middleware

import (
	"fmt"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/types/response"
)

const ContextKeyValidatedBody = "validated_body"
const ContextKeyValidatedQuery = "validated_query"

var (
	validateOnce sync.Once
	validateInst *validator.Validate
)

func getValidator() *validator.Validate {
	validateOnce.Do(func() {
		validateInst = validator.New()
	})
	return validateInst
}

// ValidateBody parses JSON body into a struct created by factory(), validates it using struct tags,
// then stores it in context locals under ContextKeyValidatedBody.
//
// On failure it returns 400 with the same shape used elsewhere: {success:false,error:"Invalid request body"}
func ValidateBody(factory func() interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		payload := factory()
		if err := c.BodyParser(payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
		if err := getValidator().Struct(payload); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok && len(validationErrors) > 0 {
				field := validationErrors[0].Field()
				tag := validationErrors[0].Tag()
				message := fmt.Sprintf("Field %s tidak valid", field)
				switch tag {
				case "required":
					message = fmt.Sprintf("Field %s wajib diisi", field)
				case "email":
					message = fmt.Sprintf("Field %s harus berupa email yang valid", field)
				case "oneof":
					message = fmt.Sprintf("Field %s memiliki nilai yang tidak valid", field)
				}
				return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse(message))
			}

			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
		c.Locals(ContextKeyValidatedBody, payload)
		return c.Next()
	}
}

// ValidateQuery parses query params into a struct created by factory(), validates it using struct tags,
// then stores it in context locals under ContextKeyValidatedQuery.
func ValidateQuery(factory func() interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		payload := factory()
		if err := c.QueryParser(payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request"))
		}
		if err := getValidator().Struct(payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request"))
		}
		c.Locals(ContextKeyValidatedQuery, payload)
		return c.Next()
	}
}
