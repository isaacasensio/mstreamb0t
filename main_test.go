package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const binaryName = "mstreamb0t"

func TestMain(m *testing.M) {
	make := exec.Command("make", "build")
	err := make.Run()
	if err != nil {
		fmt.Printf("could not make binary for %s: %v", binaryName, err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestCliArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"no arguments", []string{}, "manga name cannot be empty"},
		{"empty manga-name", []string{"-manga-names="}, "manga name cannot be empty"},
		{"not numeric interval", []string{"-interval=g"}, "invalid value"},
		{"unknown flag", []string{"-something=g"}, "flag provided but not defined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command(path.Join(dir, binaryName), tt.args...)
			output, err := cmd.CombinedOutput()
			assert.Error(t, err)

			actual := string(output)
			assert.Contains(t, actual, tt.fixture)
		})
	}
}
