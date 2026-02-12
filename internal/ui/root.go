package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"terminal-business/internal/app"
	domain "terminal-business/internal/domain/store"
	"terminal-business/internal/persistence"
	"terminal-business/internal/sim"
	"terminal-business/internal/ui/components"
	"terminal-business/internal/ui/screens/boot"
	"terminal-business/internal/ui/screens/dashboard"
	"terminal-business/internal/ui/screens/fatal"
	"terminal-business/internal/ui/screens/help"
	"terminal-business/internal/ui/screens/loadgame"
	"terminal-business/internal/ui/screens/newgame"

	tea "github.com/charmbracelet/bubbletea"
)

type createResultMsg struct {
	save      persistence.SaveFile
	gameState domain.GameState
	err       error
}

type loadResultMsg struct {
	save      persistence.SaveFile
	gameState domain.GameState
	err       error
}

type snapshotPayload struct {
	Day              int                            `json:"day"`
	Cash             int                            `json:"cash"`
	Headcount        int                            `json:"headcount"`
	PlayerInventory  []domain.InventoryItemInstance `json:"player_inventory"`
	CompanyInventory []domain.InventoryItemInstance `json:"company_inventory"`
	ActiveJobs       []domain.ActiveJob             `json:"active_jobs"`
	CompletedJobs    []domain.CompletedJob          `json:"completed_jobs"`
}

type Model struct {
	machine *app.Machine
	store   persistence.Store
	clock   app.Clock
	rng     app.RNG

	catalog      domain.Catalog
	economy      domain.EconomyConfig
	bootstrap    domain.BootstrapConfig
	gameState    domain.GameState
	simSeed      int64
	marketTick   int
	marketOffers []domain.JobOffer

	boot      boot.Model
	newGame   newgame.Model
	loadGame  loadgame.Model
	help      help.Model
	dashboard dashboard.Model
	fatal     fatal.Model

	spinner      components.Spinner
	toast        components.Toast
	recoverModal components.Modal

	loadingFrom app.State
}

func NewModel(store persistence.Store, clk app.Clock, rng app.RNG, entries []persistence.SaveIndexEntry) *Model {
	catalog := domain.DefaultCatalog()
	economy := domain.DefaultEconomyConfig()
	bootstrap := domain.DefaultBootstrapConfig(catalog)
	if err := domain.ValidateBootstrapConfig(catalog, bootstrap); err != nil {
		panic(err)
	}
	game, err := domain.BootstrapState(catalog, economy, bootstrap)
	if err != nil {
		panic(err)
	}
	dash := dashboard.New()
	dash.SetStoreCatalog(catalog)
	dash.SetGameState(game)

	m := &Model{
		machine:      app.NewMachine(),
		store:        store,
		clock:        clk,
		rng:          rng,
		catalog:      catalog,
		economy:      economy,
		bootstrap:    bootstrap,
		gameState:    game,
		simSeed:      0,
		marketTick:   -1,
		marketOffers: nil,
		boot:         boot.New(),
		newGame:      newgame.New(),
		loadGame:     loadgame.New(entries),
		help:         help.New(),
		dashboard:    dash,
		fatal:        fatal.New(""),
		spinner:      components.NewSpinner(),
		recoverModal: components.Modal{Actions: []string{"Back to Menu"}},
		loadingFrom:  app.BootMenuState,
	}
	m.refreshMarketOffers()
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	defer func() {
		if r := recover(); r != nil {
			_ = m.machine.Transition(app.EventShowFatalError)
			m.fatal = fatal.New(fmt.Sprintf("Recovered panic: %v", r))
		}
	}()

	if m.recoverModal.Open {
		return m, m.handleModal(msg)
	}

	m.spinner.Update(msg)

	switch m.machine.State() {
	case app.BootMenuState:
		return m.handleBoot(msg)
	case app.NewGameFormState:
		return m.handleNewGame(msg)
	case app.LoadGameListState:
		return m.handleLoadGame(msg)
	case app.HelpState:
		return m.handleHelp(msg)
	case app.LoadingGameState:
		return m.handleLoading(msg)
	case app.DashboardState:
		return m.handleDashboard(msg)
	case app.FatalErrorState:
		return m.handleFatal(msg)
	default:
		return m, nil
	}
}

