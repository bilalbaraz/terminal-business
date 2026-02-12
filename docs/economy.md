# terminal-business Domain + Economy Design

## Scope and Alignment
- Binary and UX alignment: `tbiz` only, fully interactive TUI, no flags/subcommands.
- Flow alignment: Boot Menu -> New Game -> Save Created -> Dashboard; Load Game -> Select Save -> Dashboard.
- This document defines domain models and deterministic simulation behavior only.
- Domain and simulation are pure: no IO, no terminal/UI concerns, no wall-clock access.

## 1) Core Domain Entities

### Company
Purpose: Persistent startup identity and lifecycle anchors.

Fields:
- `ID`: stable unique identifier (UUID string or equivalent).
- `Name`: user-facing company name.
- `Type`: enum/string key (`game`, `fintech`, `saas`), extensible.
- `CreatedAt`: simulation timestamp when company was created.
- `LastPlayedAt`: simulation timestamp updated on session load/save.

Rules:
- `ID` immutable after creation.
- `Name` and `Type` must be valid at creation; `Type` must resolve in company-type registry/config.

### CompanyMetrics
Purpose: Current economic + health snapshot derived from state and config.

Fields:
- `Cash`
- `BurnRate`
- `Revenue`
- `RunwayMonths`
- `Valuation`
- `Reputation`
- `Productivity`
- `Morale`
- `TechDebt`

Rules:
- Numeric domains and bounds are defined in invariants.
- Recomputed metrics replace prior computed values (no accumulation side effects unless formula defines it).

### InventoryItem
Purpose: Owned equipment/tools/upgrades that affect outcomes.

Fields:
- `ID`
- `Name`
- `Category` (extensible enum/string key)
- `Effects` (structured modifiers, e.g. productivity/morale/revenue maintenance cost)

Rules:
- `ID` immutable.
- `Effects` are additive/multiplicative only through declared effect schema.

### StoreCatalog
Purpose: Source of purchasable items and pricing behavior.

Fields:
- `Items[]` (catalog entries)
- `PricingRules` (base prices, scaling curves, discount/markup policies)

Rules:
- Catalog entries are immutable within a config version.
- Pricing is deterministic for same inputs and config version.

### Employee
Purpose: Workforce units that cost payroll and modify productivity/morale.

Fields:
- `ID`
- `Role` (extensible enum/string key)
- `Salary`
- `ProductivityModifier`
- `MoraleModifier`

Rules:
- `Salary >= 0`.
- Modifiers constrained to configured safe ranges.

### Payroll
Purpose: Deterministic salary obligations.

Fields:
- `PayInterval` (e.g. monthly; enum/config key)
- `TotalPayrollCost` (derived from active employees and interval policy)

Rules:
- `TotalPayrollCost` is computed, not manually edited.
- Calculation uses ordered employee set (stable ordering by employee ID).

### Investor
Purpose: Defines investor behavior templates for funding events.

Fields:
- `Name`
- `RiskProfile` (e.g. conservative, balanced, aggressive)
- `CheckSizeRange` (min/max)

Rules:
- Range must be valid (`min <= max`, both non-negative).

### TermSheet
Purpose: Proposed investment offer.

Fields:
- `InvestmentAmount`
- `Valuation`
- `EquityOffered`
- `Conditions` (optional, future-compatible structured list)

Rules:
- `InvestmentAmount > 0`, `Valuation > 0`.
- `EquityOffered` must produce valid post-deal cap table (see invariants).

### CapTable
Purpose: Ownership split model.

Fields:
- `FoundersEquity`
- `InvestorEquity`
- `OptionPoolEquity` (future; default `0` in MVP)

Rules:
- Always normalized to 100% total ownership.
- No negative equity slices.

### Event
Purpose: Deterministic game events that modify state.

Fields:
- `ID`
- `TriggerConditions` (state predicates + optional RNG thresholds)
- `Effects` (state deltas/modifiers)

