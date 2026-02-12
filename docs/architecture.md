# terminal-business Repository + Architecture Design

## Scope and Alignment
- Binary: `tbiz` only.
- CLI contract: interactive-only TUI, no flags, no subcommands.
- UX flow contract: Boot Menu (`New Game`, `Load Game`, `Help`, `Exit`) -> create/load save -> dashboard.
- System constraints: offline/local-first, minimal dependencies, deterministic simulation, pure domain.

## 1) Repository Structure (Go OSS Standard)

### Proposed tree
```text
terminal-business/
├── AGENTS.md
├── README.md
├── SECURITY.md
├── LICENSE
├── Makefile
├── go.mod
├── go.sum
├── cmd/
│   └── tbiz/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── application.go
│   │   ├── state_machine.go
│   │   ├── navigation.go
│   │   ├── errors.go
│   │   └── wiring.go
│   ├── ui/
│   │   ├── shell/
│   │   │   └── model.go
│   │   ├── screens/
│   │   │   ├── boot/
│   │   │   │   └── model.go
│   │   │   ├── newgame/
│   │   │   │   └── model.go
│   │   │   ├── loadgame/
│   │   │   │   └── model.go
│   │   │   ├── help/
│   │   │   │   └── model.go
│   │   │   └── dashboard/
│   │   │       └── model.go
│   │   ├── components/
│   │   │   ├── list/
│   │   │   ├── form/
│   │   │   ├── modal/
│   │   │   └── toast/
│   │   ├── keymap/
│   │   │   └── keymap.go
│   │   ├── focus/
│   │   │   └── focus.go
│   │   └── styles/
│   │       └── theme.go
│   ├── domain/
│   │   ├── company/
│   │   ├── employee/
│   │   ├── inventory/
│   │   ├── investors/
│   │   ├── finance/
│   │   ├── events/
│   │   ├── metrics/
│   │   └── invariants/
│   ├── sim/
│   │   ├── engine/
│   │   ├── tick/
│   │   ├── rng/
│   │   ├── events/
│   │   ├── recompute/
│   │   └── replay/
│   ├── persistence/
│   │   ├── saveindex/
│   │   ├── savestore/
│   │   ├── schema/
│   │   ├── migrate/
│   │   └── atomic/
│   ├── config/
│   │   ├── balance/
│   │   ├── defaults/
│   │   ├── schema/
│   │   └── loader/
│   ├── diag/
│   │   ├── log/
│   │   └── health/
│   └── platform/
│       ├── clock/
│       ├── filesystem/
│       └── random/
├── docs/
│   ├── ux.md
│   ├── economy.md
│   └── architecture.md
├── assets/
│   ├── branding/
│   ├── ascii/
│   └── sample/
├── scripts/
│   ├── test.sh
│   ├── lint.sh
│   └── coverage.sh
└── .github/
    └── workflows/
        ├── ci.yml
        └── release.yml
```

### Top-level file plans
- `AGENTS.md`: collaboration/automation instructions for contributors and coding agents.
- `README.md`: quick start (`go run ./cmd/tbiz`), gameplay overview, architecture links, contribution guide.
- `SECURITY.md` (plan-only): vulnerability reporting process, supported versions, response SLA.
- `Makefile` (recommended): standardized targets (`fmt`, `lint`, `test`, `coverage`, `build`).

## 2) Package Architecture (MVP)

### A) UI Layer (`internal/ui`)
Purpose: Bubble Tea rendering + interaction only.

Modules:
- `screens/boot`: boot menu model (`New Game`, `Load Game`, `Help`, `Exit`).
- `screens/newgame`: company name/type form model + inline validation display.
- `screens/loadgame`: save list model with display format `{Company Name} - {Company Type}`.
- `screens/help`: controls/help screen.
- `screens/dashboard`: shell and top-level in-game navigation.
- `components/*`: reusable list/form/modal/toast widgets.
- `keymap`: canonical key bindings and route tables.
- `focus`: deterministic focus graph and traversal helpers.
- `styles`: theme tokens (premium terminal look), responsive spacing.

Rules:
- UI does not perform file IO directly.
- UI emits intents/events to `app`; `app` returns state/view models.

