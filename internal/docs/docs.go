package docs

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/docs"
)

func scalarHTML(specURL string) string {
	// Scalar web component (CDN) loading swagger/openapi spec from specURL
	// Keep this small and dependency-free.
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>API Docs</title>
    <style>
      html, body { height: 100%%; margin: 0; }
    </style>
  </head>
  <body>
    <script id="api-reference" data-url="%s"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`, specURL)
}

// Register registers /docs UI and /docs/swagger.json spec endpoints.
func Register(app *fiber.App) {
	app.Get("/docs", func(c *fiber.Ctx) error {
		c.Type("html", "utf-8")
		return c.SendString(scalarHTML("/docs/swagger.json"))
	})
	app.Get("/docs/", func(c *fiber.Ctx) error {
		c.Type("html", "utf-8")
		return c.SendString(scalarHTML("/docs/swagger.json"))
	})

	app.Get("/docs/swagger.json", func(c *fiber.Ctx) error {
		c.Type("json", "utf-8")
		return c.SendString(docs.SwaggerInfo.ReadDoc())
	})
}
