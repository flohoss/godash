package validate

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

type sample struct {
	Host  string `mapstructure:"host" validate:"required"`
	Port  int    `mapstructure:"port" validate:"gte=1024,lte=65535"`
	Plain string
}

func TestStruct_AcceptsAValidValue(t *testing.T) {
	if err := Struct(sample{Host: "example.com", Port: 8080}); err != nil {
		t.Fatalf("Struct() = %v, want nil", err)
	}
}

// Field problems are reported with the config path and the failing rule so an
// operator can fix the YAML without guessing.
func TestStruct_ReportsConfigPathAndRule(t *testing.T) {
	err := Struct(sample{Port: 80})
	if err == nil {
		t.Fatal("expected a validation error")
	}

	message := err.Error()
	if !strings.Contains(message, "host") {
		t.Errorf("expected the mapstructure name in %q", message)
	}
	if !strings.Contains(message, "required") {
		t.Errorf("expected the failing rule in %q", message)
	}
	if !strings.HasPrefix(message, "- ") {
		t.Errorf("expected a bulleted line, got %q", message)
	}
}

func TestStruct_ReportsEveryFieldError(t *testing.T) {
	err := Struct(sample{})
	if err == nil {
		t.Fatal("expected a validation error")
	}

	if got := strings.Count(err.Error(), "\n") + 1; got != 2 {
		t.Errorf("expected 2 error lines, got %d in %q", got, err.Error())
	}
}

// A field without a mapstructure tag still gets a readable name via camelCase.
func TestStruct_FallsBackToCamelCaseName(t *testing.T) {
	type tagged struct {
		Plain string `validate:"required"`
	}

	err := Struct(tagged{})
	if err == nil {
		t.Fatal("expected a validation error")
	}
	if !strings.Contains(err.Error(), "plain") {
		t.Errorf("expected camelCase field name in %q", err.Error())
	}
}

// Errors that are not field validation errors must pass through untouched so
// callers keep the original cause instead of a wrapped copy.
func TestPretty_PassesThroughNonValidationErrors(t *testing.T) {
	want := errors.New("boom")

	if got := pretty(want); got != want {
		t.Errorf("pretty() = %v, want the original error", got)
	}
}

func TestConfigFieldName(t *testing.T) {
	type tagged struct {
		Mapped   string `mapstructure:"custom_name"`
		Ignored  string `mapstructure:"-"`
		CommaTag string `mapstructure:"first,omitempty"`
		NoTag    string
		HTTPPath string
	}

	typ := reflect.TypeOf(tagged{})
	want := map[string]string{
		"Mapped":   "custom_name",
		"Ignored":  "ignored",
		"CommaTag": "first",
		"NoTag":    "noTag",
		"HTTPPath": "hTTPPath",
	}

	for name, wantName := range want {
		field, ok := typ.FieldByName(name)
		if !ok {
			t.Fatalf("field %q not found", name)
		}
		if got := configFieldName(field); got != wantName {
			t.Errorf("configFieldName(%s) = %q, want %q", name, got, wantName)
		}
	}
}

func TestCamelCase(t *testing.T) {
	tests := map[string]string{
		"":              "",
		"Name":          "name",
		"AppUrl":        "appUrl",
		"HTTPServerURL": "hTTPServerURL",
	}

	for in, want := range tests {
		if got := camelCase(in); got != want {
			t.Errorf("camelCase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestConfigPath_DropsTheRootNamespace(t *testing.T) {
	if got := configPath("GlobalConfig.Server.AppUrl"); got != "Server.AppUrl" {
		t.Errorf("configPath() = %q, want %q", got, "Server.AppUrl")
	}
	if got := configPath("OnlyRoot"); got != "OnlyRoot" {
		t.Errorf("configPath() = %q, want %q", got, "OnlyRoot")
	}
}

// The generated message is human-readable, not the raw validator text.
func TestFormatFieldError_IsTranslated(t *testing.T) {
	var fieldErrors validator.ValidationErrors
	err := validate.Struct(sample{Port: 80})
	if !errors.As(err, &fieldErrors) {
		t.Fatalf("expected validator.ValidationErrors, got %v", err)
	}

	got := formatFieldError(fieldErrors[0])
	if got == "" {
		t.Fatal("expected a non-empty message")
	}
	if strings.Contains(got, "Error:Field validation") {
		t.Errorf("expected a translated message, got %q", got)
	}
}
