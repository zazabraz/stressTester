# Name of the binary output
BINARY_NAME=stressTester

# Go build command
BUILD_CMD=go build

# List of platforms
PLATFORMS=linux darwin windows

# Default target
.PHONY: all
all: build

# Target to build the project for all platforms
.PHONY: build
build: $(PLATFORMS)

# Targets for specific platforms
.PHONY: linux
linux:
	GOOS=linux GOARCH=amd64 $(BUILD_CMD) -o $(BINARY_NAME) main.go

.PHONY: mac
mac:
	GOOS=darwin GOARCH=amd64 $(BUILD_CMD) -o $(BINARY_NAME) main.go

.PHONY: windows
windows:
	GOOS=windows GOARCH=amd64 $(BUILD_CMD) -o $(BINARY_NAME).exe main.go

# Target to run the project with optional flags on Linux
.PHONY: run-linux
run-linux:
	@echo "Running Linux binary with flags: $(FLAGS)"
	@if [ -z "$(FLAGS)" ]; then \
		echo "Please specify FLAGS"; \
		exit 1; \
	fi
	./$(BINARY_NAME) $(FLAGS)

# Target to run the project with optional flags on macOS
.PHONY: run-mac
run-mac:
	@echo "Running macOS binary with flags: $(FLAGS)"
	@if [ -z "$(FLAGS)" ]; then \
		echo "Please specify FLAGS"; \
		exit 1; \
	fi
	./$(BINARY_NAME) $(FLAGS)

# Target to run the project with optional flags on Windows
.PHONY: run-windows
run-windows:
	@echo "Running Windows binary with flags: $(FLAGS)"
	@if [ -z "$(FLAGS)" ]; then \
		echo "Please specify FLAGS"; \
		exit 1; \
	fi
	@echo "$(FLAGS)" | cmd /c "$(BINARY_NAME).exe"