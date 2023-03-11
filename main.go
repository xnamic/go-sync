package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	var s, d string
	var w int

	flag.StringVar(&s, "s", "", "fullpath source")
	flag.StringVar(&d, "d", "", "fullpath destination")
	flag.IntVar(&w, "w", 10, "worker number")
	flag.Parse()

	if s == "" || d == "" {
		fmt.Println("source folder or destination folder can't be empty, type './go-sync --help' for more`")
		return
	}

	if w < 1 {
		fmt.Println("worker must be > 0")
		return
	}

	done := syncFolder(s, d, w)
	if !done {
		fmt.Println("sync completed but there are some errors, check error logs")
		return
	}
}

func removeEmptyDir(dir string) error {
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		empty, err := IsEmptyDir(path, info)
		if err != nil {
			fmt.Printf("error reading folder: %v\n", err)
			return err
		}

		if empty {
			os.RemoveAll(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", dir, err)
		return err
	}
	return nil
}

func IsEmptyDir(name string, fi fs.FileInfo) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	if fi.IsDir() {
		_, err = f.Readdirnames(1)
		if err == io.EOF {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
