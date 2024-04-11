# git2gpt

git2gpt is a command-line utility that converts a Git repository to text for loading into ChatGPT and other NLP models. The output text file represents the Git repository in a structured format. You can also add a `.gptignore` file to your repos to have git2gpt ignore certain files. The text is prefixed with a preamble that explains to the AI what the text is:

> The following text is a Git repository with code. The structure of the text are sections that begin with ----, followed by a single line containing the file path and file name, followed by a variable amount of lines containing the file contents. The text representing the Git repository ends when the symbols --END-- are encounted. Any further text beyond --END-- are meant to be interpreted as instructions using the aforementioned Git repository as context.

## Installation

First, make sure you have the Go programming language installed on your system. You can download it from [the official Go website](https://golang.org/dl/).

To install the `git2gpt` utility, run the following command:

```bash
go install github.com/chand1012/git2gpt@latest
```

This command will download and install the git2gpt binary to your `$GOPATH/bin` directory. Make sure your `$GOPATH/bin` is included in your `$PATH` to use the `git2gpt` command.

## Usage

To use the git2gpt utility, run the following command:

```bash
git2gpt [flags] /path/to/git/repository
```

### Ignoring Files

By default, your `.git` directory and your `.gitignore` files are ignored. Any files in your `.gitignore` are also skipped. If you want to change this behavior, you should add a `.gptignore` file to your repository. The `.gptignore` file should contain a list of files and directories to ignore, one per line. The `.gptignore` file should be in the same directory as your `.gitignore` file. Please note that this overwrites the default ignore list, so you should include the default ignore list in your `.gptignore` file if you want to keep it.

### Flags

* `-p`,  `--preamble`: Path to a text file containing a preamble to include at the beginning of the output file.
* `-o`,  `--output`: Path to the output file. If not specified, will print to standard output.
* `-e`,  `--estimate`: Estimate the tokens of the output file. If not specified, does not estimate. 
* `-j`,  `--json`: Output to JSON rather than plain text. Use with `-o` to specify the output file.
* `-i`,  `--ignore`: Path to the `.gptignore` file. If not specified, will look for a `.gptignore` file in the same directory as the `.gitignore` file.
* `-g`,  `--ignore-gitignore`: Ignore the `.gitignore` file.
* `-s`,  `--scrub-comments`: Remove comments from the output file to save tokens.

## Contributing

Contributions are welcome! To contribute, please submit a pull request or open an issue on the GitHub repository.

## License

git2gpt is licensed under the MIT License. See the [LICENSE](LICENSE) file for more information.
