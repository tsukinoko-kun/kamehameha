package compose

import (
	_ "embed"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tsukinoko-kun/kamehameha/khh/config"
)

const (
	indent  = "  "
	indent2 = "    "
	indent3 = "      "
)

//go:embed journey.Dockerfile
var journeyBaseImage []byte

type (
	Compose string
)

func Compile(c *config.Config) (Compose, error) {
	ri := rand.Int32()
	dir, err := os.MkdirTemp("", fmt.Sprintf("kamehameha_compose_%d_*", ri))
	if err != nil {
		return "", errors.Join(errors.New("failed to create temporary directory"), err)
	}

	journeyBaseDir := filepath.Join(dir, "_journeys", "_base")
	if err := os.MkdirAll(journeyBaseDir, 0755); err != nil {
		return Compose(""), errors.Join(errors.New("failed to create journey base directory"), err)
	}
	if err := os.WriteFile(journeyBaseDir+"/Dockerfile", journeyBaseImage, 0644); err != nil {
		return Compose(""), errors.Join(errors.New("failed to write journey base Dockerfile"), err)
	}

	for _, journey := range c.Journeys {
		journeyDir := filepath.Join(dir, "_journeys", sanitizeName(journey.Name))
		if err := os.MkdirAll(journeyDir, 0755); err != nil {
			return Compose(""), errors.Join(errors.New("failed to create journey directory"), err)
		}
		dockerfile := strings.Builder{}
		dockerfile.WriteString("FROM khh_journey_base\n")
		dockerfile.WriteString("RUN mkdir /journey\n")
		dockerfile.WriteString("WORKDIR /journey\n")
		dockerfile.WriteString("COPY . /journey\n")
		dockerfile.WriteString("RUN chmod +x /journey/entrypoint.sh\n")
		dockerfile.WriteString("WORKDIR /journey\n")
		dockerfile.WriteString("ENTRYPOINT [\"/journey/entrypoint.sh\"]\n")
		if err := os.WriteFile(filepath.Join(journeyDir, "Dockerfile"), []byte(dockerfile.String()), 0644); err != nil {
			return Compose(""), errors.Join(errors.New("failed to write journey Dockerfile"), err)
		}

		entrypoint := strings.Builder{}
		entrypoint.WriteString("#!/bin/sh\n")
		entrypoint.WriteString("set -e # exit on error\n")
		for i, step := range journey.Steps {
			entrypoint.WriteString(fmt.Sprintf("# step %d\n", i+1))
			entrypoint.WriteString(step)
			entrypoint.WriteString("\n")
		}
		if err := os.WriteFile(filepath.Join(journeyDir, "entrypoint.sh"), []byte(entrypoint.String()), 0755); err != nil {
			return Compose(""), errors.Join(errors.New("failed to write journey entrypoint"), err)
		}
	}

	dockerCompose := strings.Builder{}
	dockerCompose.WriteString("services:\n")
	for _, service := range c.Services {
		dockerCompose.WriteString(indent)
		dockerCompose.WriteString(sanitizeName(service.Name))
		dockerCompose.WriteString(":\n")
		dockerCompose.WriteString(indent2)
		dockerCompose.WriteString("image: ")
		dockerCompose.WriteString(service.Image)
		dockerCompose.WriteString("\n")
		dockerCompose.WriteString(indent2)
		dockerCompose.WriteString("networks:\n")
		dockerCompose.WriteString(indent3)
		dockerCompose.WriteString(fmt.Sprintf("- khh_%d\n", ri))
	}
	for _, journey := range c.Journeys {
		dockerCompose.WriteString(indent)
		dockerCompose.WriteString("journey__")
		dockerCompose.WriteString(sanitizeName(journey.Name))
		dockerCompose.WriteString(":\n")
		dockerCompose.WriteString(indent2)
		dockerCompose.WriteString("build: ./_journeys/")
		dockerCompose.WriteString(sanitizeName(journey.Name))
		dockerCompose.WriteString("\n")
		dockerCompose.WriteString(indent2)
		dockerCompose.WriteString("networks:\n")
		dockerCompose.WriteString(indent3)
		dockerCompose.WriteString(fmt.Sprintf("- khh_%d\n", ri))
	}
	dockerCompose.WriteString("networks:\n")
	dockerCompose.WriteString(indent)
	dockerCompose.WriteString(fmt.Sprintf("khh_%d:\n", ri))
	if err := os.WriteFile(dir+"/docker-compose.yml", []byte(dockerCompose.String()), 0644); err != nil {
		return Compose(""), errors.Join(errors.New("failed to write docker-compose.yml"), err)
	}

	return Compose(dir), nil
}

var sanitizeRegexp = regexp.MustCompile("[^a-zA-Z0-9_]+")

func sanitizeName(name string) string {
	return sanitizeRegexp.ReplaceAllString(name, "_")
}

func (c Compose) Up() error {
	// build base image for journeys
	cmd := exec.Command("docker", "build", "-t", "khh_journey_base", "_journeys/_base")
	cmd.Dir = string(c)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return errors.Join(errors.New("failed to build journey base image"), err)
	}

	// run compose up
	cmd = exec.Command("docker", "compose", "up", "-d", "--build")
	cmd.Dir = string(c)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return errors.Join(errors.New("failed to run docker compose up"), err)
	}
	return nil
}

func (c Compose) Down() error {
	if _, err := os.Stat(string(c)); os.IsNotExist(err) {
		return nil
	}

	cmd := exec.Command("docker", "compose", "down", "--volumes", "-rmi", "all")
	cmd.Dir = string(c)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(string(c))
		return errors.Join(errors.New("failed to run docker compose down"), err)
	}
	if err := os.RemoveAll(string(c)); err != nil {
		return errors.Join(errors.New("failed to remove docker compose directory"), err)
	}
	return nil
}
