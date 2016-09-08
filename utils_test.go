package cache

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateFileName(t *testing.T) {
	key := "test"
	name := generateFileName(key)
	assert.Equal(t, filepath.Join("6", "4f", "098f6bcd4621d373cade4e832627b4f6"), name)
}
