package cache

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("CACHE_ALLOW_IN_MEMORY_FALLBACK", "true")
	code := m.Run()
	_ = os.Unsetenv("CACHE_ALLOW_IN_MEMORY_FALLBACK")
	os.Exit(code)
}
