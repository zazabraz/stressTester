# Name of the binary output
BINARY_NAME=stressTester

# Targets for specific platforms
.PHONY: linux
linux:
	GOOS=linux go build -o $(BINARY_NAME) main.go

.PHONY: mac
mac:
	GOOS=darwin go build -o $(BINARY_NAME) main.go

.PHONY: windows
windows:
	GOOS=windows go build -o $(BINARY_NAME).exe main.go

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