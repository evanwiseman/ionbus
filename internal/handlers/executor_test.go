package handlers

import (
	"testing"

	"github.com/evanwiseman/ionbus/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	executor := Executor{}

	// Valid python program
	handler := &models.Handler{
		ID:         uuid.New(),
		Method:     "hello_world",
		RunCommand: "python3 {file}",
		Extension:  ".py",
		Program:    []byte(`print("Hello, World!")`),
		Version:    "1.0",
	}
	output, err := executor.Execute(handler, nil)
	require.NoError(t, err)
	require.Contains(t, string(output), "Hello, World!")

	// No Args
	handler = &models.Handler{
		ID:         uuid.New(),
		Method:     "no_args",
		RunCommand: "python3",
		Extension:  ".py",
		Program:    []byte(`print("Hello, World!)`),
		Version:    "1.0",
	}
	output, err = executor.Execute(handler, nil)
	require.NoError(t, err)
	require.NotNil(t, output)

	// Additional Args
	handler = &models.Handler{
		ID:         uuid.New(),
		Method:     "additional_args",
		RunCommand: "python3 {file} --verbose --count 3",
		Extension:  ".py",
		Program:    []byte(`import sys; print(sys.argv)`),
	}
	output, err = executor.Execute(handler, nil)
	require.NoError(t, err)
	require.NotNil(t, output)

	// Input via stdin
	handler = &models.Handler{
		ID:         uuid.New(),
		Method:     "stdin_test",
		RunCommand: "python3 {file}",
		Extension:  ".py",
		Program:    []byte(`data = input(); print(f"Received: {data}")`),
	}
	output, err = executor.Execute(handler, []byte("Hello Input"))
	require.NoError(t, err)
	require.Contains(t, string(output), "Received: Hello Input")

	// Empty Program
	handler = &models.Handler{
		ID:         uuid.New(),
		Method:     "empty_program",
		RunCommand: "python3 {file}",
		Extension:  ".py",
		Program:    []byte(""),
	}
	output, err = executor.Execute(handler, nil)
	require.NoError(t, err)
	require.Equal(t, "", string(output))

	// Invalid python program (missing " end)
	handler = &models.Handler{
		ID:         uuid.New(),
		Method:     "invalid_program",
		RunCommand: "python3 {file}",
		Extension:  ".py",
		Program:    []byte(`print("Hello, World!)`),
		Version:    "1.0",
	}
	output, err = executor.Execute(handler, nil)
	require.Error(t, err)
	require.Nil(t, output)
}
