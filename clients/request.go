package clients

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

type Request struct {
	Url     string
	Code    int
	Body    []byte
	Message string
	Ok      bool
}

type RequestDataInterface interface {
}

func (r *Request) Failure(code int, msg string) *Request {
	r.Code = code
	r.Message = msg
	r.Ok = false
	return r
}

func (r *Request) Success() *Request {
	r.Ok = true
	return r
}

func (r *Request) ParseData(data RequestDataInterface) *Request {
	if r.Ok {
		err := json.Unmarshal(r.Body, &data)
		if err != nil {
			return r.Failure(fiber.StatusBadRequest, string(r.Body))
		}
	}

	return r
}

func (r *Request) ParseLicenseResp(data RequestDataInterface) *Request {
	err := json.Unmarshal(r.Body, &data)
	if err != nil {
		return r.Failure(fiber.StatusBadRequest, string(r.Body))
	}

	return r
}
