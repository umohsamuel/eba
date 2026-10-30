# Detect the operating system
ifeq ($(OS),Windows_NT)
    # Windows configuration (Using an explicit foreach loop that bypasses the $_ conflict)
    TIDY_CMD = powershell -Command "foreach (\$$f in Get-ChildItem -Filter go.mod -Recurse) { Push-Location \$$f.DirectoryName; go mod tidy; Pop-Location }"
else
    # Linux / macOS configuration
    TIDY_CMD = find . -name "go.mod" -execdir go mod tidy \;
endif

# Clean and tidy all Go modules
tidy:
	@$(TIDY_CMD)
