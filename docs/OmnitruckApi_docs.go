// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplateOmnitruckApi = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/architectures": {
            "get": {
                "description": "Returns a valid list of valid platform keys along with friendly names.\nAny of these architecture keys can be used in the p query string value in various endpoints below.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get architecture keys",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/install.ps1": {
            "get": {
                "description": "The ` + "`" + `ACCEPT` + "`" + ` HTTP header with a value of ` + "`" + `text/plain` + "`" + ` must be provided in the request for a text response to be returned",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Download install script for Windows",
                "parameters": [
                    {
                        "type": "string",
                        "description": "License ID",
                        "name": "license_id",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/install.sh": {
            "get": {
                "description": "The ` + "`" + `ACCEPT` + "`" + ` HTTP header with a value of ` + "`" + `application/x-sh` + "`" + ` must be provided in the request for a shell script response to be returned",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Download install script for Linux",
                "parameters": [
                    {
                        "type": "string",
                        "description": "License ID",
                        "name": "license_id",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/package-managers": {
            "get": {
                "description": "Get the list of available package managers",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get available package managers",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/platforms": {
            "get": {
                "description": "Returns a valid list of valid platform keys along with full friendly names.\nAny of these platform keys can be used in the p query string value in various endpoints below.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get platform keys",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/omnitruck.PlatformList"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/products": {
            "get": {
                "description": "Returns a valid list of valid product keys.\nAny of these product keys can be used in the \u003cPRODUCT\u003e value of other endpoints. Please note many of these products are used for internal tools only and many have been EOL'd.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get list of available products",
                "parameters": [
                    {
                        "type": "boolean",
                        "description": "EOL Products",
                        "name": "eol",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/relatedProducts": {
            "get": {
                "description": "The ` + "`" + `ACCEPT` + "`" + ` HTTP header with a value of ` + "`" + `application/json` + "`" + ` must be provided in the request for a JSON response to be returned",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get related products from a BOM",
                "parameters": [
                    {
                        "type": "string",
                        "description": "bom",
                        "name": "bom",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "License ID",
                        "name": "license_id",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/{channel}/{product}/download": {
            "get": {
                "description": "Get details for a particular package.\nThe ` + "`" + `ACCEPT` + "`" + ` HTTP header with a value of ` + "`" + `application/json` + "`" + ` must be provided in the request for a JSON response to be returned",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Download a product package",
                "parameters": [
                    {
                        "enum": [
                            "current",
                            "stable"
                        ],
                        "type": "string",
                        "description": "Channel",
                        "name": "channel",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "chef",
                        "description": "Product",
                        "name": "product",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "ubuntu",
                        "description": "Platform, valid values are returned from the ` + "`" + `/platforms` + "`" + ` endpoint.",
                        "name": "p",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "20.04",
                        "description": "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15.",
                        "name": "pv",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "x86_64",
                        "description": "Machine architecture, valid values are returned by the ` + "`" + `/architectures` + "`" + ` endpoint.",
                        "name": "m",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "tar",
                        "description": "Package Manager, valid values depend on the platform (e.g., Linux: deb, tar; Windows: msi).",
                        "name": "pm",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "default": "latest",
                        "description": "Version of the product to be installed. A version always takes the form ` + "`" + `x.y.z` + "`" + `",
                        "name": "v",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "License ID",
                        "name": "license_id",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "default": false,
                        "description": "EOL Products",
                        "name": "eol",
                        "in": "query"
                    }
                ],
                "responses": {
                    "302": {
                        "description": "Found"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/{channel}/{product}/fileName": {
            "get": {
                "description": "The ` + "`" + `ACCEPT` + "`" + ` HTTP header with a value of ` + "`" + `application/json` + "`" + ` must be provided in the request for a JSON response to be returned",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get filename for a product package",
                "parameters": [
                    {
                        "enum": [
                            "current",
                            "stable"
                        ],
                        "type": "string",
                        "description": "Channel",
                        "name": "channel",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "chef",
                        "description": "Product",
                        "name": "product",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "ubuntu",
                        "description": "Platform, valid values are returned from the ` + "`" + `/platforms` + "`" + ` endpoint.",
                        "name": "p",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "20.04",
                        "description": "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15.",
                        "name": "pv",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "x86_64",
                        "description": "Machine architecture, valid values are returned by the ` + "`" + `/architectures` + "`" + ` endpoint.",
                        "name": "m",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "tar",
                        "description": "Package Manager, valid values depend on the platform (e.g., Linux: deb, tar; Windows: msi).",
                        "name": "pm",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "default": "latest",
                        "description": "Version of the product to be installed. A version always takes the form ` + "`" + `x.y.z` + "`" + `",
                        "name": "v",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "License ID",
                        "name": "license_id",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/{channel}/{product}/metadata": {
            "get": {
                "description": "Get details for a particular package.\nThe ` + "`" + `ACCEPT` + "`" + ` HTTP header with a value of ` + "`" + `application/json` + "`" + ` must be provided in the request for a JSON response to be returned",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get metadata for a product",
                "parameters": [
                    {
                        "enum": [
                            "current",
                            "stable"
                        ],
                        "type": "string",
                        "description": "Channel",
                        "name": "channel",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "chef",
                        "description": "Product",
                        "name": "product",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "ubuntu",
                        "description": "Platform, valid values are returned from the ` + "`" + `/platforms` + "`" + ` endpoint.",
                        "name": "p",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "20.04",
                        "description": "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15.",
                        "name": "pv",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "x86_64",
                        "description": "Machine architecture, valid values are returned by the ` + "`" + `/architectures` + "`" + ` endpoint.",
                        "name": "m",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "tar",
                        "description": "Package Manager, valid values depend on the platform (e.g., Linux: deb, tar; Windows: msi).",
                        "name": "pm",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "default": "latest",
                        "description": "Version of the product to be installed. A version always takes the form ` + "`" + `x.y.z` + "`" + `",
                        "name": "v",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "License ID",
                        "name": "license_id",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "default": false,
                        "description": "EOL Products",
                        "name": "eol",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/omnitruck.PackageMetadata"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/{channel}/{product}/packages": {
            "get": {
                "description": "Get the full list of all packages for a particular channel and product combination.\nBy default all packages for the latest version are returned. If the v query string parameter is included the packages for the specified version are returned.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get packages for a product version",
                "parameters": [
                    {
                        "enum": [
                            "current",
                            "stable"
                        ],
                        "type": "string",
                        "description": "Channel",
                        "name": "channel",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "chef",
                        "description": "Product",
                        "name": "product",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Version",
                        "name": "v",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "License ID",
                        "name": "license_id",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "default": false,
                        "description": "EOL Products",
                        "name": "eol",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/omnitruck.PackageList"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/{channel}/{product}/versions/all": {
            "get": {
                "description": "Get a list of all available version numbers for a particular channel and product combination",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get all versions of a product in a channel",
                "parameters": [
                    {
                        "enum": [
                            "current",
                            "stable"
                        ],
                        "type": "string",
                        "description": "Channel",
                        "name": "channel",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Product",
                        "name": "product",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "License ID",
                        "name": "license_id",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "default": false,
                        "description": "EOL Products",
                        "name": "eol",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/{channel}/{product}/versions/latest": {
            "get": {
                "description": "Get the latest version number for a particular channel and product combination.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get latest version of a product in a channel",
                "parameters": [
                    {
                        "enum": [
                            "current",
                            "stable"
                        ],
                        "type": "string",
                        "description": "Channel",
                        "name": "channel",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Product",
                        "name": "product",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "License ID",
                        "name": "license_id",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "ErrorResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                },
                "status_text": {
                    "type": "string"
                }
            }
        },
        "omnitruck.ArchList": {
            "type": "object",
            "additionalProperties": {
                "$ref": "#/definitions/omnitruck.PackageMetadata"
            }
        },
        "omnitruck.PackageList": {
            "type": "object",
            "additionalProperties": {
                "$ref": "#/definitions/omnitruck.PlatformVersionList"
            }
        },
        "omnitruck.PackageMetadata": {
            "type": "object",
            "properties": {
                "sha1": {
                    "type": "string"
                },
                "sha256": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                },
                "version": {
                    "type": "string"
                }
            }
        },
        "omnitruck.PlatformList": {
            "type": "object",
            "additionalProperties": {
                "type": "string"
            }
        },
        "omnitruck.PlatformVersionList": {
            "type": "object",
            "additionalProperties": {
                "$ref": "#/definitions/omnitruck.ArchList"
            }
        }
    }
}`

// SwaggerInfoOmnitruckApi holds exported Swagger Info so clients can modify it
var SwaggerInfoOmnitruckApi = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Licensed Omnitruck API",
	Description:      "Licensed Omnitruck API",
	InfoInstanceName: "OmnitruckApi",
	SwaggerTemplate:  docTemplateOmnitruckApi,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfoOmnitruckApi.InstanceName(), SwaggerInfoOmnitruckApi)
}
