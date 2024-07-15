package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/dboperations"
	log "github.com/chef/omnitruck-service/logger"
	dbconnection "github.com/chef/omnitruck-service/middleware/db"
	"github.com/chef/omnitruck-service/middleware/license"
	"github.com/chef/omnitruck-service/utils/awsutils"
	"github.com/chef/omnitruck-service/utils/template"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
)

type ApiType int

const (
	Trial ApiType = iota
	Opensource
	Commercial
)

type ErrorResponse struct {
	Code       int    `json:"code"`
	StatusText string `json:"status_text"`
	Message    string `json:"message"`
} //@name ErrorResponse

type Config struct {
	Name          string
	Listen        string
	Log           log.ILogger
	Mode          ApiType
	ServiceConfig config.ServiceConfig
}

type Service interface {
	Name() string
	Start(*sync.WaitGroup) error
	Stop() error
}

type ApiService struct {
	sync.Mutex
	Config           Config
	Log              log.ILogger
	App              *fiber.App
	Validator        omnitruck.RequestValidator
	Mode             ApiType
	DatabaseService  dboperations.IDbOperations
	TemplateRenderer template.TemplateRender
}

func New(c Config) *ApiService {
	service := ApiService{}
	service.Initialize(c)

	return &service
}

func (server *ApiService) Initialize(c Config) *ApiService {
	server.Log = c.Log
	server.Config = c
	server.Validator = omnitruck.NewValidator()
	server.Mode = c.Mode
	server.TemplateRenderer = template.NewTemplateRender()
	server.DatabaseService = dboperations.NewDbOperationsService(dbconnection.NewDbConnectionService(awsutils.NewAwsUtils(c.Log), c.ServiceConfig, c.Log), c.ServiceConfig, server.Log)

	engine := html.New("./views", ".html")

	server.App = fiber.New(fiber.Config{
		DisableStartupMessage: false,
		EnablePrintRoutes:     false,
		ReadTimeout:           300 * time.Second,
		WriteTimeout:          300 * time.Second,
		Views:                 engine,
	})

	if c.Mode == Trial || c.Mode == Opensource {
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

	if c.Mode == Trial || c.Mode == Commercial {
		server.Log.Info("Adding EOL Validator")
		eolversion := omnitruck.EolVersionValidator{}
		server.Validator.Add(&eolversion)
	}

	return server
}

func (server *ApiService) Start(wg *sync.WaitGroup) error {
	wg.Add(1)
	go server.StartService()

	return nil
}

func (server *ApiService) StartService() {
	// Setup io writer for the logger
	// Needs to be in the method where we start the service
	// So the io writer will be closed when the service ends
	lw := server.Log.LogWritter()
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
		Required: server.Config.Mode == Commercial,
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
	serverInfo := fmt.Sprintf("Starting %s server at: %s", server.Config.Name, server.Config.Listen)
	server.Log.Info(serverInfo)

	err := server.App.Listen(server.Config.Listen)
	if err != nil {
		if err == http.ErrServerClosed {
			server.Log.Error("Unable to start service", err)
		} else {
			server.Log.Error("Service stopped", err)
		}
	}
}

func (server *ApiService) Omnitruck(c *fiber.Ctx) *omnitruck.Omnitruck {
	client := omnitruck.New(server.logCtx(c))

	return &client
}

func (server *ApiService) DynamoServices(db dboperations.IDbOperations, c *fiber.Ctx) *omnitruck.DynamoServices {
	service := omnitruck.NewDynamoServices(db, server.logCtx(c))

	return &service
}

func (server *ApiService) logCtx(c *fiber.Ctx) log.ILogger {
	return server.Log.With(map[string]interface{}{
		"license_id": c.Locals("license_id"),
	})
}

func (server *ApiService) validLicense(c *fiber.Ctx) bool {
	v := c.Locals("valid_license")
	return v != nil && v.(bool)
}

func (server *ApiService) ValidateRequest(params *omnitruck.RequestParams, c *fiber.Ctx) (error, bool) {
	debugString := fmt.Sprintf("Validating request %+v", params)
	server.logCtx(c).Debug(debugString)
	context := omnitruck.Context{
		License: server.validLicense(c),
	}

	error := server.Validator.Params(params, context)
	if error != nil {
		msgs, code := server.Validator.ErrorMessages(error)
		server.logCtx(c).Error("error validating request", errors.New("errors while validating request"),map[string]interface{}{
			"errors":msgs,
		})
		return c.Status(code).JSON(ErrorResponse{
			Code:       code,
			StatusText: http.StatusText(code),
			Message:    msgs,
		}), false
	}

	return nil, true
}

func (server *ApiService) JSON(c *fiber.Ctx, data interface{}) error {
	var resultBytes bytes.Buffer
	enc := json.NewEncoder(&resultBytes)
	enc.SetEscapeHTML(false)
	err := enc.Encode(data)
	c.Context().Response.SetBodyRaw(resultBytes.Bytes())
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return err
}

func (server *ApiService) SendResponse(c *fiber.Ctx, data interface{}) error {
	return server.JSON(c, data)
}

func (server *ApiService) SendError(c *fiber.Ctx, request *clients.Request) error {

	return c.Status(request.Code).JSON(ErrorResponse{
		Code:       request.Code,
		StatusText: http.StatusText(request.Code),
		Message:    request.Message,
	})
}

func (server *ApiService) SendErrorResponse(c *fiber.Ctx, code int, msg string) error {
	return c.Status(code).JSON(ErrorResponse{
		Code:       code,
		StatusText: http.StatusText(code),
		Message:    msg,
	})
}

func (server *ApiService) HealthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"name": server.Config.Name,
		"data": "Server is up and running",
	}

	return c.JSON(res)
}

func isLatest(v string) bool {
	return len(v) == 0 || v == "latest"
}
