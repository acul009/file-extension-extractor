package copier

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const CHANNEL_BUFFER_SIZE = 10

func StartCopy(from string, to string, extensions []string, blacklist bool, goroutines int, bufferSize int) {

	files := make(chan string, CHANNEL_BUFFER_SIZE)
	go recurse(from, files)
	filteredFiles := make(chan string, CHANNEL_BUFFER_SIZE)
	go filterFiles(files, filteredFiles, extensions, blacklist)

	if goroutines > 1 {
		for i := 1; i < goroutines; i++ {
			go copyFilesWithDirStructure(from, to, filteredFiles, bufferSize)
		}
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

		copyFilesWithDirStructure(from, to, fileChannel, bufferSize)

	}
}

func copyFilesWithDirStructure(from string, to string, files chan string, bufferSize int) {
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
		copyWithBuffer(file, targetPath, buffer)
	}
}

func copyWithBuffer(from string, to string, buffer []byte) {
	source, err := os.Open(from)
	if err != nil {
		panic(err)
	}
	defer source.Close()

	destination, err := os.Create(to)
	if err != nil {
		panic(err)
	}
	defer destination.Close()

	io.CopyBuffer(destination, source, buffer)
}

func filterFiles(in chan string, out chan string, extensions []string, blacklist bool) {

	suffixList := make([]string, 0, len(extensions))
	for _, extension := range extensions {
		suffixList = append(suffixList, "."+extension)
	}

	for file := range in {
		for _, suffix := range suffixList {
			if strings.HasSuffix(file, suffix) != blacklist {
				out <- file
			}
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
