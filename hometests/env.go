package hometests

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

// EnvOverride overrides an environment variable for the duration of the test.
// Returns a closer func to change the state to the previous one after test is done.
func EnvOverride(t *testing.T, key, value string) func() {
	t.Helper()

	originalValue := os.Getenv(key)

	t.Setenv(key, value)

	return func() {
		if err := os.Setenv(key, originalValue); err != nil {
			t.Fatal(err)
		}
	}
}

// EnvOverrideMany overrides multiple environment variables for the duration of the test.
// Returns a closer func to change the state to the previous one after test is done.
func EnvOverrideMany(t *testing.T, envs map[string]string) func() {
	t.Helper()

	originalEnvs := map[string]string{}

	for key, value := range envs {
		if originalValue, ok := os.LookupEnv(key); ok {
			originalEnvs[key] = originalValue
		}

		_ = os.Setenv(key, value)
	}

	return func() {
		for key := range envs {
			origValue, has := originalEnvs[key]
			if has {
				_ = os.Setenv(key, origValue)
			} else {
				_ = os.Unsetenv(key)
			}
		}
	}
}

// LoadEnvFromTheFile loads env variables from the file.
// File will be searched in the current directory and up the tree.
// If override is true, it overrides the existing env variables.
func LoadEnvFromTheFile(t *testing.T, fileName string, override bool) {
	t.Helper()

	p, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	envPath, err := findFile(p, fileName)
	if err != nil {
		t.Fatal(err)
	}

	if override {
		if err = godotenv.Overload(envPath); err != nil {
			t.Fatal(err)
		}
	} else {
		if err = godotenv.Load(envPath); err != nil {
			t.Fatal(err)
		}
	}
}

func findFile(fPath, fName string) (string, error) {
	if fPath == "/" {
		return "", errors.New("could not find env file")
	}

	_, err := os.Stat(path.Join(fPath, fName))
	if err == nil {
		return path.Join(fPath, fName), nil
	}

	return findFile(filepath.Join(fPath, ".."), fName)
}
