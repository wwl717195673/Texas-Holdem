# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Texas-Holdem is a Go implementation of a Texas Hold'em Poker game. The project is in its early stages with minimal code.

## Build Commands

```bash
go mod init github.com/wilenwang/just_play/Texas-Holdem  # Initialize module (if not done)
go build ./...                                           # Build all packages
go test ./...                                            # Run all tests
go run .                                                 # Run main package
```

## Project Structure

```
prd/           # Product Requirements Documents directory
```

## Technology Stack

- Language: Go (Golang)
