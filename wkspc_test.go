// Copyright © 2015 Erik Brady <brady@dvln.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package test: wkspc
//      Basic testing for the 'wkspc' package

package wkspc

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/dvln/util/dir"
)

func TestWkspcRootDetermination(t *testing.T) {
	tempFolder, err := ioutil.TempDir("", "dvln-wkspc-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempFolder)

	fakeWkspcSubDir := filepath.Join(tempFolder, "fake_wkspc_subdir")

	if err := dir.CreateIfNotExists(fakeWkspcSubDir); err != nil {
		t.Fatal(err)
	}
	fileinfo, err := os.Stat(fakeWkspcSubDir)
	if err != nil {
		t.Fatalf("Should have create a folder, got %v", err)
	}
	if !fileinfo.IsDir() {
		t.Fatalf("Should have been a dir, seems it's not")
	}
	exists, err := dir.Exists(fakeWkspcSubDir)
	if err != nil {
		t.Fatal("Folder should have existed but instead an error was returned")
	}
	if !exists {
		t.Fatal("Folder should have existed but Exists() said it did not")
	}

	wsRootFolder := filepath.Join(tempFolder, ".dvln")
	if err := dir.CreateIfNotExists(wsRootFolder); err != nil {
		t.Fatal(err)
	}
	fileinfo, err = os.Stat(wsRootFolder)
	if err != nil {
		t.Fatalf("Should have create a .dvln folder, got %v", err)
	}
	if !fileinfo.IsDir() {
		t.Fatalf("Should have been a dir, seems .dvln is not")
	}
	exists, err = dir.Exists(wsRootFolder)
	if err != nil {
		t.Fatal("Folder .dvln should have existed but instead an error was returned")
	}
	if !exists {
		t.Fatal("Folder .dvln should have existed but Exists() said it did not")
	}

	root, err := Root(fakeWkspcSubDir)
	if err != nil {
		t.Fatalf("Search for workspace root from %s should not have returned an error, got %v", fakeWkspcSubDir, err)
	}
	if root == "" {
		t.Fatal("Failed to find the workspace root folder, it should have been found")
	}
	if root != tempFolder {
		t.Fatalf("Found a workspace root but didn't match expected (found: %s, expected: %s)", root, tempFolder)
	}

	privTempFolder := filepath.Join("/", "private", tempFolder)
	root, err = Root(tempFolder)
	if err != nil {
		t.Fatalf("Search for a workspace root dir (from the workspace root dir) got unexpected error: %v", err)
	}
	if root == "" {
		t.Fatal("Failed to find the workspace root directory (from that dir), it should have been found")
	}
	if root != tempFolder && root != privTempFolder {
		t.Fatalf("Root found to be %s, should have been set to %s", root, tempFolder)
	}

	os.Chdir(fakeWkspcSubDir)
	root, err = Root()
	if err != nil {
		t.Fatalf("Search for workspace root after Chdir into wkspc subdir should not have returned an error, got %v", err)
	}
	if root == "" {
		t.Fatal("After chdir, failed to find the workspace root folder, it should have been found")
	}
	if root != tempFolder && root != privTempFolder {
		t.Fatalf("After chdir, found a workspace root but didn't match expected (found: %s, expected: %s)", root, tempFolder)
	}
}
