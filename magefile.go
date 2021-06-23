// +build mage

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	env = map[string]string{
		"POSTGRES_NETWORK": "phoenix-postgres",
		"POSTGRES_VOLUME":  "phoenix-postgres",
		"GITHUB_OWNER":     "rallinator7",
		"APP_NAME":         "chatter",
		"GIT_COMMIT":       gitCommit(),
		"INIT_NAME":        "phoenix-init",
		"DATABASE_URL":     "ecto://phoenix:phoenix@db:5432/phoenix",
		"SECRET_KEY_BASE":  "JhhLO9oACpINDgzWo9xBWw+qKCrh7C6tzUhBo4rMGCbB51ssgPzZpkL812d12fL1",
	}
)

func gitCommit() string {
	commit, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		log.Fatalf("could not get commit hash: %s", err)
	}

	return commit
}

func mixDependencies() error {
	err := sh.Run("mix", "deps.get")
	if err != nil {
		return fmt.Errorf("failed setting up dependencies: %s", err)
	}

	return nil
}

// initial set up for the repository
func Configure() error {
	err := mixDependencies()
	if err != nil {
		return fmt.Errorf("error creating compose requirements: %s", err)
	}

	err = createComposeReqs()
	if err != nil {
		return fmt.Errorf("error creating compose requirements: %s", err)
	}

	err = Docker.BuildInit(Docker{})
	if err != nil {
		return fmt.Errorf("could not start init container: %s", err)
	}

	err = DB.Start(DB{})
	if err != nil {
		return fmt.Errorf("could not start database: %s", err)
	}

	err = Docker.RunInit(Docker{})
	if err != nil {
		return fmt.Errorf("could not start init container: %s", err)
	}

	err = DB.Stop(DB{})
	if err != nil {
		return fmt.Errorf("could not stop database: %s", err)
	}

	err = Docker.BuildServer(Docker{})
	if err != nil {
		return fmt.Errorf("could not start init container: %s", err)
	}

	return nil
}

func createComposeReqs() error {

	networks, err := sh.OutputWith(env, "docker", "volume", "ls")
	if err != nil {
		return fmt.Errorf("could not create volume: %s", err)
	}

	volumes, err := sh.OutputWith(env, "docker", "network", "ls")
	if err != nil {
		return fmt.Errorf("could not create volume: %s", err)
	}

	if !strings.Contains(networks, env["POSTGRES_NETWORK"]) {
		err = sh.RunWith(env, "docker", "network", "create", "$POSTGRES_NETWORK")
		if err != nil {
			return fmt.Errorf("could not create network: %s", err)
		}
	}

	if !strings.Contains(volumes, env["POSTGRES_VOLUME"]) {
		err := sh.RunWith(env, "docker", "volume", "create", "$POSTGRES_VOLUME")
		if err != nil {
			return fmt.Errorf("could not create volume: %s", err)
		}
	}

	return nil
}

type Docker mg.Namespace

// builds the main chatter server in a container
func (Docker) BuildServer() error {

	err := sh.RunWith(env, "docker", "build", "--build-arg", "DATABASE_URL=$DATABASE_URL", "--build-arg", "SECRET_KEY_BASE=$SECRET_KEY_BASE",
		"-t", "$APP_NAME:$GIT_COMMIT", "-t", "$APP_NAME:latest", "-f", "./docker/server/Dockerfile", ".")
	if err != nil {
		return fmt.Errorf("failed building server: %s", err)
	}

	return nil
}

// builds an init container for preparing the database
func (Docker) BuildInit() error {

	err := sh.RunWith(env, "docker", "build", "--build-arg", "DATABASE_URL=$DATABASE_URL", "--build-arg", "SECRET_KEY_BASE=$SECRET_KEY_BASE",
		"-t", "$INIT_NAME:$GIT_COMMIT", "-t", "$INIT_NAME:latest", "-f", "./docker/init/Dockerfile", ".")
	if err != nil {
		return fmt.Errorf("failed building init: %s", err)
	}

	return nil
}

// starts an init container that updates postgres for any new migrations
func (Docker) RunInit() error {

	err := sh.RunWith(env, "docker", "run", "-e", "DATABASE_URL=$DATABASE_URL", "-e", "SECRET_KEY_BASE=$SECRET_KEY_BASE",
		"--network", "$POSTGRES_NETWORK", "$INIT_NAME:latest")
	if err != nil {
		return fmt.Errorf("failed running init: %s", err)
	}

	return nil
}

