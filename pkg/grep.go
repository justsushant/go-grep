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

func GrepRun(fSys fs.FS, path string, stdin io.Reader, keyword string, ignoreCase, searchDir bool) ([][]string, error) {
	if !searchDir {
		result, err := Grep(fSys, path, stdin, keyword, ignoreCase)
		if err != nil {
			return nil, err
		}

		return [][]string{result}, nil
	}

	var results [][]string
	// more than one error are not handled correctly here (eg, permisson error in two files)
	fs.WalkDir(fSys, path, func(path string, d fs.DirEntry, err error) error {
		// tests to handle errors inside this func
		// not sure these are bubbling up

		if err != nil {
			// results = append(results, []string{err.Error()})
			return err
		}
		
		if d.IsDir() {
			return nil
		}

		result, errGrep := Grep(fSys, path, stdin, keyword, ignoreCase)
		if errGrep != nil {
			// results = append(results, []string{errGrep.Error()})
			return errGrep
		}

		if len(result) > 0 {
			for i := range result {
				result[i] = fmt.Sprintf("%s:%s", path, result[i])
			}
			results = append(results, result)
		}

		return nil
	})

	return results, nil
}

func Grep(fSys fs.FS, path string, stdin io.Reader, keyword string, ignoreCase bool) ([]string, error) {
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