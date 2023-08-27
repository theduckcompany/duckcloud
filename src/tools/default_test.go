package tools

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools/clock"
	"github.com/theduckcompany/duckcloud/src/tools/password"
	"github.com/theduckcompany/duckcloud/src/tools/response"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func TestDefaultToolbox(t *testing.T) {
	tools := NewToolbox(Config{})

	assert.IsType(t, new(clock.Default), tools.Clock())
	assert.IsType(t, new(uuid.Default), tools.UUID())
	assert.IsType(t, new(response.Default), tools.ResWriter())
	assert.IsType(t, new(slog.Logger), tools.Logger())
	assert.IsType(t, new(password.BcryptPassword), tools.Password())
}
