# npkl-go

A utility for calculating size and managing node_modules directory.

## Description

This program allows you to:
- Check if node_modules directory exists
- Calculate the total size of node_modules directory
- Optionally delete the node_modules directory

## Usage

```bash
go run main.go
```

## Requirements

- Go 1.21 or higher

## Features

- Multi-threaded file size calculation
- Atomic operations for safe parallel execution
- Recursive directory traversal

## Build

### Basic Build
```bash
go build -o npkl
```

### Optimized Production Build
```bash
go build -ldflags="-w -s" -o npkl
```
Flags explanation:
- `-w`: Disables DWARF generation (debug information)
- `-s`: Disables symbol table generation
These flags help reduce the executable size.

### Cross-Platform Builds

For Windows (PowerShell):
```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -ldflags="-w -s" -o npkl.exe
```

For Windows (CMD):
```cmd
set GOOS=windows && set GOARCH=amd64 && go build -ldflags="-w -s" -o npkl.exe
```

For Linux/macOS (Bash):
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o npkl     # For Linux
GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o npkl    # For macOS
```
