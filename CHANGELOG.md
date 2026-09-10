# Changelog

All notable changes to this project are documented here. The format is
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html). While on
`0.x`, a breaking change bumps the minor.

## [Unreleased]

## [0.2.0] — 2026-09-10

### Changed

- `export` now emits `export KEY='value'` shell statements instead of bare
  `KEY=value` lines, so `eval "$(tiroir export)"` applies the store directly
  into the current shell. Values are single-quoted to survive spaces and
  metacharacters. Breaking on 0.x because the output contract changed.

## [0.1.0] — 2026-09-10

### Added

- Encrypted key-value store for shell variables, held at rest as an
  AES-256-GCM ciphertext blob in `~/.tiroir` (0600) keyed by a sibling
  `~/.tiroir.key`. Defeats naive value-scanners; not a real security boundary.
- CLI: `get`, `set`, `list`, `delete`, `export`. No `run` prefix command.
- `export` prints `KEY=value` lines so a shell can source the store at login.
- Atomic writes; collisions-free temp-file rename; `0600` permissions enforced
  on both store and key.