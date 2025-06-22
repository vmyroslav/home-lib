package hometests

import (
	"os"
	"testing"
)

func TestEnvOverride(t *testing.T) {
	t.Run("should override environment variable and restore it after test", func(t *testing.T) {
		key := t.Name() + "_TEST_KEY"
		value := "TEST_VALUE"
		originalValue := "ORIGINAL_VALUE"

		_ = os.Setenv(key, originalValue)
		defer os.Unsetenv(key)

		closer := EnvOverride(t, key, value)

		if os.Getenv(key) != value {
			t.Errorf("Expected environment variable %s to be %s, but got %s", key, value, os.Getenv(key))
		}

		closer()

		if os.Getenv(key) != originalValue {
			t.Errorf("Expected environment variable %s to be %s, but got %s", key, originalValue, os.Getenv(key))
		}
	})
}

func TestEnvOverrideMany(t *testing.T) {
	t.Parallel()

	t.Run("should override multiple environment variables and restore them after test", func(t *testing.T) {
		envs := map[string]string{
			t.Name() + "_TEST_KEY1": "TEST_VALUE1",
			t.Name() + "_TEST_KEY2": "TEST_VALUE2",
		}
		originalEnvs := map[string]string{
			"TEST_KEY1": "ORIGINAL_VALUE1",
			"TEST_KEY2": "ORIGINAL_VALUE2",
		}

		for key, value := range originalEnvs {
			os.Setenv(key, value)
			defer os.Unsetenv(key) //nolint:errcheck,gocritic
		}

		closer := EnvOverrideMany(t, envs)

		for key, value := range envs {
			if os.Getenv(key) != value {
				t.Errorf("Expected environment variable %s to be %s, but got %s", key, value, os.Getenv(key))
			}
		}

		closer()

		for key, value := range originalEnvs {
			if os.Getenv(key) != value {
				t.Errorf("Expected environment variable %s to be %s, but got %s", key, value, os.Getenv(key))
			}
		}
	})
}

func TestLoadEnvFromTheFile(t *testing.T) {
	t.Parallel()

	t.Run("should load environment variables from the file", func(t *testing.T) {
		fileName := ".env.test"
		key := "TEST_KEY"
		value := "TEST_VALUE"

		file, err := os.Create(fileName)
		if err != nil {
			t.Fatal(err)
		}

		_, err = file.WriteString(key + "=" + value)
		if err != nil {
			t.Fatal(err)
		}

		file.Close()

		LoadEnvFromTheFile(t, fileName, true)

		if os.Getenv(key) != value {
			t.Errorf("Expected environment variable %s to be %s, but got %s", key, value, os.Getenv(key))
		}

		os.Remove(fileName)
	})
}
