package envers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGivenConfigWhenNewGormEnversPluginThenReturnNotNull(t *testing.T) {
	// Given
	config := NewConfig()
	// When
	actual := NewGormEnversPlugin(config)
	// Then
	assert.NotNil(t, actual, "NewGormEnversPlugin must return not nil")
}

func TestValidDbWhenInitializeThenReturnNil(t *testing.T) {
	// Given
	gormEnversPlugin := NewGormEnversPlugin(NewConfig())
	cxn := "file::memory:?mode=memory&cache=shared" //memory
	// When
	_, err := gorm.Open(sqlite.Open(cxn), &gorm.Config{
		Plugins: map[string]gorm.Plugin{gormEnversPlugin.Name(): gormEnversPlugin},
	})
	// Then
	assert.Nil(t, err, "Initialize must return nil error")
}

func TestGivenPluginWhenNameThenReturnExpectedValue(t *testing.T) {
	// Given
	gormEnversPlugin := NewGormEnversPlugin(NewConfig())
	// When
	actual := gormEnversPlugin.Name()
	// Then
	assert.Equal(t, actual, "GORM-ENVERS")
}
