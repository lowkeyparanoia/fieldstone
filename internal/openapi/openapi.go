// Package openapi provides OpenAPI/Swagger spec generation from Go code.
// This implements the "OpenAPI/Swagger spec" feature requested in PocketBase (68 reactions).
// It uses reflection to analyze handlers and generate API documentation automatically.
package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/fieldstone/fieldstone/pkg/models"
)

// Spec represents an OpenAPI 3.0 specification
type Spec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       Info                   `json:"info"`
	Servers    []Server               `json:"servers,omitempty"`
	Paths      map[string]*PathItem   `json:"paths"`
	Components Components             `json:"components"`
}

// Info provides metadata about the API
type Info struct {
	Title          string  `json:"title"`
	Description    string  `json:"description,omitempty"`
	Version        string  `json:"version"`
	TermsOfService string  `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

// Contact information
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License information
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server represents a server URL
type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable for URL template substitution
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// PathItem describes operations available on a path
type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Options *Operation `json:"options,omitempty"`
	Head    *Operation `json:"head,omitempty"`
}

// Operation describes a single API operation
type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]*Response  `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// Parameter describes an operation parameter
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, header, path, cookie
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
}

// RequestBody describes request body
type RequestBody struct {
	Description string                `json:"description,omitempty"`
	Content     map[string]MediaType  `json:"content"`
	Required    bool                  `json:"required,omitempty"`
}

// MediaType describes media type object
type MediaType struct {
	Schema   *Schema             `json:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty"`
}

// Example object
type Example struct {
	Summary       string      `json:"summary,omitempty"`
	Description   string      `json:"description,omitempty"`
	Value         interface{} `json:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty"`
}

