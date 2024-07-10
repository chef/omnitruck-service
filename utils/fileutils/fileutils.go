package fileutils

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/models"
	"github.com/gofiber/fiber/v2"
)

type FileUtilsImpl struct{}

type FileUtils interface {
	GetScript(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error)
}

func NewFileUtils() *FileUtilsImpl {
	return &FileUtilsImpl{}
}
func (fu *FileUtilsImpl) GetScript(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
	scriptInput := models.ScriptParams{
		BaseUrl:   baseUrl,
		LicenseId: params.LicenseId,
		Product:   params.Product,
	}
	templateReader, err := template.ParseFiles(filePath)
	if err != nil {
		return "", fiber.NewError(http.StatusInternalServerError, "error while parsing the template files: "+err.Error())
	}
	var scriptResponse bytes.Buffer
	err = templateReader.Execute(&scriptResponse, scriptInput)
	if err != nil {
		return "", fiber.NewError(http.StatusInternalServerError, "error while executing the template reader object: "+err.Error())
	}
	return scriptResponse.String(), nil
}
