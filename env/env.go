package env

import (
	"os"
	"strings"
)

// Get the environment as a map[string]string
func Map() map[string]string {
	env := make(map[string]string)
	for _, v := range os.Environ() {
		kv := strings.Split(v, "=")
		env[kv[0]] = kv[1]
	}
	return env
}
