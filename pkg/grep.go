package grep

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrIsDirectory = errors.New("is a directory")
)

type GrepOptions struct {
	OrigPath string
	Path string
	Stdin io.Reader
	Keyword string
	FileWName string
	IgnoreCase bool
	LinesBeforeMatch int
	SearchDir bool
	LineCount bool
}

type GrepResult struct {
	Path string
	MatchedLines []string
	LineCount int
	Error error
}

func getRelPath(fSys fs.FS, arg string) (relPath string, err error) {
	absPath, err := filepath.Abs(filepath.Clean(arg))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	root := fmt.Sprintf("%s", fSys)
	relPath, err = filepath.Rel(root, absPath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

func GrepR(fSys fs.FS, options GrepOptions) []GrepResult {
	var wg sync.WaitGroup
	var outputChans []chan GrepResult

	fs.WalkDir(fSys, options.Path, func(path string, d fs.DirEntry, err error) error {
		outputChan := make(chan GrepResult)
		outputChans = append(outputChans, outputChan)

		wg.Add(1)
		go func(outputChan chan GrepResult) {
			defer wg.Done()
			defer close(outputChan)
			var x = options.OrigPath

			if err != nil {
				outputChan <- GrepResult{Error: err}
				return
			}

			if d.IsDir() {
				return
			}

			 // Compute the relative path
			 relPath, err := filepath.Rel(options.Path, path)
			 if err != nil {
				 outputChan <- GrepResult{Error: err}
				 return
			 }

			options := GrepOptions{Path: path, OrigPath: relPath, Keyword: options.Keyword, IgnoreCase: options.IgnoreCase, LinesBeforeMatch: options.LinesBeforeMatch, LineCount: options.LineCount}
			result := Grep(fSys, options)
			if result.Error != nil {
				outputChan <- result
				return
			}
			
			if len(result.MatchedLines) == 0 && result.LineCount == 0 {
				return
			}
			
			result.Path = filepath.Clean(x + string(os.PathSeparator) + options.OrigPath)
			outputChan <- result
		} (outputChan)

		return nil
	})

	var results []GrepResult
	for _, outputChan := range outputChans {
		result := <-outputChan
		if len(result.MatchedLines) == 0 && result.LineCount == 0 {
			continue
		}
		results = append(results, result)
	}
	return results
}

func Grep(fSys fs.FS, option GrepOptions) GrepResult {
	r, cleanup, err := getReader(fSys, option)
	if err != nil {
		return GrepResult{Error: err}
	}
	defer cleanup()

	result, err := searchString(r, option)
	if err != nil {
		return GrepResult{Error: err}
	}

	res := GrepResult{
		Path: option.OrigPath,
	}
	// res := GrepResult{MatchedLines: result}
	if option.LineCount {
		res.LineCount = len(result)
	} else {
		res.MatchedLines = result
	}

	return res
}

func getReader(fSys fs.FS, option GrepOptions) (io.Reader, func(), error) {
	if option.Path != "" {
		err := isValid(fSys, option.Path, option.OrigPath)
		if err != nil {
			return nil, nil, err
		}

		file, err := fSys.Open(option.Path)
		if err != nil {
			return nil, nil, err
		}
		return file, func() {file.Close()}, nil
	}
	return option.Stdin, func() {}, nil
}

func searchString(r io.Reader, options GrepOptions) ([]string, error) {
	grepBuffBefore := NewGrepBuffer(options.LinesBeforeMatch)
	
	keyword := options.Keyword
	if options.IgnoreCase {		// normalising keyword if ignoreCase
		keyword = strings.ToLower(options.Keyword)
	}

	var result []string
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()

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

func isValid(fSys fs.FS, path, origPath string) error {
	fileInfo, err := fs.Stat(fSys, path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", origPath, fs.ErrNotExist)
		}
		return fmt.Errorf("%s: %w", path, err)
	}

	// checks for directory
	if fileInfo.IsDir() {
		return fmt.Errorf("%s: %w", origPath, ErrIsDirectory)
	}

	// checks for permissions
	// looks hacky, might have to change later
	if fileInfo.Mode().Perm()&400 == 0 {
		return fmt.Errorf("%s: %w", path, fs.ErrPermission)
	}

	return nil
}


