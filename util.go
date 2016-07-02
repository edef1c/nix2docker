package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

func toReader(f func(io.Writer) error) io.Reader {
	r, w := io.Pipe()
	go func() {
		var err error
		defer w.CloseWithError(err)
		f(w)
	}()
	return r
}

func writeJSONFile(path string, value interface{}, mode os.FileMode) error {
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, buf, mode)
}

func readJSONFile(path string, value interface{}) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(buf, value); err != nil {
		return &os.PathError{
			Op:   "unmarshal",
			Path: path,
			Err:  err,
		}
	}
	return nil
}
