# awesome-fiber-template

![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)

[中文说明](./README_zh.md)

`awesome-fiber-template` is a Fiber-focused Go backend template repository.

Instead of trying to cover many frameworks, this repository now stays focused on Fiber v3 and provides four template tiers so you can pick the amount of engineering baseline you actually want.

## Template Matrix

- [`v3/heavy`](./v3/heavy): full-featured edition with Redis, scheduler, service install/uninstall, WAF, Prometheus, Swagger, reusable helper packages, and a stronger infrastructure baseline
- [`v3/medium`](./v3/medium): balanced HTTP service edition that keeps Redis, WAF, service manager support, embedded UI, and most middleware, but removes scheduler and Prometheus complexity
- [`v3/light`](./v3/light): plain Go-style service edition that keeps the common API middleware baseline and optional embedded UI, while removing Redis, service manager support, and extra helper packages
- [`v3/extra-light`](./v3/extra-light): minimal edition with native CLI, SQLite-only setup, no built-in business demo, and only `recover + healthcheck`

## How To Choose

- Choose `heavy` if you want the most complete engineering baseline and do not mind extra infrastructure.
- Choose `medium` if you want a practical production-oriented HTTP template without scheduler and Prometheus overhead.
- Choose `light` if you want something closer to a normal Go project structure.
- Choose `extra-light` if you want the smallest clean starting point and prefer adding capabilities yourself.

## Quick Start

Pick one version and run it directly. For example:

```bash
cd v3/light
go run . serve
```

Each version is a standalone Go module with its own `go.mod`, config, README, and dependency boundary.

## Repository Goal

This repository is meant to help you skip repeated project setup work while still keeping the template boundaries clear:

- plain and readable project structure
- practical middleware and bootstrap defaults
- SQLite-first out-of-the-box demo experience where appropriate
- different template weights for different project sizes

## Notes

- The repository name now reflects the real maintenance scope: Fiber templates only.
- If you use one of these templates in your own project, replace the module path inside that version's `go.mod`.

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE) for details.