func (m *Model) handleBoot(msg tea.Msg) (tea.Model, tea.Cmd) {
	action := m.boot.Update(msg)
	switch action {
	case boot.ActionNewGame:
		_ = m.machine.Transition(app.EventGoNewGame)
	case boot.ActionLoadGame:
		_ = m.machine.Transition(app.EventGoLoadGame)
		entries, err := m.store.LoadIndex(context.Background())
		if err != nil {
			m.toast.Show("Could not read saves. Starting with empty list.")
			entries = nil
		}
		m.loadGame.SetEntries(entries)
	case boot.ActionHelp:
		_ = m.machine.Transition(app.EventGoHelp)
	case boot.ActionExit:
		_ = m.machine.Transition(app.EventExit)
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) handleNewGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	action := m.newGame.Update(msg)
	switch action {
	case newgame.ActionCancel:
		_ = m.machine.Transition(app.EventBack)
	case newgame.ActionSubmit:
		_ = m.machine.Transition(app.EventSubmitNewGame)
		m.loadingFrom = app.NewGameFormState
		return m, m.createSaveCmd(m.newGame.CompanyName, m.newGame.SelectedType())
	}
	return m, nil
}

func (m *Model) handleLoadGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	action := m.loadGame.Update(msg)
	switch action.Type {
	case loadgame.ActionCancel:
		_ = m.machine.Transition(app.EventBack)
	case loadgame.ActionNewGame:
		_ = m.machine.Transition(app.EventGoNewGame)
	case loadgame.ActionSelect:
		_ = m.machine.Transition(app.EventSelectSave)
		m.loadingFrom = app.LoadGameListState
		return m, m.loadSaveCmd(m.loadGame.SelectedSaveID())
	case loadgame.ActionDelete:
		if err := m.store.DeleteSave(context.Background(), action.SaveID); err != nil {
			m.loadGame.SetError("Delete failed. Rebuilding index...")
			_, _ = m.store.LoadIndex(context.Background())
		} else {
			m.toast.Show("Deleted " + action.Label)
			m.loadGame.SetError("")
		}
		entries, err := m.store.LoadIndex(context.Background())
		if err != nil {
			m.loadGame.SetError("Could not refresh saves list.")
			entries = nil
		}
		m.loadGame.SetEntries(entries)
	}
	return m, nil
}

func (m *Model) handleHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.help.Update(msg) == help.ActionBack {
		_ = m.machine.Transition(app.EventBack)
	}
	return m, nil
}

func (m *Model) handleLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case createResultMsg:
		if msg.err != nil {
			m.openRecoverModal(msg.err.Error(), []string{"Back to Menu"})
			return m, nil
		}
		m.gameState = msg.gameState
		m.simSeed = msg.save.SimulationSeed
		m.refreshMarketOffers()
		m.dashboard.SetCompany(msg.save.SaveIdentity)
		m.dashboard.SetGameState(msg.gameState)
		m.dashboard.SetStoreError("")
		m.dashboard.SetMarketJobs(m.marketOffers)
		m.dashboard.SetMarketError("")
		_ = m.machine.Transition(app.EventLoadSucceeded)
		m.toast.Show("Company created and loaded")
		return m, nil
	case loadResultMsg:
		if msg.err != nil {
			actions := []string{"Back to Menu"}
			if m.loadingFrom == app.LoadGameListState {
				actions = []string{"Back to List", "Back to Menu"}
			}
			m.openRecoverModal(msg.err.Error(), actions)
			return m, nil
		}
		m.gameState = msg.gameState
		m.simSeed = msg.save.SimulationSeed
		m.refreshMarketOffers()
		m.dashboard.SetCompany(msg.save.SaveIdentity)
		m.dashboard.SetGameState(msg.gameState)
		m.dashboard.SetStoreError("")
		m.dashboard.SetMarketJobs(m.marketOffers)
		m.dashboard.SetMarketError("")
		_ = m.machine.Transition(app.EventLoadSucceeded)
		m.toast.Show("Save loaded")
		return m, nil
	default:
		return m, m.spinner.Tick()
	}
}

