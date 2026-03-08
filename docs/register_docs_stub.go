package docs

import (
	"github.com/gofiber/fiber/v2"
	"os"
	"path/filepath"
)

// Register API docs routes with Scalar theme.
func Register(app *fiber.App) {
	// Read swagger.json file
	swaggerPath := filepath.Join("docs", "swagger.json")
	swaggerJSON, err := os.ReadFile(swaggerPath)
	if err != nil {
		swaggerJSON = []byte("{}")
	}

	// Serve swagger JSON at multiple paths for compatibility
	app.Get("/docs/swagger.json", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		c.Set("Access-Control-Allow-Origin", "*")
		return c.Send(swaggerJSON)
	})
	app.Get("/docs/doc.json", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		c.Set("Access-Control-Allow-Origin", "*")
		return c.Send(swaggerJSON)
	})

	// Serve Scalar API documentation
	app.Get("/docs", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		c.Set("Access-Control-Allow-Origin", "*")
		return c.SendString(scalarHTML)
	})
	app.Get("/docs/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		c.Set("Access-Control-Allow-Origin", "*")
		return c.SendString(scalarHTML)
	})
	app.Get("/docs/index.html", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		c.Set("Access-Control-Allow-Origin", "*")
		return c.SendString(scalarHTML)
	})
}

const scalarHTML = `<!doctype html>
<html>
  <head>
    <title>POS Retail Backend API - Scalar</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/docs/doc.json"
      data-proxy-url="https://proxy.scalar.com">
    </script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`