// Response describes response
type Response struct {
	Description string               `json:"description"`
	Headers     map[string]Header    `json:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// Header object
type Header struct {
	Description string  `json:"description,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

// Schema object
type Schema struct {
	Type                 string              `json:"type,omitempty"`
	Format               string              `json:"format,omitempty"`
	Description          string              `json:"description,omitempty"`
	Properties           map[string]*Schema  `json:"properties,omitempty"`
	Required             []string            `json:"required,omitempty"`
	Items                *Schema             `json:"items,omitempty"`
	Ref                  string              `json:"$ref,omitempty"`
	Enum                 []interface{}       `json:"enum,omitempty"`
	Default              interface{}         `json:"default,omitempty"`
	Example              interface{}         `json:"example,omitempty"`
	MinLength            int                 `json:"minLength,omitempty"`
	MaxLength            int                 `json:"maxLength,omitempty"`
	Minimum              *float64            `json:"minimum,omitempty"`
	Maximum              *float64            `json:"maximum,omitempty"`
	Pattern              string              `json:"pattern,omitempty"`
	AdditionalProperties *bool               `json:"additionalProperties,omitempty"`
}

// Components holds reusable objects
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty"`
	Responses       map[string]*Response       `json:"responses,omitempty"`
	Parameters      map[string]*Parameter      `json:"parameters,omitempty"`
	RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme defines security scheme
type SecurityScheme struct {
	Type             string      `json:"type"`
	Description      string      `json:"description,omitempty"`
	Name             string      `json:"name,omitempty"`
	In               string      `json:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows configuration
type OAuthFlows struct {
	Implicit    *OAuthFlow `json:"implicit,omitempty"`
	Password    *OAuthFlow `json:"password,omitempty"`
	ClientCred  *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthCode    *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow describes OAuth flow
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// Generator creates OpenAPI specs from Go types
type Generator struct {
	spec *Spec
}

// NewGenerator creates a new OpenAPI generator
func NewGenerator(title, version, description string) *Generator {
	return &Generator{
		spec: &Spec{
			OpenAPI: "3.0.3",
			Info: Info{
				Title:       title,
				Version:     version,
				Description: description,
				Contact: &Contact{
					Name:  "Fieldstone Support",
					Email: "support@fieldstone.io",
				},
				License: &License{
					Name: "MIT",
					URL:  "https://opensource.org/licenses/MIT",
				},
			},
			Servers: []Server{
				{
					URL:         "http://localhost:8090",
					Description: "Local development server",
				},
				{
					URL:         "https://api.fieldstone.io",
					Description: "Production server",
				},
			},
			Paths: make(map[string]*PathItem),
			Components: Components{
				Schemas: make(map[string]*Schema),
				SecuritySchemes: map[string]*SecurityScheme{
					"bearerAuth": {
						Type:         "http",
						Scheme:       "bearer",
						BearerFormat: "JWT",
						Description:  "JWT token authentication",
					},
				},
			},
		},
	}
}

// AddPath adds a path with operation
func (g *Generator) AddPath(path, method string, op *Operation) {
	if _, exists := g.spec.Paths[path]; !exists {
		g.spec.Paths[path] = &PathItem{}
	}

	pathItem := g.spec.Paths[path]
	switch strings.ToUpper(method) {
	case "GET":
		pathItem.Get = op
	case "POST":
		pathItem.Post = op
	case "PUT":
		pathItem.Put = op
	case "DELETE":
		pathItem.Delete = op
	case "PATCH":
		pathItem.Patch = op
	case "OPTIONS":
		pathItem.Options = op
	case "HEAD":
		pathItem.Head = op
	}
}

// AddSchema adds a schema to components
func (g *Generator) AddSchema(name string, schema *Schema) {
	g.spec.Components.Schemas[name] = schema
}

// SchemaFromType generates schema from Go type using reflection
func (g *Generator) SchemaFromType(v interface{}) *Schema {
	t := reflect.TypeOf(v)
	
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		return g.schemaFromStruct(t)
	case reflect.Slice, reflect.Array:
		return &Schema{
			Type:  "array",
			Items: g.SchemaFromType(reflect.New(t.Elem()).Interface()),
		}
	case reflect.Map:
		return &Schema{
			Type:                 "object",
			AdditionalProperties: boolPtr(true),
		}
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{Type: "integer"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	default:
		return &Schema{Type: "object"}
	}
}

func (g *Generator) schemaFromStruct(t reflect.Type) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
			// Check if required
			for _, part := range parts[1:] {
				if part == "omitempty" {
					// Not required
					break
				}
			}
		}

		fieldSchema := g.fieldSchema(field.Type)
		
		// Add description from tag
		if desc := field.Tag.Get("description"); desc != "" {
			fieldSchema.Description = desc
		}

		schema.Properties[fieldName] = fieldSchema
	}

	return schema
}

func (g *Generator) fieldSchema(t reflect.Type) *Schema {
	// Handle special types
	switch t {
	case reflect.TypeOf(time.Time{}):
		return &Schema{
			Type:   "string",
			Format: "date-time",
		}
	}

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		return g.fieldSchema(t.Elem())
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{Type: "integer"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Slice, reflect.Array:
		return &Schema{
			Type:  "array",
			Items: g.fieldSchema(t.Elem()),
		}
	case reflect.Map:
		return &Schema{
			Type:                 "object",
			AdditionalProperties: boolPtr(true),
		}
	case reflect.Struct:
		return g.schemaFromStruct(t)
	default:
		return &Schema{Type: "object"}
	}
}

// GenerateJSON returns the spec as JSON
func (g *Generator) GenerateJSON() ([]byte, error) {
	return json.MarshalIndent(g.spec, "", "  ")
}

// ServeHTTP serves the OpenAPI spec
func (g *Generator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := g.GenerateJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// BuildDefaultSpec creates the default Fieldstone API spec
func BuildDefaultSpec() *Generator {
	gen := NewGenerator(
		"Fieldstone API",
		"1.0.0",
		"Open source backend in a single Go binary. PocketBase-compatible with multi-tenancy and PostgreSQL support.",
	)

	// Add Collection schemas
	gen.AddSchema("Collection", gen.SchemaFromType(models.Collection{}))
	gen.AddSchema("Record", gen.SchemaFromType(models.Record{}))
	gen.AddSchema("User", gen.SchemaFromType(models.User{}))
	gen.AddSchema("Tenant", gen.SchemaFromType(models.Tenant{}))
	gen.AddSchema("QueryResult", gen.SchemaFromType(models.QueryResult{}))

	// Add auth endpoints
	gen.AddPath("/api/auth/register", "POST", &Operation{
		Tags:        []string{"Authentication"},
		Summary:     "Register a new user",
		OperationID: "authRegister",
		RequestBody: &RequestBody{
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{
						Type: "object",
						Properties: map[string]*Schema{
							"email":    {Type: "string", Format: "email"},
							"password": {Type: "string", MinLength: 8},
						},
						Required: []string{"email", "password"},
					},
				},
			},
		},
		Responses: map[string]*Response{
			"201": {
				Description: "User created successfully",
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{Ref: "#/components/schemas/User"},
					},
				},
			},
			"400": {Description: "Invalid input"},
			"409": {Description: "User already exists"},
		},
	})

	gen.AddPath("/api/auth/login", "POST", &Operation{
		Tags:        []string{"Authentication"},
		Summary:     "Authenticate user",
		OperationID: "authLogin",
		RequestBody: &RequestBody{
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{
						Type: "object",
						Properties: map[string]*Schema{
							"email":    {Type: "string"},
							"password": {Type: "string"},
						},
						Required: []string{"email", "password"},
					},
				},
			},
		},
		Responses: map[string]*Response{
			"200": {Description: "Login successful"},
			"401": {Description: "Invalid credentials"},
		},
	})

	// Add collections endpoints
	gen.AddPath("/api/collections", "GET", &Operation{
		Tags:        []string{"Collections"},
		Summary:     "List all collections",
		OperationID: "listCollections",
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Responses: map[string]*Response{
			"200": {
				Description: "List of collections",
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]*Schema{
								"items": {
									Type:  "array",
									Items: &Schema{Ref: "#/components/schemas/Collection"},
								},
							},
						},
					},
				},
			},
			"401": {Description: "Unauthorized"},
		},
	})

	gen.AddPath("/api/collections/{id}", "GET", &Operation{
		Tags:        []string{"Collections"},
		Summary:     "Get collection by ID",
		OperationID: "getCollection",
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Parameters: []Parameter{
			{
				Name:        "id",
				In:          "path",
				Description: "Collection ID",
				Required:    true,
				Schema:      &Schema{Type: "string"},
			},
		},
		Responses: map[string]*Response{
			"200": {
				Description: "Collection found",
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{Ref: "#/components/schemas/Collection"},
					},
				},
			},
			"404": {Description: "Collection not found"},
		},
	})

	return gen
}

func boolPtr(b bool) *bool {
	return &b
}
