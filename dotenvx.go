// Package dotenvx loads plaintext and encrypted dotenv files using the shared
// dotenvx SDK.
package dotenvx

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

type source struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type loadOptions struct {
	Cwd               string            `json:"cwd"`
	Sources           []source          `json:"sources"`
	ProcessEnv        map[string]string `json:"process_env"`
	Overload          bool              `json:"overload"`
	Strict            bool              `json:"strict"`
	ResolveReferences bool              `json:"resolve_references"`
}

type diagnostic struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

type loadResult struct {
	Values      map[string]string `json:"values"`
	Injected    map[string]string `json:"injected"`
	Diagnostics []diagnostic      `json:"diagnostics"`
	Summary     string            `json:"summary"`
	Error       string            `json:"error"`
}

// Load loads dotenv files into the process environment. It defaults to .env.
// Existing environment variables and earlier files take precedence.
func Load(files ...string) error {
	return inject(files, false)
}

// Overload loads dotenv files into the process environment, replacing existing
// values. Later files take precedence.
func Overload(files ...string) error {
	return inject(files, true)
}

// Read returns the values from dotenv files without changing the process
// environment. Later files take precedence.
func Read(files ...string) (map[string]string, error) {
	result, err := load(files, true)
	if err != nil {
		return nil, err
	}
	return result.Values, nil
}

func inject(files []string, overload bool) error {
	result, err := load(files, overload)
	if err != nil {
		return err
	}

	for key, value := range result.Injected {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("dotenvx: setting %q: %w", key, err)
		}
	}
	if result.Summary != "" {
		fmt.Fprintln(os.Stderr, result.Summary)
	}
	return nil
}

func load(files []string, overload bool) (loadResult, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return loadResult{}, err
	}
	if len(files) == 0 {
		files = []string{".env"}
	}
	sources := make([]source, 0, len(files))
	for _, file := range files {
		sources = append(sources, source{Type: "file", Value: file})
	}
	env := make(map[string]string)
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			env[key] = value
		}
	}

	result, err := loadNative(loadOptions{
		Cwd:               cwd,
		Sources:           sources,
		ProcessEnv:        env,
		Overload:          overload,
		Strict:            true,
		ResolveReferences: true,
	})
	if err != nil {
		return result, err
	}
	if len(result.Diagnostics) > 0 {
		diagnostic := result.Diagnostics[0]
		switch diagnostic.Code {
		case "MISSING_ENV_FILE":
			return result, &fs.PathError{Op: "open", Path: diagnostic.Path, Err: fs.ErrNotExist}
		case "EACCES":
			return result, &fs.PathError{Op: "open", Path: diagnostic.Path, Err: fs.ErrPermission}
		default:
			return result, errors.New(diagnostic.Message)
		}
	}
	return result, nil
}
