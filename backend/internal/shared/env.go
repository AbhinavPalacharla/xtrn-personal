package shared

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() (bool, error) {
	envPath := os.Getenv("ENV_PATH")

	if envPath == "" {
		return false, errors.New("No `ENV_PATH` env variable set")
	}

	err := godotenv.Load(envPath)

	if err != nil {
		return false, fmt.Errorf("Failed to load env variables from %s - %w\n", envPath, err)
	}

	return true, nil
}

func GetEnv(name string) (string, error) {
	value := os.Getenv(name)

	if value == "" {
		return "", fmt.Errorf("`%s` env variable not defined", name)
	}

	return value, nil
}

func init() {
	if ok, err := LoadEnv(); !ok {
		panic(err)
	}
}
