package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const BUFFSIZE = 128000

type SyncFile struct {
	Source      string
	Destination string
}

type File struct {
	Path string
	Size int64
}

type Process func(f SyncFile) error

// syncFolder will sync files between two folders
// run copy files and delete files asynchronously
// will return true if there is error while process each file
// will return false if there is one or more failed process
func syncFolder(s, d string, nWorkers int) bool {
	counterDelete := 0
	counterCopy := 0
	source := filepath.Clean(s)
	destination := filepath.Clean(d)

	err := os.MkdirAll(destination, os.ModePerm)
	if err != nil {
		log.Println(err)
	}

	f2copy, f2delete, err := fileName(source, destination)
	if err != nil {
		log.Println(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		chFiles := dispatcher(f2delete)
		deleteErrs := process(chFiles, deleteFile, nWorkers)
		for err := range deleteErrs {
			if err != nil {
				log.Println(err)
			} else {
				counterDelete++
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		chFiles := dispatcher(f2copy)
		copyErrs := process(chFiles, copyFile, nWorkers)
		for err := range copyErrs {
			if err != nil {
				log.Println(err)
			} else {
				counterCopy++
			}
		}
	}()
	wg.Wait()

	removeEmptyDir(d)

	return len(f2copy)+len(f2delete) == counterCopy+counterDelete
}

// process recive file to sync from dispatcher() function
// channel out will send result of process, copyFile() function and deleteFile() function
func process(chanIn <-chan SyncFile, p Process, nWorkers int) <-chan error {
	chanOut := make(chan error, nWorkers)
	wg := new(sync.WaitGroup)
	wg.Add(nWorkers)

	go func() {
		for workerIndex := 0; workerIndex < nWorkers; workerIndex++ {
			go func() {
				for file := range chanIn {
					err := p(file)
					chanOut <- err
				}
				wg.Done()
			}()
		}
	}()

	go func() {
		wg.Wait()
		close(chanOut)
	}()

	return chanOut
}

// dispatcher is to dispatch all files that will be synced
// IMPORTANT: paramater files map[string]string
// for copy file index of the map is source path and value of the map is destination path
// for delete file index of the map is file path to delete, and the value is empty string
func dispatcher(files map[string]string) <-chan SyncFile {
	chanOut := make(chan SyncFile)
	go func() {
		for s, d := range files {
			chanOut <- SyncFile{
				Source:      s,
				Destination: d,
			}
		}
		close(chanOut)
	}()
	return chanOut
}

// copyFile is used to copy file using io.CopyBuffer()
func copyFile(f SyncFile) error {
	source, err := os.Open(f.Source)
	if err != nil {
		return err
	}
	defer source.Close()

	if _, err := os.Stat(f.Source); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	dirName := filepath.Dir(f.Destination)
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.MkdirAll(dirName, os.ModePerm)
		if err != nil {
			return err
		}
	}

	destination, err := os.Create(f.Destination)
	if err != nil {
		return err
	}
	defer destination.Close()

	buf := make([]byte, BUFFSIZE)
	_, err = io.CopyBuffer(destination, source, buf)
	if err != nil {
		return err
	}
	return nil
}

// deleteFile will delete the file
func deleteFile(f SyncFile) error {
	return os.Remove(f.Source)
}

// fileName will return list of file to delete and to copy
func fileName(s string, d string) (map[string]string, map[string]string, error) {
	var errS, errC error

	var wg sync.WaitGroup
	var fSources, fDestinatons map[string]File

	wg.Add(1)
	go func() {
		defer wg.Done()
		fSources, errS = scanDir(s)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fDestinatons, errC = scanDir(d)
	}()
	wg.Wait()

	if errS != nil {
		return nil, nil, errS
	}
	if errC != nil {
		return nil, nil, errC
	}

	f2copy := make(map[string]string)
	f2delete := make(map[string]string)
	for i, e := range fSources {
		dest := strings.Replace(i, s, d, 1)
		if _, exist := fDestinatons[dest]; !exist {
			f2copy[i] = dest
		} else if exist && e.Size != fDestinatons[dest].Size {
			f2copy[i] = dest
		} else {
			delete(fDestinatons, dest)
		}
	}

	for e := range fDestinatons {
		f2delete[e] = ""
	}

	return f2copy, f2delete, nil
}

// scanDir will return map of File
// file path is used as index of map, since it's unique
func scanDir(dir string) (map[string]File, error) {
	sFiles := make(map[string]File)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}

	fns, err := os.ReadDir(dir)
	if err != nil {
		return sFiles, nil
	}

	for _, f := range fns {
		sname := fmt.Sprintf("%s/%s", dir, f.Name())
		stat, err := os.Stat(sname)
		if err != nil {
			return sFiles, err
		}

		// find all files if directory found
		if stat.IsDir() {
			newDir := fmt.Sprintf("%s/%s", dir, f.Name())
			newFiles, _ := scanDir(newDir)
			mergeMap(sFiles, newFiles)
		} else {
			// size is used to compare file in source and destination, incase there are currupted files
			sFiles[sname] = File{
				Path: f.Name(),
				Size: stat.Size(),
			}
		}
	}
	return sFiles, nil
}

// mergeMap is used to merge src map to dst map
func mergeMap(dst map[string]File, src map[string]File) {
	for k, v := range src {
		dst[k] = v
	}
}
