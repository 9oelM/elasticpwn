package EPUtils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ReadUrlsFromFile(path_to_url_text_file string) string {
	EPLogger(fmt.Sprintf("Reading URLs from %s\n", path_to_url_text_file))
	b := ReadFile(path_to_url_text_file)
	urlsArr := strings.Split(strings.TrimSpace(string(b)), "\n")
	// @todo trim whitespaces and empty lines
	for index, url := range urlsArr {
		urlsArr[index] = strings.TrimSuffix(url, "/")
	}
	return strings.Join(urlsArr, "\n")
}

// reads a file that's structured like
// {"something": 1 },
// {"something": 2 },
// {"something": 3 },
// and turns it into
// [{"something": 1 },
// {"something": 2 },
// {"something": 3 }]
func ConvertJSONObjectsToJSONArray(pathToJsonFile string) {
	output, err := ioutil.ReadFile(pathToJsonFile)
	ExitOnError(err)

	// this is because we can't take care of the last comma when we finish writing json
	withoutLastTrailingComma := strings.TrimSuffix(string(output), ",\n")
	// right square bracket of the JSON array
	output = append([]byte(withoutLastTrailingComma), "]"...)

	tmpFile := fmt.Sprintf("%s.tmp", pathToJsonFile)
	f, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	defer func() {
		err := f.Close()
		if err != nil {
			EPLogger(fmt.Sprintf("Failed to close %v while finalizing\n", pathToJsonFile))
			return
		}
	}()

	if err != nil {
		EPLogger(fmt.Sprintf("Failed to open %v while finalizing\n", pathToJsonFile))
		return
	}

	// left square bracket of the JSON array
	_, err = f.WriteString("[")
	ExitOnError(err)

	_, err = f.Write(output)
	ExitOnError(err)

	err = os.Remove(pathToJsonFile)
	ExitOnError(err)

	err = os.Rename(tmpFile, pathToJsonFile)
	ExitOnError(err)
}

func ReadFile(path string) []byte {
	file, err := os.Open(filepath.FromSlash(path))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

func OverwriteFile(path string, content string) {
	err := os.Remove(filepath.FromSlash(path))
	if !errors.Is(err, os.ErrNotExist) && err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(filepath.FromSlash(path))
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString(content)
}
