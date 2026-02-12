# terminal-business UX Contract (`tbiz`)

## 1. CLI + Boot Contract

### 1.1 Binary behavior (strict)
- The executable name is `tbiz`.
- `tbiz` has no flags and no subcommands.
- Running `tbiz` always launches an interactive full-screen Bubble Tea UI.
- Any extra CLI tokens (example: `tbiz foo`) are treated as invalid invocation and immediately route to a full-screen in-app notice:
  - Message: `tbiz does not support flags or subcommands. Press any key to continue.`
  - After keypress, continue to main menu.
- No networking required; startup must work offline.

### 1.2 Boot/Main Menu items (exact order)
1. New Game
2. Load Game
3. Help
4. Exit

### 1.3 Main menu layout and premium styling
- Full-screen centered composition, responsive to terminal resize.
- Vertical structure:
  1. Brand/header row: `TERMINAL BUSINESS` with subtle divider line.
  2. Short tagline row: `Build from zero to unicorn.`
  3. Menu block (focus area).
  4. Footer hint row for keybindings.
- Visual style principles:
  - High contrast, minimal clutter, strong spacing rhythm.
  - Selected item uses reverse/highlight style + left accent marker (`›`).
  - Non-selected items are muted but readable.
  - Use one accent color consistently for focus states and confirmation actions.
  - Avoid noisy borders; prefer clean padding and thin separators.
- Cross-terminal safety:
  - Must render legibly in 80x24 minimum.
  - No dependency on unsupported glyphs; provide ASCII-safe fallback for symbols.

### 1.4 Main menu interaction
- Navigation:
  - `Up` / `k`: move selection up.
  - `Down` / `j`: move selection down.
  - Selection wraps (last -> first, first -> last).
- Activation:
  - `Enter`: activate selected item.
- Exit shortcuts:
  - `Esc` or `q` from main menu opens Exit confirm dialog (not immediate quit).

### 1.5 Focus and selection behavior
- Initial focus on `New Game`.
- Exactly one item is always selected in menu context.
- On returning to main menu from child screens:
  - Preserve last selected menu item for continuity.
- Terminal resize:
  - Keep current selection/focus unchanged.
  - Reflow layout without state loss.

### 1.6 Confirm dialogs
- Required confirms:
  - Exiting app from main menu (`Exit`, `Esc`, or `q`).
- Dialog style:
  - Centered modal layer with dimmed background.
  - Title: `Confirm Exit`
  - Body: `Leave Terminal Business?`
  - Actions: `Cancel` (default focus), `Exit`
- Keys in confirm:
  - `Left/Right` or `h/l` switch action focus.
  - `Enter` confirms focused action.
  - `Esc` closes dialog (same as Cancel).

### 1.7 Error states and recovery (boot/menu)
- Save index read failure at startup:
  - Show non-blocking warning banner in menu footer: `Could not read saves. You can still start a new game.`
  - Keep app usable; retry load when entering `Load Game`.
- Unknown runtime error:
  - Show full-screen recoverable error panel with actions:
    - `Back to Menu` (default)
    - `Exit`

---

## 2. “New Game” UX Contract

### 2.1 Entry
- Trigger: selecting `New Game` from main menu.
- Opens full-screen form view (not modal).

### 2.2 Form fields (required)
1. Company Name (text input)
2. Company Type (single-select enum):
   - Game
   - Fintech
   - SaaS
   - (Designed to accept additional types later)

### 2.3 Form layout
- Centered form card with:
  - Title: `Create New Company`
  - Subtitle: `Set your startup identity.`
  - Field stack with clear labels and helper/error lines.
  - Action row at bottom: `Create Company` (primary), `Cancel` (secondary hint via `Esc`).
- Field focus order:
  1. Company Name
  2. Company Type
  3. Submit action
- Default focus: Company Name.

### 2.4 Company Name validation
- Required: non-empty after trim.
- Length: 2-40 characters.
- Allowed characters:
  - Letters, numbers, spaces, hyphen (`-`), apostrophe (`'`), ampersand (`&`).
- Disallowed:
  - Leading/trailing punctuation-only names.
  - Control characters.
- Inline validation behavior:
  - Validate on blur and on submit.
  - Show inline error directly below field.
  - Error examples:
    - `Company name is required.`
    - `Use 2-40 characters.`
    - `Only letters, numbers, spaces, -, ', & are allowed.`

### 2.5 Company Type selection UX
- Render as radio-style vertical list.
- Default selected type: `SaaS`.
- Keys while type field focused:
  - `Up/Down` or `k/j` to change selected type.
  - `Tab` moves to next focusable control.
