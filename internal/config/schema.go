package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/message"
)

//go:embed tenant.schema.json
var schemaJSON []byte

const schemaID = "pebblr://tenant.schema.json"

// ValidateSchema validates a config file against the embedded JSON Schema.
// Returns a list of human-readable error strings (empty if valid).
func ValidateSchema(configPath string) ([]string, error) {
	c := jsonschema.NewCompiler()

	var schemaDoc any
	if err := json.Unmarshal(schemaJSON, &schemaDoc); err != nil {
		return nil, fmt.Errorf("parsing embedded schema: %w", err)
	}
	if err := c.AddResource(schemaID, schemaDoc); err != nil {
		return nil, fmt.Errorf("adding schema resource: %w", err)
	}
	sch, err := c.Compile(schemaID)
	if err != nil {
		return nil, fmt.Errorf("compiling schema: %w", err)
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var doc any
	if err := json.Unmarshal(configData, &doc); err != nil {
		return nil, fmt.Errorf("parsing config JSON: %w", err)
	}

	validationErr := sch.Validate(doc)
	if validationErr == nil {
		return nil, nil
	}

	ve, ok := validationErr.(*jsonschema.ValidationError)
	if !ok {
		return []string{validationErr.Error()}, nil
	}

	return flattenErrors(ve), nil
}

var printer = message.NewPrinter(message.MatchLanguage("en"))

// flattenErrors walks the validation error tree and collects leaf messages.
func flattenErrors(ve *jsonschema.ValidationError) []string {
	if len(ve.Causes) == 0 {
		path := "/" + strings.Join(ve.InstanceLocation, "/")
		return []string{fmt.Sprintf("%s: %s", path, ve.ErrorKind.LocalizedString(printer))}
	}
	var out []string
	for _, child := range ve.Causes {
		out = append(out, flattenErrors(child)...)
	}
	return out
}

// LoadAndValidate runs the full validation pipeline (schema + semantic) and
// returns the loaded config. This is the single validation path shared by
// both the serve command and the config validate command.
func LoadAndValidate(configPath string) (*TenantConfig, []string, error) {
	var errors []string

	// Phase 1: JSON Schema (structural).
	schemaErrors, err := ValidateSchema(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("schema validation: %w", err)
	}
	errors = append(errors, schemaErrors...)

	// If schema validation found structural issues, semantic validation
	// will likely panic or produce confusing errors — return early.
	if len(errors) > 0 {
		return nil, errors, nil
	}

	// Phase 2: Semantic validation via Load.
	cfg, err := Load(configPath)
	if err != nil {
		errors = append(errors, err.Error())
		return nil, errors, nil
	}

	return cfg, nil, nil
}
