# ORZ LLM Context

This document is for large models and automation agents editing this repository.

## Project Goal

ORZ is not trying to be a general-purpose Go web framework for the whole ecosystem.
It is a team standardization layer for internal Go services.

The primary goals are:

- keep project structure and runtime contracts consistent
- reduce repeated setup work for config, logging, database, HTTP, repository, and paging
- make the common path fast without blocking escape hatches for custom business logic

## Module Topology

This repository is a multi-module workspace.

- root module: `github.com/go-orz/orz`
- driver modules:
  - `github.com/go-orz/orz/drivers/sqlite`
  - `github.com/go-orz/orz/drivers/mysql`
  - `github.com/go-orz/orz/drivers/postgres`
- local integration test module: `tests/pagebuilder`
- local example modules: `examples/*`

Local development uses the top-level `go.work` file.

Important rule:

- keep `replace github.com/go-orz/orz => ../..` in driver module `go.mod` files for local standalone development
- keep the driver module `require github.com/go-orz/orz vX.Y.Z` on a real release version for publishing

Reason:

- `replace` is only honored when the driver module itself is the main module
- external consumers ignore dependency-module `replace` directives
- published driver submodules still need a real root-module version in `require`
- `go.work` keeps local multi-module development convenient across the repo

## Release Model

Driver submodules must be released together with a compatible root-module version.

Current coordinated release target in driver `go.mod` files: `v0.2.11`

Expected tags for a coordinated release:

- root module tag: `v0.2.11`
- sqlite driver tag: `drivers/sqlite/v0.2.11`
- mysql driver tag: `drivers/mysql/v0.2.11`
- postgres driver tag: `drivers/postgres/v0.2.11`

If the next release version changes, update all three driver `go.mod` files before tagging.

Do not publish a driver submodule tag that points at a root-module version which does not contain the required API.

## Runtime Contracts

These contracts should be treated as framework invariants.

- HTTP responses use the `Response{code,message,data}` envelope
- paging helpers return `PageResult{items,total}`
- framework initialization order is fixed internally, regardless of option call order
- driver registration is explicit through blank imports of driver submodules
- running without HTTP is valid and should behave like daemon mode

## Query Safety Rules

This repository allows dynamic query construction, so identifier validation is a hard requirement.

Required rules:

- all sort fields must be validated before interpolation
- all matcher fields must be validated before interpolation
- all keyword fields must be validated before interpolation
- raw values must still be passed as SQL parameters, never interpolated into SQL strings

The canonical helper for matcher/keyword field validation is:

- `qualifyQueryField()` in `repository.go`

If you add a new query helper that accepts a field name, route it through the same validation path.

Do not assume `CamelToSnake()` is a sanitizer. It is only a naming conversion helper.

## Design Bias

When making edits, prefer these biases:

- predictable behavior over magic behavior
- explicit configuration over subjective defaults
- narrow, boring public APIs over clever abstractions
- compatibility with direct GORM/Echo escape hatches

Avoid:

- adding new implicit global behavior
- hiding ordering or initialization requirements behind option names
- making helper APIs look declarative if they still depend on call order

## Golden Path

The intended common path for service code is:

1. load config
2. initialize logger
3. initialize database
4. initialize HTTP
5. configure application routes/services
6. use repository/service helpers for standard CRUD and paging
7. drop to raw GORM or SQL for nonstandard cases

## Test Strategy

Root-module tests should stay as pure unit tests when possible.

Real driver-dependent tests should live in dedicated integration modules such as `tests/pagebuilder`.

Recommended validation commands:

```bash
go test ./...
go test ./... ./drivers/sqlite ./drivers/mysql ./drivers/postgres
go test ./... ./tests/pagebuilder
```

Example modules should still compile:

```bash
go test -mod=readonly ./...
```

Run those from each example module directory.

## Editing Guardrails

When changing this repository:

- keep README and examples aligned with the real public API
- update tests when runtime contracts change
- do not widen trust boundaries silently
- do not bypass query-field validation for convenience
- do not add database-driver dependencies back into the root module
