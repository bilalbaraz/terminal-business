# terminal-business Persistence Design

## Scope and Alignment
- Binary/UI alignment: `tbiz` is interactive-only; persistence supports Boot -> New/Load -> Dashboard flows.
- Local-first only: all data is on local filesystem, no network dependencies.
- Deterministic-safe: persistence captures all required state for deterministic simulation replay.
- Cross-platform support: Linux, macOS, Windows native data directories.

## 1) High-Level Persistence Architecture

### Goals
- Support multiple saves concurrently.
- Provide fast Load Game listing without full save deserialization.
- Guarantee crash-safe atomic writes.
- Support schema versioning and migration.
- Recover from index/save corruption without UI crashes.

### Core components
- `SaveStore`: read/write full save files by `SaveID`.
- `SaveIndexStore`: read/write compact save index used by Load Game list.
- `AtomicWriter`: shared write primitive (temp -> fsync -> rename -> dir fsync).
- `MigrationRunner`: upgrades old schemas to current schema deterministically.
- `RecoveryService`: rebuilds index by scanning save directory and validating entries.

### Data flow
1. New game create:
- Build initial `SaveFile` with identity + seed + tick 0 + snapshot.
- Persist save atomically.
- Update index atomically.
- Return loaded state to app.

2. Save existing game:
- Update snapshot + `LastPlayedAt` + `TickCounter`.
- Persist save atomically.
- Update index atomically.

3. Load game:
- Read index entry -> open referenced save.
- Validate checksum/schema.
- Run migration if needed.
- Update `LastPlayedAt` and persist on successful load flow.

## 2) Save File Model

### `SaveFile` structure
- `SaveIdentity`
- `SimulationSeed`
- `TickCounter`
- `DomainStateSnapshot`
- `Version` (schema version for save file payload)
- `Checksum` (recommended integrity field)
- `AppVersion` (optional, diagnostics/migration support)

### `SaveIdentity` structure
- `SaveID` (UUID)
- `CompanyID`
- `CompanyName`
- `CompanyType`
- `CreatedAt`
- `LastPlayedAt`
- `Version`

### Rules
- `SaveIdentity` is immutable except `LastPlayedAt` and future explicit rename fields.
- `SaveID` is internal identifier and never primary display label.
- `CompanyName` is user-facing label used in Load Game display.
- `SimulationSeed` is immutable after save creation.
- `TickCounter` monotonically increases by deterministic tick execution.
- `LastPlayedAt` updates on successful load and successful save commit.

### On-disk format
- JSON for portability and future export/import compatibility.
- Deterministic encoding requirements:
  - Stable field names.
  - Canonical ordering where feasible.
  - Fixed numeric precision policy to reduce drift issues.

## 3) Save Index Model (Required)

### Purpose
- Fast startup and fast Load Game list rendering without scanning every save file.

### `SaveIndex` structure
- `Version`
- `Entries[]`

Each entry contains:
- `SaveID`
- `CompanyName`
- `CompanyType`
- `LastPlayedAt`
- `SaveFilePath`
- `Version`

### Rules
- Sort order: `LastPlayedAt DESC`.
- Deterministic tie-breaker: `SaveID ASC`.
- Index must be rebuildable entirely from save files.
- Index updates are atomic and coordinated with save writes.
- Startup path reads index first; no full directory scan on healthy index.

### Load Game display contract
- Row format is exactly:
  - `"{Company Name} - {Company Type}"`

## 4) Storage Locations (OS Native)

Base app data directory: `terminal-business`

Linux:
- Base: `$XDG_DATA_HOME/terminal-business` or `~/.local/share/terminal-business`

macOS:
- Base: `~/Library/Application Support/terminal-business`

Windows:
- Base: `%AppData%/terminal-business`

### Directory layout
- Saves directory:
  - `<base>/saves/`
  - Save path pattern: `<base>/saves/<save_id>.json`
- Save index:
  - `<base>/index/saves-index.json`
- Logs (optional):
  - `<base>/logs/tbiz.log`
- Backups/exports (future):
  - `<base>/exports/`

## 5) Atomic Write Strategy (Crash Safe)

All save and index writes use the same protocol:
1. Write content to temp file in same directory (`*.tmp`).
2. `fsync` temp file.
3. Atomic rename temp -> final file.
4. `fsync` parent directory (if platform supports it).

### Guarantees
- Prevents partial file visibility.
- Prevents truncated files being treated as committed state.
- Reduces save/index mismatch window.

