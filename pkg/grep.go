package grep

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
)

var (
	ErrIsDirectory = errors.New("is a directory")
)

func GrepRun(fSys fs.FS, path string, stdin io.Reader, keyword string, ignoreCase bool) ([]string, error) {
	r, cleanup, err := getReader(fSys, path, stdin)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	result, err := searchString(r, keyword, ignoreCase)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func getReader(fSys fs.FS, fileName string, stdin io.Reader) (io.Reader, func(), error) {
	if fileName != "" {
		err := isValid(fSys, fileName)
		if err != nil {
			return nil, nil, err
		}

		file, err := fSys.Open(fileName)
		if err != nil {
			return nil, nil, err
		}
		return file, func() {file.Close()}, nil
	}

	return stdin, func() {}, nil
}

func searchString(r io.Reader,  keyword string, ignoreCase bool) ([]string, error) {
	var result []string
	scanner := bufio.NewScanner(r)

	if ignoreCase {
		keyword = strings.ToLower(keyword)
	}

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var line string
		line = scanner.Text()

		if ignoreCase {
			line = strings.ToLower(scanner.Text())
		}

		if strings.Contains(line, keyword) {
			result = append(result, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func isValid(fSys fs.FS, fileName string) error {
	fileInfo, err := fs.Stat(fSys, fileName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", fileName, fs.ErrNotExist)
		}
		return fmt.Errorf("%s: %w", fileName, err)
	}

	// checks for directory
	if fileInfo.IsDir() {
		return fmt.Errorf("%s: %w", fileName, ErrIsDirectory)
	}

	// checks for permissions
	// looks hacky, might have to change later
	if fileInfo.Mode().Perm()&400 == 0 {
		return fmt.Errorf("%s: %w", fileName, fs.ErrPermission)
	}

	return nil
}

// func GrepRun(fSys fs.FS, fileName string, stdin io.Reader, keyword string, ignoreCase bool) ([]string, error) {
// 	var scanner *bufio.Scanner

// 	if fileName != "" {
// 		// check for validity
// 		err := isValid(fSys, fileName)
// 		if err != nil {
// 			return nil, err
// 		}

// 		// open file
// 		file, err := fSys.Open(fileName)
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer file.Close()

// 		scanner = bufio.NewScanner(file)
// 	} else {
// 		scanner = bufio.NewScanner(stdin)
// 	}

// 	if ignoreCase {
// 		keyword = strings.ToLower(keyword)
// 	}

// 	var result []string
// 	scanner.Split(bufio.ScanLines)
// 	for scanner.Scan() {
// 		var line string
// 		line = scanner.Text()

// 		if ignoreCase {
// 			line = strings.ToLower(scanner.Text())
// 		}

// 		if strings.Contains(line, keyword) {
// 			result = append(result, scanner.Text())
// 		}
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return nil, err
// 	}

// 	// fmt.Println(result)
// 	return result, nil
// }