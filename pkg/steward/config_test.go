package steward

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	c := NewConfig()
	assert.NotNil(t, c)

	options := c.getOptions()
	assert.Len(t, options, 5)
}