### Save + index consistency strategy
- Commit order:
  1. Commit save file atomically.
  2. Commit index atomically.
- Recovery rule:
  - If save exists but index is stale, recovery rebuild/index refresh fixes mismatch.
- Optional enhancement (recommended): lightweight write-ahead journal marker for two-phase recovery intent.

## 6) Deterministic Replay Guarantees

Persistence must include everything needed for exact replay:
- `SimulationSeed`
- `TickCounter`
- `DomainStateSnapshot`
- Save/config version references

### Determinism contract
- Same save payload + same tick count progression + same balance config version => same simulation state.

### Forbidden patterns
- `time.Now()` inside domain or simulation.
- Global RNG usage.
- Hidden time-based triggers not represented in state.

## 7) Corruption Handling and Recovery

### Corrupted save file
Behavior:
- Mark entry as invalid for current session.
- Exclude from selectable Load Game rows.
- Record warning for UI (`N corrupted saves skipped`).
- Continue without crash.

### Missing save file referenced by index
Behavior:
- Drop stale index entry during validation pass.
- Mark index dirty and rewrite atomically.

### Corrupted index file
Behavior:
- Attempt parse recovery only if safe.
- Otherwise rebuild index by scanning `saves/` directory.
- Validate each save file before insertion.
- Skip broken saves, emit warnings, persist rebuilt index atomically.

### Recovery trigger points
- On app startup if index read fails.
- On entering Load Game if index fails validation.
- On explicit maintenance action (future settings UI).

### UI safety rule
- Persistence errors must surface as non-fatal modal/toast states.
- UI process must not terminate due to recoverable persistence errors.

## 8) Schema Versioning + Migration

### Version fields
- Save file includes `Version` (schema version).
- Save index includes `Version`.
- Optional `AppVersion` for diagnostics.

### Migration strategy
- Migrations are ordered, incremental, and deterministic.
- Execution order:
  - `vN -> vN+1 -> ... -> current`
- Each migration:
  - pure transformation on decoded save model
  - explicit pre/post validation
  - idempotence check where feasible

### Load path
1. Decode raw save.
2. Compare save schema version to current schema.
3. If old: run ordered migrations.
4. Validate invariants and replay-critical fields.
5. Persist migrated save atomically (optional eager rewrite policy).

### Backward compatibility policy
- Prefer additive changes with defaults.
- Breaking changes require migration or explicit unsupported-version error.
- Unsupported saves must fail gracefully with actionable message.

### Migration testing approach
- Fixture-driven tests for each historical version.
- Golden before/after transformation snapshots.
- Roundtrip tests ensuring migrated saves remain deterministic in simulation.

## 9) Export / Backup Strategy

### Export goals
- Allow users to back up saves manually.
- Support future import feature.

### Export format
- Full `SaveFile` JSON.
- Human-readable and stable schema keys.
- Includes schema version and checksum.

### Backup behavior
- Export is read-only from active save and does not mutate live state.
- Export path defaults to local filesystem (`exports/`), user-selectable in future UI.

### Future import compatibility
- Import pipeline will reuse same decode/validate/migrate flow as normal load.
- Conflicting `SaveID` handled by policy (future): reject, overwrite, or clone with new `SaveID`.

## 10) Performance Requirements

Targets (typical local SSD/dev machine):
- Save commit: <100ms
- Load commit/read path: <100ms
- Load Game list rendering: near-instant via index (no full scan)

### Performance constraints
- Avoid full directory scan on every boot.
- Use compact index entries to minimize parse overhead.
- Stream/encode JSON without excessive intermediate allocations.
- Keep hot-path allocations bounded in save/load operations.

## 11) Testing Strategy (Design Commitment)

### Required test suites
- Atomic write tests:
  - temp file + rename semantics
  - crash-simulation/fault injection around each commit step
- Index rebuild tests:
  - rebuild from valid save set
  - deterministic sort order validation
- Corruption recovery tests:
  - corrupted save, missing save, corrupted index scenarios
- Deterministic replay persistence tests:
  - persisted seed/ticks/snapshot produce identical replay outcome
- Migration tests:
  - per-version upgrade chain + invariant validation

### Test conventions
- Use temp directories for all filesystem tests.
- Use `testdata/` fixtures for legacy schema samples.
- Use table-driven cases for edge conditions.

### Coverage target
- Future implementation requirement: 100% coverage for `internal/persistence` package set.

## 12) Non-Functional Constraints
- Offline only.
- No cloud sync assumptions.
- No networking.
- Cross-platform filesystem compatibility.
- Determinism safety is mandatory.
