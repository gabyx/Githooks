package common

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

// LoadJSON loads and parses JSON file into a representation.
func LoadJSON(file string, repr interface{}) error {
	jsonFile, err := os.Open(file)
	if err != nil {
		return ErrorF("Could not open file '%s'.", file)
	}
	defer jsonFile.Close()

	err = ReadJSON(jsonFile, repr)
	if err != nil {
		return CombineErrors(err,
			ErrorF("Could not read JSON from file '%s'.", file))
	}

	return nil
}

// StoreJSON stores a representation in a JSON file.
func StoreJSON(file string, repr interface{}) error {
	jsonFile, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	err = WriteJSON(jsonFile, repr)
	if err != nil {
		return CombineErrors(err,
			ErrorF("Could not write JSON to file '%s'.", file))
	}

	return nil
}

// WriteJSON writes the JSON representation of `repr` to `writer`.
func WriteJSON(writer io.Writer, repr interface{}) error {
	bytes, err := json.Marshal(repr)
	if err != nil {
		return err
	}

	_, err = writer.Write(bytes)

	return err
}

// ReadJSON reads the JSON representation of `repr` from `reader`.
func ReadJSON(reader io.Reader, repr interface{}) error {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, repr)
}
