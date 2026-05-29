# AGENTS.md

This repository is `fcstask-backend`.

This file is the single source of truth for coding agents working in this repo.
The default expectation is to improve code quality without changing behavior unless the task explicitly requires a behavior change.

## Working Mode

When making changes:

1. preserve current behavior first
2. reduce coupling second
3. add regression tests for the seam you touched
4. keep the runtime graph easy to scan

If you are unsure between two refactors, choose the one that narrows dependencies and changes the public contract the least.

## Core Rules

1. Do not change behavior silently.
   Preserve HTTP statuses, JSON field names, route registration, middleware order, and current runtime flow unless the task explicitly asks for change.

2. Add regression tests with every structural refactor.
   If behavior is preserved, tests must prove it.

3. Prefer explicit boundaries over convenience coupling.
   Push business rules into `internal/service`.
   Keep HTTP-specific work in `internal/handler`.
   Keep storage logic in `internal/db/repo`.

4. Write idiomatic Go.
   Use small constructors, focused structs, early returns, `gofmt`, and clear names.

5. Avoid hidden side effects.
   Helpers may reduce duplication, but control flow must stay obvious.

## Repository Architecture

### `internal/app`

Composition root only.

Responsibilities:

- build repositories
- build services
- build handlers
- register middleware
- register routes
- build the HTTP server

Rules:

- must not contain business rules
- should stay declarative and easy to scan
- should follow the runtime flow:
  1. build repositories
  2. build services
  3. build handlers
  4. register middleware
  5. register routes
  6. build the HTTP server

### `internal/controller`

Transport orchestration only.

Responsibilities:

- own the route catalog
- register all HTTP routes from one place
- bridge generated OpenAPI handlers to the handler layer

Rules:

- keep it thin
- do not place business logic here
- do not split route registration across multiple files or entry points
- keep route protection metadata next to route definitions

### `internal/handler`

HTTP adapter layer.

Responsibilities:

- parse and validate HTTP input
- read auth and session data from Echo context
- call service interfaces
- map results and domain errors back to HTTP
- format response payloads

Rules:

- depend on `I...Service` interfaces, not concrete service structs
- use shared helpers instead of repeating request plumbing
- do not call repositories directly
- do not implement business rules here unless they are strictly transport-specific
- use `serviceError` for translating service failures into HTTP responses

Current helper entry points:

- `authenticatedUser`
- `authenticatedSession`
- `mustAuthenticatedUser`
- `bindRequest`
- `parseUUIDParam`
- `serviceError`

### `internal/service`

Service layer.

Responsibilities:

- business rules
- orchestration
- permission checks
- workflow decisions

Rules:

- stay independent from Echo and HTTP
- return `*service.Error` for caller-facing failures
- expose focused methods that adapters can call directly
- own permission checks and workflow decisions

### `internal/db/repo`

Persistence adapter layer.

Responsibilities:

- GORM queries
- transaction boundaries
- storage-specific behavior

Rules:

- no HTTP concerns
- no handler-level parsing logic
- no business policy logic unless it is purely persistence-related
- keep ORM details encapsulated as much as possible

### `internal/db/model`

Persistence models and transport-shaped entities currently shared across layers.

Treat them as existing constraints of the codebase unless the task explicitly asks for a deeper redesign.

## Interface and Naming Rules

Prefer:

- capability-based interfaces such as `IAuthService`, `ICourseService`, `IAdminTaskService`
- short constructor names such as `NewUserHandler`, `NewCourseService`
- helper names that describe exact behavior

Avoid for new code:

- giant constructors that wire half the application inline
- helper names that look pure but also mutate HTTP state invisibly
- splitting route definitions from protected-route metadata

## Handler Rules

Handlers should:

- depend on `I...Service` interfaces, not concrete service structs
- call shared helpers such as `bindRequest`, `parseUUIDParam`, `authenticatedUser`, `authenticatedSession`, and `mustAuthenticatedUser`
- keep response mapping inside the handler layer
- keep control flow explicit after helper calls

Handlers should not:

- implement authorization rules inline if the rule can live in `service`
- reach into repository code
- duplicate UUID parsing or JSON binding logic across endpoints

## Service Rules

Services should:

- expose focused behavior methods that adapters can call directly
- own permission checks and workflow decisions
- remain independent from Echo and HTTP details
- return `*service.Error` for caller-facing failures

Services should not:

- know about HTTP status codes
- write HTTP responses
- depend on handler or controller packages

## Repository Rules

Repositories should:

- encapsulate SQL and GORM details
- expose intent-based methods
- keep transaction boundaries explicit
- stay thin and storage-focused

Repositories should not:

- contain permission logic
- contain HTTP concerns
- leak ORM-specific behavior upward unless unavoidable

## Refactor Rules

When refactoring:

1. start from the existing behavior
2. reduce coupling
3. move policy downward into `service`
4. centralize repeated transport plumbing in `handler`
5. keep route registration in one controller catalog
6. keep `internal/app` declarative and easy to scan

Do not broaden the public contract just because the internal design became cleaner.

## Testing Rules

Preferred coverage split:

- handler tests for HTTP parsing, status codes, payload mapping, and auth/session context handling
- service tests for business rules and permission logic
- app or composition tests for route and middleware invariants

Regression targets currently worth protecting:

- protected path list in `internal/app`
- shared request and context helper behavior in `internal/handler`
- course access decisions in `internal/service`
- route catalog registration and protected-path derivation in `internal/controller`

## Formatting and Style

- always run `gofmt`
- prefer early returns
- keep imports grouped by standard library, third-party, and internal packages
- keep comments rare and only for non-obvious intent
- preserve existing JSON field names and HTTP messages unless the task explicitly allows change

## Agent Prompt

You are contributing to `fcstask-backend`.

Produce production-grade Go code that preserves existing behavior unless the task explicitly asks for a behavior change.

Follow these constraints:

1. treat `internal/app` as the composition root only
2. treat `internal/controller` as the single route catalog and registration layer
3. treat `internal/handler` as the HTTP adapter layer
4. treat `internal/service` as the service layer
5. treat `internal/db/repo` as persistence adapters
6. prefer interface boundaries at the adapter edge
7. do not change behavior silently
8. keep helpers explicit
9. write idiomatic Go
10. add tests with every structural change

Implementation checklist:

- start from the existing behavior and preserve it
- push business logic downward into `service`
- keep HTTP-only code in `handler`
- centralize repeated transport plumbing with small helpers
- keep all route definitions and protected flags in one controller catalog
- run `gofmt` and targeted `go test` after edits

If unsure, choose the refactor that reduces coupling without broadening the public contract.

## Pre-Completion Checklist

Before finishing a change:

- run `gofmt`
- run targeted `go test` for touched packages
- if a structural change was made, add or update regression tests
- verify composition wiring still matches the intended runtime graph
- verify no behavior was silently changed
