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

func SearchString(fSys fs.FS, path string, stdin io.Reader, keyword string, ignoreCase bool) ([]string, error) {
	var scanner *bufio.Scanner

	if path != "" {
		// check for validity
		err := isValid(fSys, path)
		if err != nil {
			return nil, err
		}

		// open file
		file, err := fSys.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	} else {
		scanner = bufio.NewScanner(stdin)
	}

	if ignoreCase {
		keyword = strings.ToLower(keyword)
	}

	
	var result []string
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

	// fmt.Println(result)
	return result, nil	
}

func isValid(fSys fs.FS, path string) error {
	fileInfo, err := fs.Stat(fSys, path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", path, fs.ErrNotExist)
		}
		return fmt.Errorf("%s: %w", path, err)
	}

	// checks for directory
	if fileInfo.IsDir() {
		return fmt.Errorf("%s: %w", path, ErrIsDirectory)
	}

	// checks for permissions
	// looks hacky, might have to change later
	if fileInfo.Mode().Perm()&400 == 0 {
		return fmt.Errorf("%s: %w", path, fs.ErrPermission)
	}

	return nil
}