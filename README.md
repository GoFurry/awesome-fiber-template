# awesome-go-template

![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)

[中文说明](./README_zh.md)

`awesome-go-template` is a backend template repository for quickly bootstrapping Go services with different web frameworks.

## Purpose

This repository is maintained as a collection of framework-specific scaffolds so you can:

- compare project structures across frameworks
- start a new backend service faster
- reuse common engineering patterns and starter code

## Repository Structure

- `fiber/v3/basic`: the most complete scaffold at the moment, including application bootstrap, config loading, routing, middleware, database integration, Redis, scheduling, metrics, and optional web UI embedding
- `chi`: framework placeholder / starter entry
- `echo`: framework placeholder / starter entry
- `gin`: framework placeholder / starter entry
- `net`: standard library based starter entry

## Quick Start

The currently available full example is under `fiber/v3/basic`.

```bash
cd fiber/v3/basic
go run . serve
```

Configuration example:

- server config: `fiber/v3/basic/config/server.yaml`
- database config supports `sqlite`, `postgres`, and `mysql`

## Notes

- This repository is intended to evolve as a template collection rather than a single runnable application at the root level.
- Different framework directories may be at different completion stages.

## License

This project is licensed under the MIT License. See `LICENSE` for details.