Rules:
- Trigger evaluation pure and deterministic with supplied RNG.
- Effects must be side-effect free and composable.

---

## 2) Domain Invariants (Mandatory)

1. Equity sum must always equal 100%.
- Invariant: `FoundersEquity + InvestorEquity + OptionPoolEquity == 100.00%` (within configured precision/rounding policy).
- Violation handling: reject state transition.

2. Cash cannot silently go negative.
- Invariant: any transition that would move `Cash < 0` must emit explicit deficit state/result.
- Allowed strategies (choose one in implementation):
  - hard-fail transition with domain error, or
  - enter explicit `Insolvent` status with negative cash recorded intentionally.
- Silent underflow is forbidden.

3. Payroll must be deterministic.
- Invariant: same employee set + same salaries + same interval + same config => identical payroll output.
- Stable ordering must be enforced before aggregation.

4. Runway must be mathematically correct.
- Invariant:
  - if `BurnRate > 0`: `RunwayMonths = Cash / BurnRate`
  - if `BurnRate == 0`: `RunwayMonths = +Inf` or configured sentinel.
  - if explicit insolvency is modeled and `Cash < 0`: runway is `0` (or insolvency sentinel per config).

5. Metrics recompute must be idempotent.
- Invariant: applying recompute twice without state changes yields identical metrics.
- No hidden mutation in read/recompute paths.

6. Save identity must never change once created.
- Invariant: `SaveID`, `CompanyID`, `CreatedAt`, and original creation `Version` remain immutable.
- `CompanyName` can change only through explicit rename flow (future), not during load/recompute.

---

## 3) Save Identity Model

### SaveIdentity schema
- `SaveID` (UUID)
- `CompanyID`
- `CompanyName`
- `CompanyType`
- `CreatedAt`
- `LastPlayedAt`
- `Version`

### Rules
- `SaveID` is internal-only and never shown in primary UI list rows.
- Load Game display string is exactly:
  - `"{Company Name} - {Company Type}"`
- `LastPlayedAt` must update on successful load and on successful save commit.
- Save listing sort order: `LastPlayedAt DESC`.
- Tie-breaker for deterministic ordering: `SaveID ASC`.
- Save identity object must be decodable without loading full simulation state (fast list rendering).

---

## 4) Economy System (MVP)

All formulas are deterministic, config-driven, and pure.

### Inputs
- State inputs: employees, owned inventory, events, company metrics, cap table.
- Config inputs: constants, multipliers, penalties, bounds, rounding policy.

### Formula definitions
1. Burn Rate
- `BurnRate = BaseBurn + Payroll + ToolMaintenance`
- Where:
  - `BaseBurn`: config baseline by company type/stage
  - `Payroll`: deterministic payroll output
  - `ToolMaintenance`: sum of recurring maintenance from inventory/equipment

2. Revenue
- `Revenue = BaseRevenue * Productivity * ReputationModifier`
- Where:
  - `ReputationModifier` typically derived as function of reputation score (config curve/table)

3. Productivity
- `Productivity = BaseProductivity + EquipmentBonus + EmployeeBonus - TechDebtPenalty`
- Clamp to configured min/max bounds.

4. Morale
- `Morale = BaseMorale + OfficeBonus + EventBonus - OverworkPenalty`
- Clamp to configured min/max bounds.

5. Runway
- `RunwayMonths = Cash / BurnRate` (see invariant edge cases)

6. Valuation
- `Valuation = Revenue * ValuationMultiplier`
- Multiplier from config and may depend on company type/reputation band.

### Formula properties
- Deterministic: no hidden randomness, no wall-clock calls.
- Config-driven: constants and curves loaded from balance config.
- Isolated: each formula callable independently for unit/property tests.

---

## 5) Simulation Tick Order

