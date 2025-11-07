package handlers

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/evanwiseman/ionbus/internal/models"
)

type Executor struct{}

func (e *Executor) Execute(handler *models.Handler, input []byte) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "*"+handler.Extension)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(handler.Program)
	if err != nil {
		return nil, err
	}
	tmpFile.Close()

	// Replace {file} placeholder in RunArgs with actual path
	cmdStr := strings.Replace(handler.RunCommand, "{file}", tmpFile.Name(), 1)
	tokens := tokenize(cleanText(cmdStr))
	name := tokens[0]
	args := tokens[1:]

	cmd := exec.Command(name, args...)
	cmd.Stdin = bytes.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func cleanText(text string) string {
	return strings.TrimSpace(text)
}

func tokenize(text string) []string {
	return strings.Split(text, " ")
}
