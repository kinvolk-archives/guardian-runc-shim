package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"testing"
)

const TEST_DATA_DIR = "testdata"

func TestModifyConfig(t *testing.T) {
	// Read test data.
	dirs, err := ioutil.ReadDir(TEST_DATA_DIR)
	if err != nil {
		log.Fatalf("Listing test data directories: %v", err)
	}

	for _, d := range dirs {
		// Skip files.
		if !d.IsDir() {
			continue
		}

		inJSON, err := ioutil.ReadFile(filepath.Join(TEST_DATA_DIR, d.Name(), "in.json"))
		if err != nil {
			t.Fatalf("Reading test data: %v", err)
		}

		outJSON, err := ioutil.ReadFile(filepath.Join(TEST_DATA_DIR, d.Name(), "out.json"))
		if err != nil {
			t.Fatalf("Reading test data: %v", err)
		}

		var in, out map[string]interface{}

		err = json.Unmarshal(inJSON, &in)
		if err != nil {
			t.Fatal("Error parsing JSON")
		}

		err = json.Unmarshal(outJSON, &out)
		if err != nil {
			t.Fatal("Error parsing JSON")
		}

		res, err := modifyConfig(in)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(res, out) {
			t.Fatalf("Invalid config returned after modification: got %v, want %v", res, out)
		}
	}
}
