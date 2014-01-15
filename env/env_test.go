package env

import (
	"fmt"
	"testing"
	"os"
)

func TestClearedEnv(t *testing.T) {
	os.Clearenv()
	env := Map()
	if len(env) > 0 {
		t.Error("env.Map() didn't return an empty map")
	}
}

func TestEnvNotNil(t *testing.T) {
	env := Map()
	if env == nil {
		t.Error("env.Map() returned nil!")
	}
}

func ExampleMap() {
	for k, v := range Map() {
		fmt.Printf("%s=%s\n", k, v)
	}
}
