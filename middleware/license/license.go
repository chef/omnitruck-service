package license

import (
	"github.com/chef/omnitruck-service/clients"
	"github.com/gofiber/fiber/v2"
	"regexp"
)


var re = regexp.MustCompile(`platforms|architectures|products|swagger`)


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
}

var ConfigDefault = Config{
	Required:      true,
	Next:          nil,
	LicenseClient: nil,
	Unauthorized:  nil,
}

func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	cfg := ConfigDefault

	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Unauthorized == nil {
		cfg.Unauthorized = func(code int, msg string, c *fiber.Ctx) error {
			return c.Status(code).JSON(msg)
		}
	}

	if cfg.LicenseClient == nil {
		cfg.LicenseClient = clients.NewLicenseClient()
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
		swaggerPath := c.Path()
		if !re.MatchString(swaggerPath) {

			if len(id) == 0 {
				if cfg.Required {
					return cfg.Unauthorized(403, "Missing license_id query param", c)
				}
				// No license id found but not required
				return c.Next()
			}

			resp := clients.Response{}
			request := cfg.LicenseClient.Validate(id, &resp)

			// Invalid license of some sort returned from license API
			if request.Code >= 400 {
				return cfg.Unauthorized(403, resp.Message, c)
			}
		}
		c.Locals("valid_license", true)

		return c.Next()
	}
}
