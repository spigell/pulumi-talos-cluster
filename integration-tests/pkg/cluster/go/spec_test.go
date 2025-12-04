package cluster

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("does-not-exist.yaml")
	require.Error(t, err)
}

func TestLoadMalformedYAML(t *testing.T) {
	path := filepath.Join("..", "fixtures", "load-malformed.yaml")

	_, err := Load(path)
	require.Error(t, err)
}
