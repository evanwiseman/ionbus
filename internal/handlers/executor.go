package handlers

import (
	"bytes"
	"os"
	"os/exec"

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

	cmd := exec.Command(handler.Language, tmpFile.Name())
	cmd.Stdin = bytes.NewReader(input)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}
