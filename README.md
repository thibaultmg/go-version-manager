# gvs - Go Version Manager

A simple, no dependency, POSIX-compliant command-line tool for managing multiple Go versions. `gvs` allows you to easily install, switch between, and remove different Go SDKs.

The repository is named `go-version-manager` for discoverability, but the binary is `gvs` for a better command-line experience.

## Features

- List and install any official Go version.
- Switch between installed Go versions seamlessly.
- Cleanly remove old Go versions.
- Shell completion for Bash and Zsh.

## Installation

### From Source (Recommended)

To build and install `gvs` from source, you'll need a working Go environment.

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/your-username/go-version-manager.git
    cd go-version-manager
    ```

2.  **Build and Install:**
    The `Makefile` provides a simple way to install the binary. By default, it will install `gvs` to `~/.local/bin/`.
    ```sh
    make install
    ```
    Make sure `~/.local/bin` is in your `$PATH`.

    To install to a different location (like `/usr/local`), you can override the `PREFIX` variable:
    ```sh
    make install PREFIX=/usr/local
    ```

## Usage

`gvs` provides several commands to manage your Go versions.

### list-remote

List available Go versions for installation.

```sh
gvs list-remote
```

To limit the number of versions shown:

```sh
gvs list-remote -limit 20
```

### list

List all locally installed Go versions.

```sh
gvs list
```

### install

Install a specific Go version.

```sh
gvs install 1.21.0
```

### use

Switch the active Go version. This command creates a symlink at `~/.gvs/bin/go`, so you must add this directory to your `$PATH`.

```sh
gvs use 1.21.0
```

After running `use`, add the following to your shell's configuration file (`~/.bash_profile`, `~/.zshrc`, etc.) if you haven't already:

```sh
export PATH="$HOME/.gvs/bin:$PATH"
```

### remove

Uninstall a Go version.

```sh
gvs remove 1.21.0
```

## Shell Completion

`gvs` provides shell completion for Bash and Zsh. The `Makefile` provides targets to simplify installation.

### Zsh (Recommended)

The install target will place a completion script in `~/.zsh/completions`.

1.  Run the make target:
    ```sh
    make install-zsh-completion
    ```
2.  Add the instructions provided by the command to your `~/.zshrc`.
3.  Restart your shell.

### Bash

The install target uses the standard `bash-completion` framework.

1.  Run the make target:
    ```sh
    make install-bash-completion
    ```
2.  Follow the instructions provided by the command to install the `bash-completion` package (e.g., via Homebrew) and configure your `~/.bash_profile`.
3.  Restart your shell.

## Contributing

Contributions are welcome! Please feel free to submit a pull request.

Before submitting, please ensure your code is formatted and passes the linter checks:
```sh
make lint
```

## License

This project is open source and available under the [MIT License](LICENSE).
