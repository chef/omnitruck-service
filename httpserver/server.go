package httpserver

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/dboperations"
	logrus "github.com/chef/omnitruck-service/logger"
	dbconnection "github.com/chef/omnitruck-service/middleware/db"
	"github.com/chef/omnitruck-service/middleware/license"
	"github.com/chef/omnitruck-service/utils/awsutils"
	"github.com/chef/omnitruck-service/utils/template"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Name          string
	Listen        string
	Log           *log.Entry
	Mode          constants.ApiType
	ServiceConfig config.ServiceConfig
}

type Service interface {
	Name() string
	Start(*sync.WaitGroup) error
	Stop() error
}

type ApiServer struct {
	sync.Mutex
	Config           Config
	Log              *log.Entry
	App              *fiber.App
	Validator        omnitruck.RequestValidator
	Mode             constants.ApiType
	DatabaseService  dboperations.IDbOperations
	TemplateRenderer template.TemplateRenderer
	Replicated       replicated.IReplicated
	LicenseClient    clients.ILicense
	locals           map[string]interface{}
}

func New(c Config) *ApiServer {
	service := ApiServer{}
	service.Initialize(c)

	return &service
}

func (server *ApiServer) Initialize(c Config) *ApiServer {
	server.Log = c.Log
	server.Config = c
	server.Validator = omnitruck.NewValidator()
	server.Mode = c.Mode
	server.DatabaseService = dboperations.NewDbOperationsService(dbconnection.NewDbConnectionService(awsutils.NewAwsUtils(), c.ServiceConfig), c.ServiceConfig)
	server.TemplateRenderer = template.NewTemplateRenderer()

	engine := html.New("./views", ".html")
	server.Replicated = replicated.NewReplicatedImpl(c.ServiceConfig.ReplicatedConfig, logrus.NewLogrusStandardLogger())
	server.LicenseClient = clients.NewLicenseClient()

	server.App = fiber.New(fiber.Config{
		DisableStartupMessage: false,
		EnablePrintRoutes:     false,
		ReadTimeout:           time.Duration(c.ServiceConfig.ReadWriteTimeout) * time.Second,
		WriteTimeout:          time.Duration(c.ServiceConfig.ReadWriteTimeout) * time.Second,
		Views:                 engine,
	})

	if c.Mode == constants.Trial || c.Mode == constants.Opensource {
		channel := omnitruck.ContainsValidator{
			Field:      "Channel",
			Values:     []string{"stable"},
			Code:       400,
			AllowEmpty: true,
		}
		server.Validator.Add(&channel)
	}

	// Commented for now
	// if c.Mode == Trial {
	// 	version := omnitruck.ContainsValidator{
	// 		Field:      "Version",
	// 		Values:     []string{"latest"},
	// 		Code:       400,
	// 		AllowEmpty: true,
	// 		Skip: func(c omnitruck.Context) bool {
	// 			return c.License
	// 		},
	// 	}
	// 	server.Validator.Add(&version)
	// }

	if c.Mode == constants.Trial || c.Mode == constants.Commercial {
		server.Log.Info("Adding EOL Validator")
		eolversion := omnitruck.EolVersionValidator{}
		server.Validator.Add(&eolversion)
	}

	return server
}

func (server *ApiServer) Start(wg *sync.WaitGroup) error {
	wg.Add(1)
	go server.StartService()

	return nil
}

func (server *ApiServer) StartService() {
	// Setup io writer for the logger
	// Needs to be in the method where we start the service
	// So the io writer will be closed when the service ends
	lw := server.Log.Writer()
	defer lw.Close()
	server.App.Use(logger.New(logger.Config{
		Format: "LicenseId :- ${locals:license_id} : Method :- ${method} : IP :- ${ip} : EndPoint :- ${path} : channel :- ${channel} : product :- ${product} : platform :- ${platform} : platform version :- ${platformVersion} : architecture :- ${architecture} : version :- ${version} : status :- ${status} : latency :- ${latency} : Time :- [${time}] : request-id :- ${locals:requestid} \n",
		Output: lw,
		CustomTags: map[string]logger.LogFunc{
			"channel": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(fmt.Sprint(c.Params("channel")))
			},
			"product": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(fmt.Sprint(c.Params("product")))
			},
			"version": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(fmt.Sprint(c.Query("v")))
			},
			"platform": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(fmt.Sprint(c.Query("p")))
			},
			"platformVersion": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(fmt.Sprint(c.Query("pv")))
			},
			"architecture": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(fmt.Sprint(c.Query("m")))
			},
		},
	}))

	server.App.Use(cors.New())
	// This will catch panics in the app and prevent it from crashing the server
	// TODO: Figure out if we can better handle logging these, currently it just returns a panic message to the user
	server.App.Use(recover.New())

	server.App.Use(license.New(license.Config{
		URL:      server.Config.ServiceConfig.LicenseServiceUrl,
		Required: true,
		Mode:     server.Mode,
		Next: func(c *fiber.Ctx) bool {
			switch c.Path() {
			case "/status":
				return true
			case "/":
				return true
			case "/swagger":
				return true
			}

			return false
		},
	}))

	// Make sure we build the router last so the middleware has a chance to execute before hand
	server.buildRouter()

	server.Log.Infof("Starting %s server at: %s", server.Config.Name, server.Config.Listen)

	err := server.App.Listen(server.Config.Listen)
	if err != nil {
		if err == http.ErrServerClosed {
			server.Log.WithError(err).Error("Unable to start service")
		} else {
			server.Log.WithError(err).Fatal("Service stopped")
		}
	}
}

func (server *ApiServer) HealthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"name": server.Config.Name,
		"data": "Server is up and running",
	}

	return c.JSON(res)
}
