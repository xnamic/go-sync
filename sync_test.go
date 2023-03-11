package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMergeMap(t *testing.T) {
	tests := []struct {
		a        map[string]File
		b        map[string]File
		expected map[string]File
	}{
		{
			a:        map[string]File{"a": {Path: "patha", Size: 1000}, "c": {Path: "pathc", Size: 1000}},
			b:        map[string]File{"e": {Path: "pathe", Size: 1000}, "g": {Path: "pathg", Size: 1000}},
			expected: map[string]File{"a": {Path: "patha", Size: 1000}, "c": {Path: "pathc", Size: 1000}, "e": {Path: "pathe", Size: 1000}, "g": {Path: "pathg", Size: 1000}},
		},
		{
			a:        map[string]File{},
			b:        map[string]File{},
			expected: map[string]File{},
		},
	}

	for i, test := range tests {
		mergeMap(test.a, test.b)
		if !reflect.DeepEqual(test.a, test.expected) {
			t.Errorf("fest %d: expected %v but got %v", i+1, test.expected, test.a)
		}
	}
}

func TestScanDir(t *testing.T) {
	dir := "/tmp/testdir/scandir"
	os.MkdirAll(dir, os.ModePerm)
	defer os.RemoveAll(dir)
	f, err := os.Create(filepath.Join(dir, "file1.txt"))
	if err != nil {
		t.Errorf("failed to create file: %v", err)
	}
	f.Close()

	files, err := scanDir(dir)
	if err != nil {
		t.Errorf("failed to scan directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}

	dir = "/tmp/nonexistentdir/scandir"
	_, err = scanDir(dir)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestFileName(t *testing.T) {
	s := "/tmp/testdir/filename"
	d := "/tmp/testdirsync/filename"

	filename1 := "file1.txt"
	filename2 := "file2.txt"

	os.MkdirAll(s, os.ModePerm)
	os.MkdirAll(d, os.ModePerm)
	defer os.RemoveAll(s)
	defer os.RemoveAll(d)

	f1, err := os.Create(filepath.Join(s, filename1))
	if err != nil {
		t.Errorf("failed to create file: %v", err)
	}
	f2, err := os.Create(filepath.Join(s, filename2))
	if err != nil {
		t.Errorf("failed to create file: %v", err)
	}
	f1.Close()
	f2.Close()

	f2copy, f2delete, err := fileName(s, d)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	f2copy_expected := map[string]string{
		"/tmp/testdir/filename/file1.txt": "/tmp/testdirsync/filename/file1.txt",
		"/tmp/testdir/filename/file2.txt": "/tmp/testdirsync/filename/file2.txt",
	}
	f2delete_expected := map[string]string{}

	if !reflect.DeepEqual(f2copy, f2copy_expected) || !reflect.DeepEqual(f2delete, f2delete_expected) {
		t.Errorf("expected %v and %v but got %v and %v", f2copy_expected, f2delete_expected, f2copy, f2delete)
	}
}

func TestDeleteFile(t *testing.T) {
	s := "/tmp/testdir/deletefile"
	filename := "file1.txt"
	os.MkdirAll(s, os.ModePerm)
	defer os.RemoveAll(s)

	f, err := os.Create(filepath.Join(s, filename))
	if err != nil {
		t.Errorf("failed to create file: %v", err)
	}
	f.Close()

	file := SyncFile{
		Source: f.Name(),
	}

	err = deleteFile(file)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	s := "/tmp/testdir/copyfile"
	d := "/tmp/testdirsync/copyfile"
	filename := "file1.txt"
	os.MkdirAll(s, os.ModePerm)
	os.MkdirAll(d, os.ModePerm)
	defer os.RemoveAll(s)
	defer os.RemoveAll(d)

	f, err := os.Create(filepath.Join(s, filename))
	if err != nil {
		t.Errorf("failed to create file: %v", err)
	}
	f.Close()

	file := SyncFile{
		Source:      f.Name(),
		Destination: filepath.Join(d, filename),
	}

	err = copyFile(file)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDispatcher(t *testing.T) {
	got := []SyncFile{}
	expected := []SyncFile{
		{
			Source:      "/tmp/testdir/file1.txt",
			Destination: "/tmp/testdirsync/file1.txt",
		}}

	files := map[string]string{
		"/tmp/testdir/file1.txt": "/tmp/testdirsync/file1.txt",
	}
	res := dispatcher(files)

	for f := range res {
		got = append(got, f)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected = %v but got %v", expected, got)
	}
}

func TestSyncFolder(t *testing.T) {
	s := "/tmp/testdir/syncfolder"
	d := "/tmp/testdirsync/syncfolder"
	file1 := "file1.txt"
	file2 := "file1.txt"
	os.MkdirAll(s, os.ModePerm)
	os.MkdirAll(d, os.ModePerm)
	defer os.RemoveAll(s)
	defer os.RemoveAll(d)

	f1, err := os.Create(filepath.Join(s, file1))
	if err != nil {
		t.Errorf("failed to create file: %v", err)
	}
	f2, err := os.Create(filepath.Join(s, file2))
	if err != nil {
		t.Errorf("failed to create file: %v", err)
	}
	f1.Close()
	f2.Close()

	done := syncFolder(s, d, 2)
	if !done {
		t.Error("there is error while copying or deleting file")
	}
}
