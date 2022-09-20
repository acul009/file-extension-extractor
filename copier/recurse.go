package copier

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func StartCopy(from string, to string, extensions []string, blacklist bool, goroutines int) {

	files := make(chan string)
	go recurse(from, files)
	filteredFiles := make(chan string)
	go filterFiles(files, filteredFiles, extensions, blacklist)

	for file := range filteredFiles {
		fmt.Println(file)
	}
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