Tick inputs:
- Prior state snapshot
- Config snapshot version
- RNG interface + current seed state
- Current simulation timestamp (provided by scheduler, not wall-clock)

Tick order (authoritative):
1. Payroll processing
- Compute payroll due for interval.
- Apply cash delta with explicit insolvency handling.

2. Revenue update
- Compute revenue and apply cash increase.

3. Morale update
- Recompute morale from current workload/events modifiers.

4. Event trigger evaluation
- Evaluate triggers using state + injectable RNG.
- Apply event effects in stable order (`Event.ID ASC`).

5. Metrics recomputation
- Recompute all derived metrics once using post-event state.

6. LastPlayedAt update (if session active)
- Update save/session metadata timestamp from simulation clock input.

Determinism requirements:
- Fixed order is mandatory.
- No concurrent mutation during tick.
- Same pre-state/config/RNG state => same post-state.

---

## 6) Deterministic Simulation Guarantees

1. RNG is injectable.
- Domain receives RNG via interface dependency, never via globals.

2. Seed is stored in save.
- Save state includes RNG seed (and cursor/state as needed by RNG algorithm).

3. No `time.Now()` inside domain/simulation.
- Time is an explicit input parameter from outer application layer.

4. No global RNG usage.
- No package-level random source in domain/simulation packages.

5. Reproducibility contract.
- Simulation output is identical for:
  - same save snapshot
  - same seed/RNG state
  - same tick count
  - same config version

---

## 7) Balance Configuration System

### Goals
- Externalized balancing without code changes.
- Safe evolution across versions.

### Required config domains
- Economy constants (`BaseBurn`, `BaseRevenue`, etc.)
- Multipliers (valuation, reputation curves)
- Event probabilities/weights and trigger params
- Store pricing and maintenance costs
- Bounds/clamps and rounding precision policy

### Versioning and compatibility
- Config must include explicit `ConfigVersion`.
- Saves persist `ConfigVersion` used at last commit.
- Backward compatibility approach:
  - Prefer additive fields with defaults.
  - Provide migration adapters for breaking changes.
  - If incompatible, fail with explicit, user-safe error and fallback instructions.

### Determinism constraints
- Config load order must be deterministic.
- Missing values resolved via deterministic defaults.
- Floating-point usage must follow fixed precision/rounding policy to avoid drift.

---

## 8) Extensibility Requirements

Model design must support expansion without core rewrites:
- New company types via registry/config entries, not hard-coded switches.
- New store categories and effect channels through typed effect schema.
- New employee traits/modifiers via trait map with bounded influence.
- Advanced finance systems (debt, liquidation preference, SAFEs) by extending TermSheet/CapTable modules.
- Department systems by grouping employees into teams with local modifiers feeding global metrics.

Extensibility rules:
- Unknown enum keys should fail validation early, not panic later.
- New optional fields must preserve behavior for old saves/configs.

---

## 9) Testing Expectations (Design Commitments)

### Required test categories
1. Property-style invariant tests
- Equity normalization, non-silent negative cash handling, idempotent recompute.

2. Deterministic replay tests
- Same save + seed + tick count + config => byte-equivalent or semantically equivalent state.

3. Multi-tick progression tests
- Validate stable outcomes across long tick sequences.

4. Save/load determinism tests
- Serialize -> deserialize -> simulate should match uninterrupted simulate path.

### Coverage target (implementation requirement)
- `domain` + `simulation` packages target: 100% test coverage.

---

## 10) Purity and Dependency Boundaries

- Domain package responsibilities:
  - Entity definitions, invariants, pure calculations, state transitions.
- Simulation package responsibilities:
  - Tick orchestration, deterministic event application, RNG consumption via interface.
- Infrastructure/UI responsibilities (outside domain):
  - File IO, save storage, config file loading, Bubble Tea rendering, user input.

Hard rule:
- No IO in domain/simulation packages.
- Offline/local-first operation is guaranteed by infrastructure layer using local storage only.