func (m *Model) handleDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	action := m.dashboard.Update(msg)
	switch action.Type {
	case dashboard.ActionBack:
		_ = m.machine.Transition(app.EventBack)
	case dashboard.ActionBuy:
		next, err := domain.ApplyPurchase(m.gameState, m.catalog, m.economy, action.ItemID)
		if err != nil {
			if errors.Is(err, domain.ErrInsufficientFunds) {
				m.dashboard.SetStoreError("Not enough cash to buy this item.")
				m.toast.Show("Purchase failed: insufficient funds")
				return m, nil
			}
			if errors.Is(err, domain.ErrMaxOwned) {
				m.dashboard.SetStoreError("You already own the maximum quantity.")
				m.toast.Show("Purchase failed: max owned")
				return m, nil
			}
			m.dashboard.SetStoreError("Purchase failed.")
			m.toast.Show("Purchase failed")
			return m, nil
		}
		m.gameState = next
		m.dashboard.SetGameState(next)
		m.dashboard.SetStoreError("")
		item, _ := m.catalog.Item(action.ItemID)
		m.toast.Show("Purchased " + item.DisplayName)
	case dashboard.ActionAcceptJob:
		next, err := domain.AcceptJob(m.gameState, m.marketOffers, action.JobID, "founder", m.gameState.Day)
		if err != nil {
			if errors.Is(err, domain.ErrNotOperational) {
				m.dashboard.SetMarketError("You need Desk + Chair + Computer to take jobs.")
				m.toast.Show("Market blocked: not operational")
				return m, nil
			}
			if errors.Is(err, domain.ErrCapacityFull) || errors.Is(err, domain.ErrAssigneeBusy) {
				m.dashboard.SetMarketError("Capacity full: one active job per person.")
				m.toast.Show("Market blocked: capacity full")
				return m, nil
			}
			m.dashboard.SetMarketError("Could not accept job.")
			m.toast.Show("Job accept failed")
			return m, nil
		}
		m.gameState = next
		m.dashboard.SetGameState(next)
		m.dashboard.SetMarketError("")
		m.toast.Show("Job accepted")
	case dashboard.ActionAdvanceDay:
		m.gameState = sim.AdvanceDay(m.gameState, m.catalog, m.economy)
		m.refreshMarketOffers()
		m.dashboard.SetGameState(m.gameState)
		m.dashboard.SetMarketJobs(m.marketOffers)
		m.toast.Show("Advanced to next day")
	}
	return m, nil
}

func (m *Model) handleFatal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.fatal.Update(msg) {
	case fatal.ActionBack:
		_ = m.machine.Transition(app.EventRecoverToBoot)
	case fatal.ActionQuit:
		_ = m.machine.Transition(app.EventExit)
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) openRecoverModal(message string, actions []string) {
	m.recoverModal = components.Modal{
		Open:          true,
		Title:         "Error",
		Body:          message,
		Actions:       actions,
		SelectedIndex: 0,
	}
}

func (m *Model) handleModal(msg tea.Msg) tea.Cmd {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	switch k.String() {
	case "left", "h":
		m.recoverModal.Left()
	case "right", "l":
		m.recoverModal.Right()
	case "esc":
		m.recoverModal.Open = false
		if m.loadingFrom == app.LoadGameListState {
			_ = m.machine.Transition(app.EventRecoverToLoad)
		} else {
			_ = m.machine.Transition(app.EventRecoverToBoot)
		}
	case "enter":
		selected := m.recoverModal.Selected()
		m.recoverModal.Open = false
		if selected == "Back to List" {
			_ = m.machine.Transition(app.EventRecoverToLoad)
		} else {
			_ = m.machine.Transition(app.EventRecoverToBoot)
		}
	}
	return nil
}

