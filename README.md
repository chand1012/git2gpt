# git2gpt

git2gpt is a command-line utility that converts a Git repository to a text file. The output text file represents the Git repository in a structured format. Each file in the repository is represented by a section that starts with a line containing four hyphens (----), followed by a line with the file's path and name, and then the file's content. The text file ends with a line containing the symbols --END--.

## Installation

First, make sure you have the Go programming language installed on your system. You can download it from [the official Go website](https://golang.org/dl/).

To install the `git2gpt` utility, run the following command:

```bash
go install github.com/chand1012/git2gpt
```

This command will download and install the git2gpt binary to your `$GOPATH/bin` directory. Make sure your `$GOPATH/bin` is included in your `$PATH` to use the `git2gpt` command.

## Usage

To use the git2gpt utility, run the following command:

```bash
git2gpt [flags] /path/to/git/repository
```

### Flags

* `-p`,    `--preamble`: Path to a text file containing a preamble to include at the beginning of the output file.
* `-o`,    `--output`: Path to the output file. If not specified, will print to standard output.

## Contributing

Contributions are welcome! To contribute, please submit a pull request or open an issue on the GitHub repository.

## License

git2gpt is licensed under the MIT License. See the [LICENSE](LICENSE) file for more information.