// starts the chatter server
func (Docker) RunServer() error {

	err := DB.Start(DB{})
	if err != nil {
		return fmt.Errorf("could not start database: %s", err)
	}

	err = sh.RunWith(env, "docker", "run", "-e", "DATABASE_URL=$DATABASE_URL", "-e", "SECRET_KEY_BASE=$SECRET_KEY_BASE",
		"--network", "$POSTGRES_NETWORK", "-p", "4000:4000", "-d", "--name", "chatter_server", "$APP_NAME:latest")
	if err != nil {
		return fmt.Errorf("failed starting server: %s", err)
	}

	return nil
}

// stops the chatter server
func (Docker) StopServer() error {

	err := sh.RunWith(env, "docker", "stop", "chatter_server")
	if err != nil {
		return fmt.Errorf("failed tests: %s", err)
	}

	err = DB.Stop(DB{})
	if err != nil {
		return fmt.Errorf("could not stop database: %s", err)
	}

	return nil
}

type DB mg.Namespace

//starts the database
func (DB) Start() error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get root directory: %s", err)
	}

	l := filepath.Join(path, "docker", "postgres")

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

//stops the database
func (DB) Stop() error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get root directory: %s", err)
	}

	l := filepath.Join(path, "docker", "postgres")

	err = os.Chdir(l)
	if err != nil {
		return fmt.Errorf("could not change directories: %s", err)
	}

	o := runtime.GOOS
	switch o {
	case "windows":
		err = sh.Run("docker", "compose", "down")
		if err != nil {
			return fmt.Errorf("could not run docker compose: %s", err)
		}
	case "darwin":
		err = sh.Run("docker", "compose", "down")
		if err != nil {
			return fmt.Errorf("could not run docker compose: %s", err)
		}
	case "linux":
		err = sh.Run("docker-compose", "down")
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

type CI mg.Namespace

// runs unit tests for the chatter server
func (CI) UnitTest() error {
	err := sh.Run("mix", "deps.get")
	if err != nil {
		return fmt.Errorf("failed setting up dependencies: %s", err)
	}

	err = sh.RunWith(env, "mix", "test")
	if err != nil {
		return fmt.Errorf("failed running tests: %s", err)
	}

	fmt.Println("All tests passed!")

	return nil
}

// builds CI Images for pushing to ghcr
func (CI) Build() error {
	err := sh.RunWith(env, "docker", "build", "--build-arg", "DATABASE_URL=$DATABASE_URL", "--build-arg", "SECRET_KEY_BASE=$SECRET_KEY_BASE",
		"-t", "ghcr.io/$GITHUB_OWNER/$APP_NAME:$GIT_COMMIT", "-t", "ghcr.io/$GITHUB_OWNER/$APP_NAME:latest", "-f", "./docker/server/Dockerfile", ".")
	if err != nil {
		return fmt.Errorf("failed building server images: %s", err)
	}

	err = sh.RunWith(env, "docker", "build", "--build-arg", "DATABASE_URL=$DATABASE_URL", "--build-arg", "SECRET_KEY_BASE=$SECRET_KEY_BASE",
		"-t", "ghcr.io/$GITHUB_OWNER/$INIT_NAME:$GIT_COMMIT", "-t", "ghcr.io/$GITHUB_OWNER/$INIT_NAME:latest", "-f", "./docker/init/Dockerfile", ".")
	if err != nil {
		return fmt.Errorf("failed building init: %s", err)
	}

	return nil
}

// pushes images to ghcr
func (CI) Push() error {
	err := sh.RunWith(env, "docker", "push", "ghcr.io/$GITHUB_OWNER/$APP_NAME:latest")
	if err != nil {
		return fmt.Errorf("failed tests: %s", err)
	}

	err = sh.RunWith(env, "docker", "push", "ghcr.io/$GITHUB_OWNER/$INIT_NAME:latest")
	if err != nil {
		return fmt.Errorf("failed tests: %s", err)
	}

	err = sh.RunWith(env, "docker", "push", "ghcr.io/$GITHUB_OWNER/$APP_NAME:$GIT_COMMIT")
	if err != nil {
		return fmt.Errorf("failed tests: %s", err)
	}

	err = sh.RunWith(env, "docker", "push", "ghcr.io/$GITHUB_OWNER/$INIT_NAME:$GIT_COMMIT")
	if err != nil {
		return fmt.Errorf("failed tests: %s", err)
	}

	return nil
}
