package license

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/chef/omnitruck-service/clients"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNoLicenseIDWithRequiredTrue(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		URL:      "http://example.com",
		Required: true,
		LicenseClient: &clients.MockLicense{
			ValidateFunc: func(id, url string, resp *clients.Response) *clients.Request {
				return &clients.Request{Ok: false, Code: 403, Message: "invalid license"}
			},
			IsTrialFunc: func(l string) bool {
				return false
			},
			IsFreeFunc: func(l string) bool {
				return false
			},
		},
		Unauthorized: func(code int, msg string, c *fiber.Ctx) error {
			return c.Status(code).SendString(msg)
		},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Missing license_id query param", string(body))
}

func TestInvalidLicenseProvided(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		URL:      "http://example.com",
		Required: true,
		LicenseClient: &clients.MockLicense{
			ValidateFunc: func(id, url string, resp *clients.Response) *clients.Request {
				return &clients.Request{Ok: false, Code: 403, Message: "invalid license"}
			},
			IsTrialFunc: func(l string) bool {
				return false
			},
			IsFreeFunc: func(l string) bool {
				return false
			},
		},
		Unauthorized: func(code int, msg string, c *fiber.Ctx) error {
			return c.Status(code).SendString(msg)
		},
	}))

	req := httptest.NewRequest("GET", "/?license_id=bad-license", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestValidLicenseProvided(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		URL:      "http://example.com",
		Required: true,
		Mode:     2, // Commercial mode
		LicenseClient: &clients.MockLicense{
			ValidateFunc: func(id, url string, resp *clients.Response) *clients.Request {
				resp.Message = "valid license"
				return &clients.Request{Ok: true, Code: 200, Message: "valid license"}
			},
			IsTrialFunc: func(l string) bool {
				return false
			},
			IsFreeFunc: func(l string) bool {
				return false
			},
		},
		Unauthorized: func(code int, msg string, c *fiber.Ctx) error {
			return c.Status(code).SendString(msg)
		},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/?license_id=valid-license", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestNextSkipsLicense(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		URL:      "http://example.com",
		Required: true,
		LicenseClient: &clients.MockLicense{
			ValidateFunc: func(id, url string, resp *clients.Response) *clients.Request {
				resp.Message = "valid license"
				return &clients.Request{Ok: true, Code: 200, Message: "valid license"}
			},
			IsTrialFunc: func(l string) bool {
				return false
			},
			IsFreeFunc: func(l string) bool {
				return false
			},
		},
		Next: func(c *fiber.Ctx) bool {
			return true
		},
		Unauthorized: func(code int, msg string, c *fiber.Ctx) error {
			return c.Status(code).SendString(msg)
		},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestNonRequired_NoLicenseID(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		URL:      "http://example.com",
		Required: false,
		LicenseClient: &clients.MockLicense{
			ValidateFunc: func(id, url string, resp *clients.Response) *clients.Request {
				resp.Message = "valid license"
				return &clients.Request{Ok: true, Code: 200, Message: "valid license"}
			},
			IsTrialFunc: func(l string) bool {
				return false
			},
			IsFreeFunc: func(l string) bool {
				return false
			},
		},
		Unauthorized: func(code int, msg string, c *fiber.Ctx) error {
			return c.Status(code).SendString(msg)
		},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