- Only one type can be selected.

### 2.6 Submission behavior
- `Enter` on submit or from last field triggers validation.
- If valid:
  1. Create new save record with stable unique ID.
  2. Persist locally.
  3. Immediately load created save into main Dashboard.
- Show transient non-blocking status during write/load:
  - `Creating company...`

### 2.7 Cancellation behavior
- `Esc` at any point in New Game view returns to main menu.
- No save file is created on cancel.
- Optional guard (recommended): if any field changed, show confirm:
  - `Discard new company setup?` -> `Keep Editing` / `Discard`

### 2.8 New Game error handling
- Persistence write failure:
  - Keep user on form.
  - Show inline form-level error: `Could not create save locally. Try again.`
  - Preserve entered values.
- Post-create load failure:
  - Show error panel with:
    - `Retry Load`
    - `Back to Menu`
  - Save should remain on disk if creation succeeded.

---

## 3. “Load Game” UX Contract

### 3.1 Entry
- Trigger: selecting `Load Game` from main menu.
- Opens full-screen save list view.

### 3.2 Save row display format (exact)
- Primary text for every row must be:
  - `{Company Name} - {Company Type}`

### 3.3 Save list sorting and determinism
- Primary sort (recommended): `last_played_at` descending (most recent first).
- Secondary stable tie-breaker: save ID ascending.
- Sorting must be deterministic across runs.

### 3.4 Save list layout and metadata
- Header: `Load Game`
- Body: scrollable vertical list with one selected row.
- Footer: key hints (`Enter: Load`, `Esc: Back`).
- Optional subtle metadata (no clutter):
  - Right-aligned muted timestamp, e.g., `Last played: 2026-02-12 14:05`.
  - Must not alter required primary format string.

### 3.5 Interaction
- `Up/Down` or `k/j`: move selection.
- Selection wraps for consistent navigation.
- `Enter`: load selected save and enter Dashboard.
- `Esc` or `q`: return to main menu.

### 3.6 Empty state
- If no valid saves found:
  - Show centered message: `No saves found.`
  - Show shortcuts:
    - `N`: Create New Game
    - `Esc`: Back to Menu
- `N` routes directly into New Game form.

### 3.7 Corrupted save handling
- On list build:
  - If a save is corrupted/unreadable, skip it from selectable list.
  - Show non-blocking warning count in footer: `1 corrupted save skipped.`
- Provide recovery entry point:
  - Key `R` opens recovery details panel listing skipped save IDs/filenames.
  - Recovery panel actions: `Back` only (MVP).
- If selected save fails at load time:
  - Show modal: `Save could not be loaded.`
  - Actions: `Back to List` (default), `Back to Menu`

---

## 4. Global Keybindings and Interaction Rules

### 4.1 Global conventions
- `Up/Down` and `k/j` are equivalent where list navigation exists.
- `Enter` activates current focused control.
- `Esc` means back/cancel in child views and confirm-exit at root menu.
- `q` mirrors `Esc` for quick back/exit behavior.

### 4.2 Focus model
- Exactly one interactive context is active at a time:
  - View content, modal dialog, or error panel.
- When modal is open, background view is inert.
- Focus return rule:
  - Closing modal restores prior focus target.

### 4.3 Feedback and latency
- Any operation over ~100ms should show a status indicator.
- Success transitions should feel immediate and avoid unnecessary confirmation screens.
- Errors should be actionable with clear next step (retry/back).

---

## 5. Save Identity and Listing Expectations (UX-level)

### 5.1 Local-first persistence
- Saves are stored locally and fully usable offline.
- No cloud dependency, no login, no network calls.

### 5.2 Save identity model
- Each save has:
  - Stable unique ID (non-user-editable, persistent across sessions).
  - Company Name (user-facing).
  - Company Type (enum value).
- Human-friendly display name in lists:
  - `{Company Name} - {Company Type}`

### 5.3 Listing performance and determinism
- Load Game screen should open quickly even with many saves.
- Save list order must be stable and deterministic for identical metadata.
- Corrupted records must not break list rendering or crash UI.

### 5.4 Minimum metadata to support UX
- `id`
- `company_name`
- `company_type`
- `created_at`
- `last_played_at`
- `schema_version` (for future migration/recovery messaging)

---

## 6. Help Screen (Boot Menu dependency)
- Since `Help` is a top-level menu item, include a lightweight full-screen help panel with:
  - Core controls (`arrows/jk`, `enter`, `esc/q`).
  - Quick explanation of New vs Load flow.
  - `Esc` or `q` returns to main menu.
- Help is informational only, no side effects.

