package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	var called bool
	t.Setenv("APP_ENV", "test")
	AddConfigPath("testdata")
	OnConfigChange(func() { called = true })
	InitConfig(context.Background())
	assert.Equal(t, "test", Registry.Get("APP_ENV"))
	assert.Equal(t, "test", Registry.Get("file"))
	assert.False(t, called)
}
