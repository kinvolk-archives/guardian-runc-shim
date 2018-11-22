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

		in, err := ioutil.ReadFile(filepath.Join(TEST_DATA_DIR, d.Name(), "in.json"))
		if err != nil {
			t.Fatalf("Reading test data: %v", err)
		}

		out, err := ioutil.ReadFile(filepath.Join(TEST_DATA_DIR, d.Name(), "out.json"))
		if err != nil {
			t.Fatalf("Reading test data: %v", err)
		}

		res, err := modifyConfig(in)
		if err != nil {
			t.Fatal(err)
		}

		// Unmarshal JSON into structs so that we can check the result while ignoring whitespace
		// and ordering differences.
		var resStruct, outStruct map[string]interface{}

		err = json.Unmarshal(res, &resStruct)
		if err != nil {
			t.Fatal(err)
		}

		err = json.Unmarshal(out, &outStruct)
		if err != nil {
			t.Fatal(err)
		}

		// Check result.
		if !reflect.DeepEqual(resStruct, outStruct) {
			t.Fatalf("Invalid config returned after modification: got \n%v, want \n%v",
				resStruct, outStruct)
		}
	}
}
