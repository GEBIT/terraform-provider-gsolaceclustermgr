package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	// assert equality
	assert.Equal(t, 123, 123, "they should be equal")
}

func TestGetRouterPrefix1(t *testing.T) {
	// assert equality
	assert.Equal(t, "test123", getRouterPrefix("test123primarycn"), "primarycn suffix")
	assert.Equal(t, "test123", getRouterPrefix("test123primarycn"), "primary suffix")
	assert.Equal(t, "test123", getRouterPrefix("test123monitoring"), "monitor suffix")
	assert.Equal(t, "test123", getRouterPrefix("test123backup"), "backup suffix")
	assert.Equal(t, "test123unexpected", getRouterPrefix("test123unexpected"), "not matching")
}
