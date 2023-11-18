package envers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGivenWhenNewConfigThenReturnConfigWithShowSQLFalse(t *testing.T) {
	// Given / When
	config := NewConfig()
	// Then
	assert.False(t, config.ShowSQL, "NewConfig mus return show SQL false")
}