### B) App / Orchestration Layer (`internal/app`)
Purpose: application state machine and dependency wiring.

Responsibilities:
- State transitions: boot -> create/load -> dashboard.
- Navigation coordinator across UI screens.
- Dependency injection: sim engine, persistence services, config, rng, clock.
- Error boundary strategy: convert runtime errors to user-visible modal/toast states.
- Panic-safe rendering boundary: recover at shell boundary, log locally, route to safe error view.

### C) Domain Layer (`internal/domain`) (PURE)
Purpose: business entities, invariants, and rules.

Responsibilities:
- Entity models: company, metrics, employees, inventory, investors, cap table, term sheets, events.
- Validation/invariant enforcement (equity=100%, non-silent negative cash, idempotent recompute constraints).
- Pure rules: hiring effects, purchase effects, term sheet validity.

Rules:
- No IO, no Bubble Tea, no persistence coupling, no non-deterministic clock/rng access.
- Stdlib-only dependency policy.

### D) Simulation Layer (`internal/sim`) (DETERMINISTIC)
Purpose: deterministic tick progression.

Responsibilities:
- Tick engine with fixed order.
- Seeded event selection via injectable RNG.
- Metrics recomputation orchestration.
- Replay hooks/checkpoints for deterministic verification.

Rules:
- Depends on domain types/rules.
- No UI deps, no global rng/time.

### E) Persistence Layer (`internal/persistence`) (LOCAL-FIRST)
Purpose: local save storage and retrieval.

Responsibilities:
- Save identity index for fast load list.
- Save read/write with atomic commit strategy (write temp + fsync + rename).
- Schema versions and migration pipeline.
- Corruption policy: detect, quarantine/skip bad save, return actionable warning.

Rules:
- No UI dependencies.
- May serialize domain DTOs; never pushes storage concerns back into domain.

### F) Config / Balance Layer (`internal/config`)
Purpose: versioned economy and balancing constants.

Responsibilities:
- Load config from local file path(s) with embedded defaults fallback.
- Validate version/schema.
- Provide deterministic constants and multiplier tables to sim/domain consumers.

Rules:
- Config snapshots are immutable during a simulation session.

### G) Logging / Diagnostics (`internal/diag`)
Purpose: local-only observability.

Responsibilities:
- Optional structured logs to local file.
- Panic/error diagnostics without telemetry.
- No network exporters.

## 3) Dependency Rules (Explicit)

### Allowed dependencies
- `internal/ui` -> reads/uses `internal/app` contracts and app-provided view state; may reference domain/sim/persistence DTOs via app boundary types.
- `internal/app` -> depends on `ui`, `domain`, `sim`, `persistence`, `config`, `diag` to wire runtime.
- `internal/sim` -> depends on `internal/domain`.
- `internal/persistence` -> may depend on domain serialization types.
- `internal/domain` -> stdlib only.

### Forbidden dependencies
- `internal/domain` -> `ui` / `app` / `persistence` / `sim`.
- `internal/sim` -> `ui`.
- `internal/persistence` -> `ui`.

### Interface placement and circular-dependency prevention
- Consumer-side interface rule:
  - `app` defines interfaces it consumes (e.g., `SaveRepository`, `TickEngine`, `BalanceProvider`, `Clock`, `RNG`).
  - Concrete implementations live in provider packages (`persistence`, `sim`, `config`, `platform`).
- Domain interfaces live in domain only when they represent domain concepts, not infrastructure concerns.
- Use narrow interfaces per use-case to avoid import cycles and oversized contracts.
- Prefer DTO boundaries between UI and app to keep Bubble Tea models decoupled from persistence structs.

## 4) Entry UX State Machine (`tbiz`, interactive-only)

### State model
- `BootMenuState`
- `NewGameFormState`
- `LoadGameListState`
- `LoadGameState` (transitional loading state)
- `DashboardState`
- `HelpState`
- `ErrorModalState` (overlay, non-terminal)
- `ExitConfirmState` (overlay)

