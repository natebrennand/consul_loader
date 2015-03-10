// +build integration

package main

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"
)

var (
	testTree = tree(map[string]interface{}{
		"key1": float64(1),
		"key2": float64(2),
		"subtree": map[string]interface{}{
			"key3": float64(3),
		},
	})
	expectedJSON = `{"key1":1,"key2":2,"subtree":{"key3":3}}`
)

func randFile() string {
	return path.Join(os.TempDir(), strconv.Itoa(rand.Intn(1000)))
}

func TestWriteJSONFile(t *testing.T) {
	tmpFile := randFile()
	defer os.Remove(tmpFile)
	writeJSONFile(testTree, tmpFile)

	contents, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	if expectedJSON != string(contents) {
		t.Errorf("Expected: %s\nRecieved: %s", expectedJSON, string(contents))
	}
}

// diffTree compares two nested trees
func diffTree(a, b tree, t *testing.T) {
	for k, v := range a {
		if subtree, isTree := v.(map[string]interface{}); isTree {
			val, valid := b[k]
			if !valid {
				t.Fatalf("Read tree missing key, %s", k)
			}

			subtreeB, valid := val.(map[string]interface{})
			if !valid {
				t.Fatalf("Read tree missing subtree, %s", k)
			}

			diffTree(subtree, subtreeB, t)
		} else {
			_, valid := b[k]
			if !valid {
				t.Fatalf("Read tree missing key, %s", k)
			} else if a[k] != b[k] {
				t.Fatalf("Expected: %s\nRecieved: %s", a[k], b[k])
			}
		}
	}
}

func TestReadJSONFile(t *testing.T) {
	tmpFile := randFile()
	defer os.Remove(tmpFile)
	err := ioutil.WriteFile(tmpFile, []byte(expectedJSON), os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to write temporary file => {%s}", err)
	}

	loadedTree := readJSONFile(tmpFile)
	diffTree(testTree, loadedTree, t)
}
