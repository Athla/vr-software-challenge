package config_test

import (
	"testing"

	"github.com/Athla/vr-software-challenge/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Errorf("Unable to load config due: %s", err)
		return
	}
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestValidateConfig(t *testing.T) {
	cfg, _ := config.Load()
	err := cfg.Validate()
	assert.NoError(t, err)
}
