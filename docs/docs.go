// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/segments": {
            "post": {
                "tags": [
                    "segments"
                ],
                "summary": "Creating a segment",
                "parameters": [
                    {
                        "description": "Segment slug",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/internal_httpserver_handlers_segments_create.Request"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/internal_httpserver_handlers_segments_create.Response"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    }
                }
            }
        },
        "/segments/{slug}": {
            "get": {
                "tags": [
                    "segments"
                ],
                "summary": "Getting a segment",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Segment slug",
                        "name": "slug",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/internal_httpserver_handlers_segments_get.Response"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "segments"
                ],
                "summary": "Deleting a segment",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Segment slug",
                        "name": "slug",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    }
                }
            }
        },
        "/users": {
            "post": {
                "tags": [
                    "users"
                ],
                "summary": "Creating a user",
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/internal_httpserver_handlers_users_create.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    }
                }
            }
        },
        "/users/{user-id}/download-segments-history": {
            "get": {
                "produces": [
                    "text/csv json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Downloading user segments history",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "user-id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "example": "2023-09",
                        "description": "Year and month",
                        "name": "period",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    }
                }
            }
        },
        "/users/{user-id}/segments": {
            "get": {
                "tags": [
                    "users"
                ],
                "summary": "Getting user segments",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "user-id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/internal_httpserver_handlers_users_get.Response"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    }
                }
            },
            "patch": {
                "tags": [
                    "users"
                ],
                "summary": "Updating user segments",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "user-id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Segments to add/remove",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/internal_httpserver_handlers_users_update.Request"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/segmentify_internal_lib_response.ErrResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "internal_httpserver_handlers_segments_create.Request": {
            "type": "object",
            "required": [
                "slug"
            ],
            "properties": {
                "slug": {
                    "type": "string"
                }
            }
        },
        "internal_httpserver_handlers_segments_create.Response": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                },
                "slug": {
                    "type": "string"
                }
            }
        },
        "internal_httpserver_handlers_segments_get.Response": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                },
                "slug": {
                    "type": "string"
                }
            }
        },
        "internal_httpserver_handlers_users_create.Response": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                }
            }
        },
        "internal_httpserver_handlers_users_get.Response": {
            "type": "object",
            "properties": {
                "user-id": {
                    "type": "integer"
                },
                "user-segments": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "internal_httpserver_handlers_users_update.Request": {
            "type": "object",
            "required": [
                "segments_to_add",
                "segments_to_remove"
            ],
            "properties": {
                "segments_to_add": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/segmentify_internal_models.SegmentToAdd"
                    }
                },
                "segments_to_remove": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "segmentify_internal_lib_response.ErrResponse": {
            "type": "object",
            "properties": {
                "detail": {
                    "type": "string"
                }
            }
        },
        "segmentify_internal_models.SegmentToAdd": {
            "type": "object",
            "required": [
                "slug"
            ],
            "properties": {
                "exprire_at": {
                    "type": "string"
                },
                "slug": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Segmentify",
	Description:      "Dynamic user segmentation service",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
