package grep

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"sync"
)

var (
	ErrIsDirectory = errors.New("is a directory")
)

type GrepOptions struct {
	Path string
	Stdin io.Reader
	Keyword string
	IgnoreCase bool
	LinesBeforeMatch int
	LineCount bool
}

func Grep(fSys fs.FS, options GrepOptions) ([]string, error) {
	r, cleanup, err := getReader(fSys, options.Path, options.Stdin)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	result, err := searchString(r, options)
	if err != nil {
		return nil, err
	}

	// if no matches, return nil
	if len(result) == 0 {
		return nil, nil
	}

	if options.LineCount {
		return []string{fmt.Sprintf("%d", len(result))}, nil
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

func searchString(r io.Reader, options GrepOptions) ([]string, error) {
	grepBuffBefore := NewGrepBuffer(options.LinesBeforeMatch)
	
	keyword := options.Keyword
	if options.IgnoreCase {		//normalising keyword if ignoreCase
		keyword = strings.ToLower(options.Keyword)
	}

	var result []string
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var line string
		line = scanner.Text()

		// normalising line if ignoreCase
		if options.IgnoreCase {
			line = strings.ToLower(scanner.Text())
		}

		// comparison and saving lines if matched
		if strings.Contains(line, keyword) {
			if options.LinesBeforeMatch > 0 {
				result = append(result, grepBuffBefore.Dump()...)
			}
			result = append(result, scanner.Text())
		}
		
		// save lines to buffer
		if options.LinesBeforeMatch > 0 {
			grepBuffBefore.Push(scanner.Text())
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

func GrepR(fSys fs.FS, options GrepOptions) ([][]string, error) {
	var wg sync.WaitGroup
	var outputChans []chan string
	var results [][]string

	fs.WalkDir(fSys, options.Path, func(path string, d fs.DirEntry, err error) error {
		outputChan := make(chan string)
		outputChans = append(outputChans, outputChan)

		wg.Add(1)
		go func(outputChan chan string) {
			defer wg.Done()
			defer close(outputChan)

			if err != nil {
				outputChan <- err.Error()
				return
			}

			if d.IsDir() {
				return
			}

			options := GrepOptions{Path: path, Keyword: options.Keyword, IgnoreCase: options.IgnoreCase, LinesBeforeMatch: options.LinesBeforeMatch, LineCount: options.LineCount}
			result, errGrep := Grep(fSys, options)
			if errGrep != nil {
				outputChan <- errGrep.Error()
				return
			}
			
			for _, r := range result {
				outputChan <- fmt.Sprintf("%s:%s", path, r)
			}
			
		}(outputChan)

		return nil
	})

	for _, outputChan := range outputChans {
		var result []string
		for resultStr := range outputChan {
			result = append(result, resultStr)
		}
		
		if len(result) > 0 {
			results = append(results, result)
		}
	}

	if len(results) == 0 {
		return nil, nil
	}

	return results, nil
}