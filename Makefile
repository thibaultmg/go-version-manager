BINARY_NAME=gvs
GOBIN=$(CURDIR)/bin
LINTER=$(GOBIN)/golangci-lint
PREFIX?=$(HOME)/.local

# Build the gvs binary
build: fmt lint
	@mkdir -p $(GOBIN)
	go build -o $(GOBIN)/$(BINARY_NAME) .

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME) to $(PREFIX)/bin/"
	@mkdir -p $(PREFIX)/bin
	@cp $(GOBIN)/$(BINARY_NAME) $(PREFIX)/bin/
	@echo "Installation complete. Make sure '$(PREFIX)/bin' is in your PATH."

# Install bash completion script using the standard bash-completion framework
install-bash-completion: build
	@echo "Installing bash completion script to ~/.local/share/bash-completion/completions/"
	@mkdir -p ~/.local/share/bash-completion/completions
	@$(GOBIN)/$(BINARY_NAME) completion bash > ~/.local/share/bash-completion/completions/$(BINARY_NAME)
	@echo "Bash completion script installed."
	@echo "\nTo enable completion, please install the 'bash-completion' package."
	@echo "On macOS with Homebrew, you can run: brew install bash-completion@2"
	@echo "\nThen, add the following line to your ~/.bash_profile or ~/.bashrc:"
	@echo '  [[ -r "/usr/local/etc/profile.d/bash_completion.sh" ]] && . "/usr/local/etc/profile.d/bash_completion.sh"'
	@echo "\nFinally, restart your shell for the changes to take effect."


# Install zsh completion script
install-zsh-completion: build
	@echo "Installing zsh completion script to ~/.zsh/completions/_gvs"
	@mkdir -p ~/.zsh/completions
	@$(GOBIN)/$(BINARY_NAME) completion bash > ~/.zsh/completions/_$(BINARY_NAME)
	@echo "Zsh completion script installed."
	@echo "\nPlease add the following to your ~/.zshrc if it's not already there (order matters):"
	@echo 'fpath=(~/.zsh/completions $fpath)'
	@echo 'autoload -U bashcompinit'
	@echo 'bashcompinit'
	@echo 'autoload -U compinit'
	@echo 'compinit'
	@echo "\nThen restart your shell."

# Tooling
tools: $(LINTER)

$(LINTER):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(GOBIN) v2.5.0

fmt:
	go fmt ./...

lint: tools
	$(LINTER) run

clean:
	rm -rf $(GOBIN)

.PHONY: build install install-bash-completion install-zsh-completion lint clean fmt
