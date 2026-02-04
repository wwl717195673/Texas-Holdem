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

## Notes
- 每一个方法都加上方法级别的中文注释
- 当一个方法超过 30 行代码时，给这个方法的关键节点也加上中文注释
- 每次代码实现都生成 1 个实现文档，写明本次生成做了什么 怎么做的,文件名为本次实现的基本功能 简短一点 markdown格式