func (m *Model) createSaveCmd(name, companyType string) tea.Cmd {
	return func() tea.Msg {
		now := m.clock.Now()
		initialState, err := domain.BootstrapState(m.catalog, m.economy, m.bootstrap)
		if err != nil {
			return createResultMsg{err: err}
		}
		seed := m.rng.Int63()
		save := persistence.SaveFile{
			SaveIdentity: persistence.SaveIdentity{
				CompanyName:  name,
				CompanyType:  companyType,
				CreatedAt:    now,
				LastPlayedAt: now,
				Version:      1,
			},
			SimulationSeed:      seed,
			TickCounter:         int64(initialState.Day),
			DomainStateSnapshot: encodeSnapshot(initialState),
			Version:             1,
		}
		created, err := m.store.CreateSave(context.Background(), save)
		if err != nil {
			return createResultMsg{err: err}
		}
		if err := m.store.TouchSave(context.Background(), created.SaveIdentity.SaveID, now); err != nil {
			return createResultMsg{err: err}
		}
		loaded, err := m.store.LoadSave(context.Background(), created.SaveIdentity.SaveID)
		if err != nil {
			return createResultMsg{err: err}
		}
		state, err := decodeSnapshot(loaded.DomainStateSnapshot, m.catalog, m.economy)
		return createResultMsg{save: loaded, gameState: state, err: err}
	}
}

func (m *Model) loadSaveCmd(saveID string) tea.Cmd {
	return func() tea.Msg {
		now := m.clock.Now()
		loaded, err := m.store.LoadSave(context.Background(), saveID)
		if err != nil {
			return loadResultMsg{err: err}
		}
		if err := m.store.TouchSave(context.Background(), saveID, now); err != nil {
			return loadResultMsg{err: err}
		}
		loaded.SaveIdentity.LastPlayedAt = now
		state, err := decodeSnapshot(loaded.DomainStateSnapshot, m.catalog, m.economy)
		if err != nil {
			return loadResultMsg{err: err}
		}
		loaded.TickCounter = int64(state.Day)
		return loadResultMsg{save: loaded, gameState: state, err: nil}
	}
}

func (m *Model) refreshMarketOffers() {
	if m.marketTick == m.gameState.Day && m.marketOffers != nil {
		return
	}
	m.marketTick = m.gameState.Day
	m.marketOffers = domain.MarketOffersForTick(m.simSeed, m.gameState.Day, 6)
	m.dashboard.SetMarketJobs(m.marketOffers)
}

func encodeSnapshot(state domain.GameState) map[string]any {
	return map[string]any{
		"day":               state.Day,
		"cash":              state.Cash,
		"headcount":         state.Headcount,
		"player_inventory":  state.PlayerInventory.Entries(),
		"company_inventory": state.CompanyInventory.Entries(),
		"active_jobs":       state.ActiveJobs,
		"completed_jobs":    state.CompletedJobs,
	}
}

func decodeSnapshot(snapshot map[string]any, catalog domain.Catalog, economy domain.EconomyConfig) (domain.GameState, error) {
	if snapshot == nil {
		bootstrap := domain.DefaultBootstrapConfig(catalog)
		return domain.BootstrapState(catalog, economy, bootstrap)
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return domain.GameState{}, err
	}
	var payload snapshotPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return domain.GameState{}, err
	}
	state := domain.GameState{
		Day:              payload.Day,
		Cash:             payload.Cash,
		Headcount:        payload.Headcount,
		PlayerInventory:  domain.NewInventoryFromEntries(payload.PlayerInventory),
		CompanyInventory: domain.NewInventoryFromEntries(payload.CompanyInventory),
		ActiveJobs:       payload.ActiveJobs,
		CompletedJobs:    payload.CompletedJobs,
	}
	state = domain.RecomputeMetrics(state, catalog, economy)
	return state, nil
}

func (m *Model) View() string {
	base := ""
	switch m.machine.State() {
	case app.BootMenuState:
		base = m.boot.View()
	case app.NewGameFormState:
		base = m.newGame.View()
	case app.LoadGameListState:
		base = m.loadGame.View()
	case app.HelpState:
		base = m.help.View()
	case app.LoadingGameState:
		base = fmt.Sprintf("\n\n  Loading %s", m.spinner.View())
	case app.DashboardState:
		base = m.dashboard.View()
	case app.FatalErrorState:
		base = m.fatal.View()
	}
	if m.recoverModal.Open {
		base += fmt.Sprintf("\n\n[%s] %s\n(%s)", m.recoverModal.Title, m.recoverModal.Body, m.recoverModal.Selected())
	}
	if m.toast.Visible {
		base += "\n\n" + m.toast.Message
	}
	return base
}
