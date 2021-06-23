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
		"DEV_CLUSTER":      "dev-cluster",
		"REGISTRY_NAME":    "kind-registry",
		"REGISTRY_PORT":    "5000",
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

	volumes, err := sh.OutputWith(env, "docker", "volume", "ls")
	if err != nil {
		return fmt.Errorf("could not create volume: %s", err)
	}

	networks, err := sh.OutputWith(env, "docker", "network", "ls")
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
		return fmt.Errorf("failed to stop server: %s", err)
	}

	err = DB.Stop(DB{})
	if err != nil {
		return fmt.Errorf("could not stop database: %s", err)
	}

	err = sh.RunWith(env, "docker", "rm", "chatter_server")
	if err != nil {
		return fmt.Errorf("failed to remove server: %s", err)
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

type Kube mg.Namespace

// creates the development kubernetes cluster
func (Kube) CreateDev() error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get root directory: %s", err)
	}

	l := filepath.Join(path, "kubernetes")

	err = os.Chdir(l)
	if err != nil {
		return fmt.Errorf("could not change directories: %s", err)
	}

	err = createClusterRepo()
	if err != nil {
		return fmt.Errorf("could not create cluster repository: %s", err)
	}

	err = sh.Run("kind", "create", "cluster", "--config", "cluster.yaml")
	if err != nil {
		return fmt.Errorf("failed starting dev cluster: %s", err)
	}

	endpoints, err := sh.OutputWith(env, "docker", "network", "inspect", "kind")
	if err != nil {
		return fmt.Errorf("could not get containers: %s", err)
	}

	if !strings.Contains(endpoints, env["REGISTRY_NAME"]) {
		err = sh.RunWith(env, "docker", "network", "connect", "\"kind\"", "\"$REGISTRY_NAME\"")
		if err != nil {
			return fmt.Errorf("failed starting registry: %s", err)
		}

		return nil
	}

	err = os.Chdir(path)
	if err != nil {
		return fmt.Errorf("could not change directories: %s", err)
	}

	return nil
}

func createClusterRepo() error {
	containers, err := sh.OutputWith(env, "docker", "container", "ls", "-a")
	if err != nil {
		return fmt.Errorf("could not get containers: %s", err)
	}

	if !strings.Contains(containers, env["REGISTRY_NAME"]) {
		err = sh.RunWith(env, "docker", "run", "-e", "DATABASE_URL=$DATABASE_URL", "-e", "SECRET_KEY_BASE=$SECRET_KEY_BASE",
			"--network", "$POSTGRES_NETWORK", "-p", "$REGISTRY_PORT:$REGISTRY_PORT", "-d", "--name", "$REGISTRY_NAME", "registry:2")
		if err != nil {
			return fmt.Errorf("failed starting registry: %s", err)
		}

		return nil
	}

	err = sh.RunWith(env, "docker", "start", "$REGISTRY_NAME")
	if err != nil {
		return fmt.Errorf("failed starting registry: %s", err)
	}

	return nil
}

// deletes the dev kubernetes cluster
func (Kube) DeleteDev() error {
	err := sh.RunWith(env, "kind", "delete", "cluster", "--name", "$DEV_CLUSTER")
	if err != nil {
		return fmt.Errorf("failed starting dev cluster: %s", err)
	}

	err = sh.RunWith(env, "docker", "stop", "$REGISTRY_NAME")
	if err != nil {
		return fmt.Errorf("failed starting registry: %s", err)
	}

	return nil
}

// deploys postgres and the chatter server into kubernetes
func (Kube) DeployServices() error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get root directory: %s", err)
	}

	l := filepath.Join(path, "kubernetes", "postgres")

	err = os.Chdir(l)
	if err != nil {
		return fmt.Errorf("could not change directories: %s", err)
	}

	err = sh.RunWith(env, "helm", "install", "postgres", ".")
	if err != nil {
		return fmt.Errorf("failed installing postgres: %s", err)
	}

	return nil
}

// pushes the init and chatter service to the kind repo
func (Kube) Push() error {
	err := sh.RunWith(env, "docker", "tag", "$APP_NAME:latest", "localhost:$REGISTRY_PORT/$APP_NAME:latest")
	if err != nil {
		return fmt.Errorf("failed tagging docker image:  %s", err)
	}

	err = sh.RunWith(env, "docker", "tag", "$INIT_NAME:latest", "localhost:$REGISTRY_PORT/$INIT_NAME:latest")
	if err != nil {
		return fmt.Errorf("failed tagging docker image:  %s", err)
	}

	err = sh.RunWith(env, "docker", "push", "localhost:$REGISTRY_PORT/$APP_NAME:latest")
	if err != nil {
		return fmt.Errorf("failed pusing docker image:  %s", err)
	}

	err = sh.RunWith(env, "docker", "push", "localhost:5000/phoenix-init:latest")
	if err != nil {
		return fmt.Errorf("failed pusing docker image:  %s", err)
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
		return fmt.Errorf("failed pusing docker image: %s", err)
	}

	err = sh.RunWith(env, "docker", "push", "ghcr.io/$GITHUB_OWNER/$INIT_NAME:latest")
	if err != nil {
		return fmt.Errorf("failed pusing docker image:  %s", err)
	}

	err = sh.RunWith(env, "docker", "push", "ghcr.io/$GITHUB_OWNER/$APP_NAME:$GIT_COMMIT")
	if err != nil {
		return fmt.Errorf("failed pusing docker image:  %s", err)
	}

	err = sh.RunWith(env, "docker", "push", "ghcr.io/$GITHUB_OWNER/$INIT_NAME:$GIT_COMMIT")
	if err != nil {
		return fmt.Errorf("failed pusing docker image:  %s", err)
	}

	return nil
}
