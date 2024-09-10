# Grep exercise

This excerise deals with a command-line program that implements Unix grep like functionality.

This exercise has been solved in a TDD fashion. Please refer to the execise [here](https://one2n.io/go-bootcamp/go-projects/grep-in-go/grep-exercise).
 

## Features
This program supports searching files, directory recusively, and STDIN. It can also write the output to file, and perform case-sensitive search.

Options are as follows:
  - **-r**: recursive search in a directory
  - **-i**: case-sensitive search
  - **-o**: write output to file
  - **-A**: print n lines after the match
  - **-B**: print n lines before the match
  - **-C**: only print count of matches instead of actual matched lines

## Usage

1. Run the below command to build the binary. It has been saved in the bin directory.
```
make build
```

2. For using the program, arguments and options can be passed as usual. Some examples are as follows:

- For searching on a single file:
    ```
    ./bin/go-grep <search-string> <file-name>
    ```
- For searching in a directory recusively:
    ```
    ./bin/go-grep <search-string> <file-name> -r
    ```
- For searching in a directory recusively and writing output to file:
    ```
    ./bin/go-grep <search-string> <file-name> -r -o <output-file>
    ```