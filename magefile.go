// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func gitCommit() (string, error) {
	commit, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("Failed to get GIT version: %s", err)
	}
	return commit, nil
}

func Configure() error {
	err := sh.Run("docker", "volume", "create", "postgres")
	if err != nil {
		return fmt.Errorf("could not create volume: %s", err)
	}

	err = sh.Run("docker", "network", "create", "chatter")
	if err != nil {
		return fmt.Errorf("could not create network: %s", err)
	}
	return nil
}

type Test mg.Namespace

// runs unit tests for the chatter server
func (Test) Unit() error {

	err := Docker.ComposeUp(Docker{})
	if err != nil {
		return fmt.Errorf("could not start containers: %s", err)
	}

	err = sh.Run("mix", "test")
	if err != nil {
		return fmt.Errorf("failed tests: %s", err)
	}

	fmt.Println("All tests passed!")

	return nil
}

type Docker mg.Namespace

// builds the main chatter server in a container
func (Docker) BuildServer() error {
	commit, err := gitCommit()
	if err != nil {
		return fmt.Errorf("err building image: %s", err)
	}

	appVersion, err := sh.Output("mix", "app.version")
	if err != nil {
		return fmt.Errorf("couldn't get app version: %s", err)
	}

	env := map[string]string{
		"APP_NAME":        "chatter",
		"APP_VSN":         appVersion,
		"GIT_COMMIT":      commit,
		"DATABASE_URL":    "ecto://phoenix:phoenix@db:5432/phoenix",
		"SECRET_KEY_BASE": "JhhLO9oACpINDgzWo9xBWw+qKCrh7C6tzUhBo4rMGCbB51ssgPzZpkL812d12fL1",
	}

	err = sh.RunWith(env, "docker", "build", "--build-arg", "DATABASE_URL=$DATABASE_URL", "--build-arg", "SECRET_KEY_BASE=$SECRET_KEY_BASE", "-t", "$APP_NAME:$GIT_COMMIT", "-t", "$APP_NAME:latest", "-f", "./docker/init/Dockerfile", ".")
	if err != nil {
		return fmt.Errorf("failed tests: %s", err)
	}

	return nil
}

// builds an init container for preparing the database
func (Docker) BuildInit() error {
	commit, err := gitCommit()
	if err != nil {
		return fmt.Errorf("err building image: %s", err)
	}

	env := map[string]string{
		"APP_NAME":        "phoenix-init",
		"GIT_COMMIT":      commit,
		"DATABASE_URL":    "ecto://phoenix:phoenix@db:5432/phoenix",
		"SECRET_KEY_BASE": "JhhLO9oACpINDgzWo9xBWw+qKCrh7C6tzUhBo4rMGCbB51ssgPzZpkL812d12fL1",
	}

	err = sh.RunWith(env, "docker", "build", "--build-arg", "DATABASE_URL=$DATABASE_URL", "--build-arg", "SECRET_KEY_BASE=$SECRET_KEY_BASE", "-t", "$APP_NAME:$GIT_COMMIT", "-t", "$APP_NAME:latest", "-f", "./docker/init/Dockerfile", ".")
	if err != nil {
		return fmt.Errorf("failed tests: %s", err)
	}

	return nil
}

//starts the database as well as the chatter server
func (Docker) ComposeUp() error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get root directory: %s", err)
	}

	l := filepath.Join(path, "docker", "server")

	err = os.Chdir(l)
	if err != nil {
		return fmt.Errorf("could not change directories: %s", err)
	}

	o := runtime.GOOS
	switch o {
	case "windows":
		err = sh.Run("docker", "compose", "up", "-d")
		if err != nil {
			return fmt.Errorf("could not run docker compose: %s", err)
		}
	case "darwin":
		err = sh.Run("docker", "compose", "up", "-d")
		if err != nil {
			return fmt.Errorf("could not run docker compose: %s", err)
		}
	case "linux":
		err = sh.Run("docker-compose", "up", "-d")
		if err != nil {
			return fmt.Errorf("could not run docker compose: %s", err)
		}
	}

	err = os.Chdir(path)
	if err != nil {
		return fmt.Errorf("could not change directories: %s", err)
	}

	return nil
}
