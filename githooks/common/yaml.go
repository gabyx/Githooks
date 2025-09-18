package common

import (
	"io"
	"os"

	"github.com/goccy/go-yaml"
)

// LoadYAML loads and parses JSON file into a representation.
func LoadYAML(file string, repr any) error {
	yamlFile, err := os.Open(file)
	if err != nil {
		return ErrorF("Could not open file '%s'.", file)
	}
	defer func() { _ = yamlFile.Close() }()

	bytes, err := io.ReadAll(yamlFile)
	if err != nil {
		return CombineErrors(err, ErrorF("Could not read file '%s'.", file))
	}

	err = yaml.Unmarshal(bytes, repr)
	if err != nil {
		return CombineErrors(err, ErrorF("Could not unmarshal file '%s'.", file))
	}

	return nil
}

// StoreYAML stores a representation in a JSON file.
func StoreYAML(file string, repr any) error {
	yamlFile, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0664) // nolint: mnd
	if err != nil {
		return err
	}
	defer func() { _ = yamlFile.Close() }()

	bytes, err := yaml.Marshal(repr)
	if err != nil {
		return CombineErrors(err, ErrorF("Could not marshal representation to file '%s'.", file))
	}

	if _, err = yamlFile.Write(bytes); err != nil {
		return CombineErrors(err, ErrorF("Could not write file '%s'.", file))
	}

	return nil
}
