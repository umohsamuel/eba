# Detect the operating system
ifeq ($(OS),Windows_NT)
    SHELL := powershell.exe
    .SHELLFLAGS := -NoProfile -Command
    # Windows configuration (Using an explicit foreach loop that bypasses the $_ conflict)
    TIDY_CMD = foreach ($$f in Get-ChildItem -Filter go.mod -Recurse) { Push-Location $$f.DirectoryName; go mod tidy; $$code = $$LASTEXITCODE; Pop-Location; if ($$code -ne 0) { exit $$code } }
else
    # Linux / macOS configuration
    TIDY_CMD = find . -name go.mod | while IFS= read -r f; do (cd "$$(dirname "$$f")" && go mod tidy) || exit 1; done
endif

# Clean and tidy all Go modules
tidy:
	@$(TIDY_CMD)
