package license

import (
	"github.com/chef/omnitruck-service/clients"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

type InvalidLicense struct {
	Code int
	Msg  string
}

func (e *InvalidLicense) Error() string {
	return e.Msg
}

type Config struct {
	Required      bool
	Next          func(c *fiber.Ctx) bool
	LicenseClient *clients.License
	Unauthorized  func(code int, msg string, c *fiber.Ctx) error
	Log           *log.Entry
}

var ConfigDefault = Config{
	Required:      true,
	Next:          nil,
	LicenseClient: nil,
	Unauthorized:  nil,
	Log:           log.WithField("pkg", "middleware/license"),
}

func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	cfg := ConfigDefault

	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Log == nil {
		cfg.Log = ConfigDefault.Log
	}

	if cfg.Unauthorized == nil {
		cfg.Unauthorized = func(code int, msg string, c *fiber.Ctx) error {
			return c.Status(code).JSON(msg)
		}
	}

	if cfg.LicenseClient == nil {
		cfg.LicenseClient = clients.NewLicenseClient(cfg.Log)
	}

	return cfg
}

func New(config ...Config) fiber.Handler {
	cfg := configDefault(config...)

	return func(c *fiber.Ctx) (err error) {
		id := c.Query("license_id")
		c.Locals("valid_license", false)
		c.Locals("license_id", id)

		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		if len(id) == 0 {
			if cfg.Required {
				return cfg.Unauthorized(403, "Missing license_id query param", c)
			} else {
				// No license id found but not required
				return c.Next()
			}
		}

		log.Info("Validating license")

		resp := clients.Response{}
		request := cfg.LicenseClient.Validate(id, &resp)

		// Invalid license of some sort returned from license API
		if request.Code >= 400 {
			return cfg.Unauthorized(403, resp.Message, c)
		}

		c.Locals("valid_license", true)

		return c.Next()
	}
}
