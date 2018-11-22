package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestModifyConfig(t *testing.T) {
	// Read test data.
	origJSON, err := ioutil.ReadFile(filepath.Join("testdata", "orig.json"))
	if err != nil {
		t.Fatalf("Reading test data: %v", err)
	}

	expectedJSON, err := ioutil.ReadFile(filepath.Join("testdata", "expected.json"))
	if err != nil {
		t.Fatalf("Reading test data: %v", err)
	}

	var orig, expected map[string]interface{}

	err = json.Unmarshal(origJSON, &orig)
	if err != nil {
		t.Fatal("Error parsing original JSON")
	}

	err = json.Unmarshal(expectedJSON, &expected)
	if err != nil {
		t.Fatal("Error parsing expected JSON")
	}

	res, err := modifyConfig(orig)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Invalid config returned after modification: got %v, want %v", res, expected)
	}
}
