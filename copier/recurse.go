package copier

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const CHANNEL_BUFFER_SIZE = 10

func StartCopy(from string, to string, extensions []string, blacklist bool, goroutines int, bufferSize int, move bool) {

	files := make(chan string, CHANNEL_BUFFER_SIZE)
	go recurse(from, files)
	filteredFiles := make(chan string, CHANNEL_BUFFER_SIZE)
	go filterFiles(files, filteredFiles, extensions, blacklist)

	if goroutines > 1 {

		var wg sync.WaitGroup = sync.WaitGroup{}
		wg.Add(goroutines - 1)
		for i := 1; i < goroutines; i++ {
			go func() {
				copyFilesWithDirStructure(from, to, filteredFiles, bufferSize, move)
				wg.Done()
			}()
		}
		wg.Wait()
	} else {
		fileList := make([]string, 0)
		for file := range filteredFiles {
			fileList = append(fileList, file)
		}
		fileChannel := make(chan string, CHANNEL_BUFFER_SIZE)

		go func() {
			for _, file := range fileList {
				fileChannel <- file
			}
			close(fileChannel)
		}()

		copyFilesWithDirStructure(from, to, fileChannel, bufferSize, move)

	}
}

func copyFilesWithDirStructure(from string, to string, files chan string, bufferSize int, move bool) {
	to = strings.TrimRight(to, "/")

	buffer := make([]byte, bufferSize*1024)

	for file := range files {
		relativeLocation := strings.Trim(file[len(from):], "/")
		targetPath := to + "/" + relativeLocation
		dir := filepath.Dir(targetPath)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		if move {
			moveFile(file, targetPath, buffer)
		} else {
			copyWithBuffer(file, targetPath, buffer)
		}
	}
}

var copyRequired = false

func moveFile(from string, to string, buffer []byte) {
	if !copyRequired {
		err := os.Rename(filepath.FromSlash(from), filepath.FromSlash(to))
		if err != nil {
			copyRequired = true
			moveFile(from, to, buffer)
		}
	} else {
		err := copyWithBuffer(from, to, buffer)
		if err != nil {
			panic(err)
		}
		os.Remove(filepath.FromSlash(from))
	}
}

func copyWithBuffer(from string, to string, buffer []byte) error {
	source, err := os.Open(filepath.FromSlash(from))
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(filepath.FromSlash(to))
	if err != nil {
		return err
	}
	defer destination.Close()

	io.CopyBuffer(destination, source, buffer)
	return err
}

func filterFiles(in chan string, out chan string, extensions []string, blacklist bool) {

	suffixList := make([]string, 0, len(extensions))
	for _, extension := range extensions {
		suffixList = append(suffixList, "."+extension)
	}

file:
	for file := range in {
		for _, suffix := range suffixList {
			if strings.HasSuffix(file, suffix) {
				if !blacklist {
					out <- file
				}
				continue file
			}
		}
		if blacklist {
			out <- file
		}
	}
	close(out)
}

func recurse(from string, files chan string) error {
	queue := []string{from}
	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		fileinfo, err := os.Stat(path)
		if err != nil {
			return err
		}
		if fileinfo.IsDir() {
			fileInfo, err := ioutil.ReadDir(path)
			if err != nil {
				return err
			}
			for _, file := range fileInfo {
				filepath := path + "/" + file.Name()
				queue = append(queue, filepath)
			}
		} else {
			files <- path
		}
	}
	close(files)
	return nil
}
