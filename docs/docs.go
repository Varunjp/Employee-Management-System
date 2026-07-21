// Package docs registers the OpenAPI (Swagger 2.0) specification consumed
// by echo-swagger's /swagger/* handler. It is structured the same way the
// swaggo/swag CLI (`swag init`) generates this file, so if handlers are
// changed later, running `swag init` will regenerate it in place using the
// @-annotations found in the handler and router files.
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "swagger": "2.0",
    "info": {
        "title": "Employee Management API",
        "description": "A clean-architecture Employee Management System REST API built with Golang, Echo, PostgreSQL (raw pgx queries), Redis caching, and JWT authentication.",
        "termsOfService": "http://swagger.io/terms/",
        "contact": { "name": "API Support" },
        "license": { "name": "MIT" },
        "version": "1.0"
    },
    "host": "localhost:8080",
    "basePath": "/api/v1",
    "paths": {
        "/auth/login": {
            "post": {
                "tags": ["auth"],
                "summary": "Obtain a JWT access token",
                "description": "Authenticates with a username/password and returns a bearer token required for the create, update, and delete employee endpoints.",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "parameters": [
                    {
                        "in": "body",
                        "name": "request",
                        "required": true,
                        "schema": { "$ref": "#/definitions/LoginRequest" }
                    }
                ],
                "responses": {
                    "200": { "description": "OK", "schema": { "$ref": "#/definitions/LoginResponse" } },
                    "400": { "description": "Bad Request", "schema": { "$ref": "#/definitions/ErrorResponse" } },
                    "401": { "description": "Unauthorized", "schema": { "$ref": "#/definitions/ErrorResponse" } }
                }
            }
        },
        "/employees": {
            "get": {
                "tags": ["employees"],
                "summary": "List all employees",
                "description": "Retrieves a list of all employees. Results are served from a Redis cache when available.",
                "produces": ["application/json"],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": { "type": "array", "items": { "$ref": "#/definitions/EmployeeResponse" } }
                    },
                    "500": { "description": "Internal Server Error", "schema": { "$ref": "#/definitions/ErrorResponse" } }
                }
            },
            "post": {
                "security": [{ "BearerAuth": [] }],
                "tags": ["employees"],
                "summary": "Create an employee",
                "description": "Creates a new employee record.",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "parameters": [
                    {
                        "in": "body",
                        "name": "request",
                        "required": true,
                        "schema": { "$ref": "#/definitions/CreateEmployeeRequest" }
                    }
                ],
                "responses": {
                    "201": { "description": "Created", "schema": { "$ref": "#/definitions/EmployeeResponse" } },
                    "400": { "description": "Bad Request", "schema": { "$ref": "#/definitions/ErrorResponse" } },
                    "401": { "description": "Unauthorized", "schema": { "$ref": "#/definitions/ErrorResponse" } },
                    "500": { "description": "Internal Server Error", "schema": { "$ref": "#/definitions/ErrorResponse" } }
                }
            }
        },
        "/employees/{id}": {
            "get": {
                "tags": ["employees"],
                "summary": "Get an employee by ID",
                "description": "Retrieves details of a specific employee. Results are served from a Redis cache when available.",
                "produces": ["application/json"],
                "parameters": [
                    { "type": "integer", "description": "Employee ID", "name": "id", "in": "path", "required": true }
                ],
                "responses": {
                    "200": { "description": "OK", "schema": { "$ref": "#/definitions/EmployeeResponse" } },
                    "400": { "description": "Bad Request", "schema": { "$ref": "#/definitions/ErrorResponse" } },
                    "404": { "description": "Not Found", "schema": { "$ref": "#/definitions/ErrorResponse" } }
                }
            },
            "put": {
                "security": [{ "BearerAuth": [] }],
                "tags": ["employees"],
                "summary": "Update an employee",
                "description": "Updates details of a specific employee.",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "parameters": [
                    { "type": "integer", "description": "Employee ID", "name": "id", "in": "path", "required": true },
                    {
                        "in": "body",
                        "name": "request",
                        "required": true,
                        "schema": { "$ref": "#/definitions/UpdateEmployeeRequest" }
                    }
                ],
                "responses": {
                    "200": { "description": "OK", "schema": { "$ref": "#/definitions/EmployeeResponse" } },
                    "400": { "description": "Bad Request", "schema": { "$ref": "#/definitions/ErrorResponse" } },
                    "401": { "description": "Unauthorized", "schema": { "$ref": "#/definitions/ErrorResponse" } },
                    "404": { "description": "Not Found", "schema": { "$ref": "#/definitions/ErrorResponse" } }
                }
            },
            "delete": {
                "security": [{ "BearerAuth": [] }],
                "tags": ["employees"],
                "summary": "Delete an employee",
                "description": "Deletes a specific employee by ID.",
                "parameters": [
                    { "type": "integer", "description": "Employee ID", "name": "id", "in": "path", "required": true }
                ],
                "responses": {
                    "204": { "description": "No Content" },
                    "400": { "description": "Bad Request", "schema": { "$ref": "#/definitions/ErrorResponse" } },
                    "401": { "description": "Unauthorized", "schema": { "$ref": "#/definitions/ErrorResponse" } },
                    "404": { "description": "Not Found", "schema": { "$ref": "#/definitions/ErrorResponse" } }
                }
            }
        },
        "/health": {
            "get": {
                "tags": ["health"],
                "summary": "Health check",
                "description": "Returns 200 OK if the service is up and running.",
                "produces": ["application/json"],
                "responses": {
                    "200": { "description": "OK", "schema": { "type": "object", "additionalProperties": { "type": "string" } } }
                }
            }
        }
    },
    "definitions": {
        "EmployeeResponse": {
            "type": "object",
            "properties": {
                "id": { "type": "integer", "example": 1 },
                "name": { "type": "string", "example": "John Doe" },
                "position": { "type": "string", "example": "Software Engineer" },
                "salary": { "type": "integer", "example": 60000 },
                "hired_date": { "type": "string", "example": "2024-06-01" },
                "created_at": { "type": "string", "example": "2024-06-10T12:00:00Z" },
                "updated_at": { "type": "string", "example": "2024-06-10T12:00:00Z" }
            }
        },
        "CreateEmployeeRequest": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "example": "John Doe" },
                "position": { "type": "string", "example": "Software Engineer" },
                "salary": { "type": "integer", "example": 60000 },
                "hired_date": { "type": "string", "example": "2024-06-01" }
            }
        },
        "UpdateEmployeeRequest": {
            "type": "object",
            "properties": {
                "name": { "type": "string", "example": "John Doe" },
                "position": { "type": "string", "example": "Senior Software Engineer" },
                "salary": { "type": "integer", "example": 80000 },
                "hired_date": { "type": "string", "example": "2024-06-01" }
            }
        },
        "LoginRequest": {
            "type": "object",
            "properties": {
                "username": { "type": "string", "example": "admin" },
                "password": { "type": "string", "example": "admin123" }
            }
        },
        "LoginResponse": {
            "type": "object",
            "properties": {
                "access_token": { "type": "string" },
                "token_type": { "type": "string", "example": "Bearer" },
                "expires_in": { "type": "integer", "example": 3600 }
            }
        },
        "ErrorResponse": {
            "type": "object",
            "properties": {
                "message": { "type": "string", "example": "employee not found" }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "Type \"Bearer\" followed by a space and the JWT token, e.g. \"Bearer eyJhbGciOi...\"."
        }
    }
}`

// SwaggerInfo holds exported Swagger metadata, mirroring what `swag init`
// produces so that other tooling relying on this variable keeps working.
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Employee Management API",
	Description:      "A clean-architecture Employee Management System REST API built with Golang, Echo, PostgreSQL (raw pgx queries), Redis caching, and JWT authentication.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
