package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPortParser(t *testing.T) {
	key := "PORT"
	var err error
	defer func() {
		os.Unsetenv(key)
	}()

	os.Setenv(key, "1024")
	_, err = Parse()
	assert.Equal(t, err, nil, "Parsing should pass")

	os.Setenv(key, "-12")
	_, err = Parse()
	assert.Equal(t, err.Error(), "Key: 'Config.Port' Error:Field validation for 'Port' failed on the 'min' tag", "Validation should fail")

	os.Setenv(key, "60000")
	_, err = Parse()
	assert.Equal(t, err.Error(), "Key: 'Config.Port' Error:Field validation for 'Port' failed on the 'max' tag", "Validation should fail")

	os.Setenv(key, "abc")
	_, err = Parse()
	assert.Equal(t, err.Error(), "env: parse error on field \"Port\" of type \"int\": strconv.ParseInt: parsing \"abc\": invalid syntax", "Parsing should fail")
}

func TestTimeZoneParser(t *testing.T) {
	key := "TZ"
	var err error
	defer func() {
		os.Unsetenv(key)
	}()

	os.Setenv(key, "Europe/Berlin")
	_, err = Parse()
	assert.Equal(t, err, nil, "Parsing should pass")

	os.Setenv(key, "Etc/UTC")
	_, err = Parse()
	assert.Equal(t, err, nil, "Parsing should pass")

	os.Setenv(key, "abc")
	_, err = Parse()
	assert.Equal(t, err.Error(), "Key: 'Config.TimeZone' Error:Field validation for 'TimeZone' failed on the 'timezone' tag", "Validation should fail")

	os.Setenv(key, "-1")
	_, err = Parse()
	assert.Equal(t, err.Error(), "Key: 'Config.TimeZone' Error:Field validation for 'TimeZone' failed on the 'timezone' tag", "Validation should fail")
}