### ASCII transition diagram
```text
[BootMenuState]
  New Game  -> [NewGameFormState]
  Load Game -> [LoadGameListState]
  Help      -> [HelpState]
  Exit      -> [ExitConfirmState] --confirm--> (process exit 0)
                                     --cancel--> [BootMenuState]

[HelpState]
  Esc/q -> [BootMenuState]

[NewGameFormState]
  Submit(valid) -> [LoadGameState:create] -> [DashboardState]
  Submit(error) -> [ErrorModalState] -> [NewGameFormState]
  Cancel        -> [BootMenuState]

[LoadGameListState]
  Select Save -> [LoadGameState:existing] -> [DashboardState]
  Esc/q       -> [BootMenuState]
  Empty + N   -> [NewGameFormState]
  Corrupt row skipped -> toast/warning (remain in list)

[DashboardState]
  in-game navigation: Store/HR/Investors/Inbox/Reports/Settings
  Save action (UI) -> persist -> success toast/failure modal
  Back-to-menu action (UI path) -> [BootMenuState] (after save policy handling)
```

### Error handling transitions
- Failed create save in `NewGameFormState`:
  - Stay in form, show inline/form-level error modal; preserve inputs.
- Failed load in `LoadGameState`:
  - Route to `ErrorModalState` with actions `Back to List` / `Back to Menu`.
- Corrupted save in list:
  - Exclude from selectable rows, show warning count/toast, optional recovery panel route.
- Panic in UI update/render:
  - recover in app shell, log locally, show safe error state, allow return to boot menu.

### Deterministic seed handling
- New game:
  - Generate and persist `SimulationSeed` at save creation.
- Load game:
  - Restore seed/RNG cursor from save before first tick.
- Runtime:
  - Seed evolves only through deterministic rng calls inside tick engine.

## 5) Testing Strategy + Coverage Guarantees

### Test layout conventions
- Prefer co-located package tests (`*_test.go`) in each package.
- Use `internal/<pkg>/testdata/` for fixtures/golden files.
- Use black-box tests (`package foo_test`) for public contract tests, white-box (`package foo`) when internal behavior needs direct validation.

### Domain tests (mandatory 100%)
- Invariant/property-style tests:
  - equity totals, cash underflow behavior, runway math, idempotent recompute.
- Edge-case tables:
  - zero burn, extreme modifiers, min/max clamp boundaries.

### Simulation tests (mandatory 100%)
- Seeded deterministic replay tests:
  - same save + seed + ticks + config => same final state.
- Multi-tick progression tests:
  - fixed scenario snapshots across N ticks.
- Tick-order tests:
  - verify authoritative operation ordering.

### Persistence tests
- Atomic write tests using temp dirs and fault injection.
- Save identity index ordering tests (`LastPlayedAt DESC`, tie `SaveID ASC`).
- Schema version + migration tests.
- Corruption detection/skip behavior tests.

### UI tests
- State transition tests for each screen model.
- Key routing tests (`arrows/jk`, `enter`, `esc/q`).
- View helper tests for stable rendering primitives (golden-friendly where valuable).

### CI coverage enforcement plan
- CI workflow gates:
  - `go test ./...`
  - race-enabled tests for non-UI packages where practical.
  - coverage report generation.
- Hard threshold gates:
  - `internal/domain`: 100.0%
  - `internal/sim`: 100.0%
- Suggested enforcement mechanism:
  - `scripts/coverage.sh` parses `go test -coverprofile` and fails if thresholds are below target.

## 6) Roadmap Milestones

### MVP
- Boot menu + new/load/help/exit flows.
- Save identity index + local persistence with atomic writes.
- Basic dashboard shell.
- Minimal domain + deterministic tick engine.
- Store MVP items: desk/chair/computer.
- HR basics: hire/list payroll impact.
- 100% coverage for `domain` + `sim`.

### v1
- Investors + term sheet flow + cap table reporting.
- Expanded store tools/effects.
- Deeper event system + inbox timeline.
- Better reporting (tables, trends, text sparklines).

### v2
- Departments and advanced finance systems.
- Long-term strategy layers.
- Optional mod/plugin architecture.
- Balance packs / scenario mode.

## 7) Non-Functional Commitments

- Offline/local-first only.
- Minimal dependencies (stdlib-first, selective third-party use).
- Determinism is non-negotiable in sim/domain behavior.
- Domain remains pure and infrastructure-agnostic.
