---
allowed-tools: Read, Write, Edit, Bash, WebFetch
description: Recurring dependency update workflow
disable-model-invocation: true
---

Work through each phase of the workflow in order. Pause at the human checkpoints before continuing.

## Phase 1 — Dev & CI envs: discover latest versions

Perform all lookups in parallel:

- **Latest stable Go release**: fetch `https://go.dev/dl/` and extract the latest full stable version (e.g. `1.27.2`).
- **Latest Node.js LTS release**: fetch `https://nodejs.org/dist/index.json` and find the latest TLS version (e.g. `24.15.0`). This endpoint returns a JSON array of every Node.js release, sorted newest-first. Each entry includes an `lts` field: a string codename (e.g. "Jod", "Iron") for LTS releases, or the boolean `false` for non-LTS.
- **Latest uv release**: fetch `https://github.com/astral-sh/uv/releases/latest` and extract the tag (e.g. `0.11.0`).
- **Latest DuckDB release**: fetch `https://api.github.com/repos/duckdb/duckdb/releases/latest` and extract `.tag_name` (e.g. `v1.5.2`).

## Phase 2 — Dev & CI envs: update versions

Using the versions discovered above:

1. **`extras/docker/Dockerfile`**:
   - Update `GO_VERSION`, `NODE_VERSION`, `UV_VERSION`, and `DUCKDB_VERSION` environment variables.
   - Do not update `FPM_VERSION`.
2. **`.github/workflows/main.yml`**:
   - Update `GO_VERSION` environment variable.
3. **`docker-compose.yml`**:
   - Update `GO_VERSION` in the example command.

FPM version is referenced in `extras/docker/Dockerfile` and `extras/github/docker/Dockerfile-*`, but do not update it in this workflow.

## Phase 3 — Human checkpoint: rebuild Docker dev env

Show a diff/summary of all changes made so far.

Then say:

> **Action required**: please rebuild the Docker dev environment, review it, and confirm when it's ready to be used.

Wait for the human to confirm before continuing.

## Phase 4 — Go: discover latest versions

Perform all lookups in parallel (reuse the Go version already fetched in Phase 1):

- **Latest golangci-lint release**: fetch `https://api.github.com/repos/golangci/golangci-lint/releases/latest` and extract `.tag_name` (e.g. `v2.13.0`).
- **Latest `modernize` version**: fetch `https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize?tab=versions` and extract latest version.

## Phase 5 — Go: update Go and tools

Apply the following updates to the codebase:

1. **`go.mod`**:
   - `go` directive: update to the **major.minor** part only (e.g. `go 1.27`).
   - `toolchain` directive: update to the full version (e.g. `toolchain go1.27.2`).

2. **`Makefile`**:
   - `go mod tidy -compat=` argument: update to the **major.minor** that matches the new `go` directive (e.g. `-compat=1.27`).
   - `golangci-lint` install line in the `lint` target: update the version tag (e.g. `v2.13.0`).
   - `modernize` `go run` line in the `modernize` target: update the `@vX.Y.Z` suffix.

## Phase 6 — Go: upgrade Go module dependencies

Run the following sequence inside the dev container, stopping on the first failure:

```bash
# 1. Upgrade all direct and indirect dependencies.
docker compose exec --user dev dev bash -i -c 'cd /mnt/host && go get -u ./...'

# 2. Regenerate mocks, tidy, format, lint, vet, modernize, test.
docker compose exec --user dev dev bash -i -c 'cd /mnt/host && make mocks mod fmt lint vet modernize test'
```

If any step fails, diagnose and fix the issue, then re-run from the failing step. Do not proceed to the next phase until the full sequence passes cleanly.

## Phase 7 — Node.js: upgrade dependencies

`assets/web/` contains a Node.js project. Apply the following steps inside the dev container (i.e. `docker compose exec --user dev dev bash -i -c 'cd /mnt/host/assets/web && ...`):

1. `rm -rf package-lock.json node_modules && npm update --save && npm audit fix`

2. Check outdate packages with `npm outdated`. For each outdated package, update it using `npm install --save-exact <package>@latest` (or `--save-dev` if it's a dev dependency), then re-run `npm outdated` to check for any new updates revealed by the previous update. Repeat until no more updates are available.

3. Use the `web-build` target in the Makefile to verify the build is still working. This will generate/overwrite some files in `assets/static/`. Make sure to revert all these changes once the build is confirmed working.

## Phase 8 — Human checkpoint: review

Show a full summary of every change made so far, including all version bumps (old → new) and any other changes.

Then say:

> **Review requested**: please check the changes above. Reply "ok" (or equivalent) to continue, or describe any issues to address first.

Wait for approval before continuing.

## Phase 9 — Update CHANGELOG.md

Prepend a new entry to `CHANGELOG.md` following the existing format:

```
- ?:
    + Updated dependencies:
        * Dev & CI environments:
            - <name>: <old> → <new>.
            - ...
        * Go:
            - <name>: <old> → <new>.
            - ...
        * Node.js:
            - <name>: <old> → <new>.
            - ...
```

Rules:
- The `?` is a literal placeholder — do NOT substitute a version number; it will be filled in later during the release process.
- List each changed item as its own `*` bullet with **old → new** versions explicitly.
- Only include **direct** dependencies (i.e. those without `// indirect` in `go.mod`), plus named tool versions (Go, uv, golangci-lint, modernize, etc.). **Do not list indirect Go module dependencies.**
- If unreleased changes are already listed, merge them appropriately into the new entry.
