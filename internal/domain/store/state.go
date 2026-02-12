package store

type Metrics struct {
	Cash         int     `json:"cash"`
	BurnRate     int     `json:"burn_rate"`
	Revenue      int     `json:"revenue"`
	RunwayMonths float64 `json:"runway_months"`
	Valuation    int     `json:"valuation"`
	Reputation   int     `json:"reputation"`
	Productivity int     `json:"productivity"`
	Morale       int     `json:"morale"`
	TechDebt     int     `json:"tech_debt"`
}

type EconomyConfig struct {
	BaseBurn             int `json:"base_burn"`
	BaseRevenue          int `json:"base_revenue"`
	BaseProductivity     int `json:"base_productivity"`
	BaseMorale           int `json:"base_morale"`
	ValuationMultiplier  int `json:"valuation_multiplier"`
	PayrollCost          int `json:"payroll_cost"`
	ToolMaintenance      int `json:"tool_maintenance"`
	ReputationBPSPerUnit int `json:"reputation_bps_per_unit"`
}

type GameState struct {
	Day              int            `json:"day"`
	Cash             int            `json:"cash"`
	Headcount        int            `json:"headcount"`
	PlayerInventory  Inventory      `json:"player_inventory"`
	CompanyInventory Inventory      `json:"company_inventory"`
	ActiveJobs       []ActiveJob    `json:"active_jobs"`
	CompletedJobs    []CompletedJob `json:"completed_jobs"`
	Metrics          Metrics        `json:"metrics"`
}

func DefaultEconomyConfig() EconomyConfig {
	return EconomyConfig{
		BaseBurn:             100,
		BaseRevenue:          40,
		BaseProductivity:     10,
		BaseMorale:           10,
		ValuationMultiplier:  20,
		PayrollCost:          0,
		ToolMaintenance:      0,
		ReputationBPSPerUnit: 20,
	}
}

func NewInitialState(startingCash int, catalog Catalog, cfg EconomyConfig) GameState {
	s := GameState{
		Cash:             startingCash,
		Headcount:        1,
		PlayerInventory:  NewInventory(),
		CompanyInventory: NewInventory(),
		ActiveJobs:       []ActiveJob{},
		CompletedJobs:    []CompletedJob{},
	}
	return RecomputeMetrics(s, catalog, cfg)
}

func RecomputeMetrics(state GameState, catalog Catalog, cfg EconomyConfig) GameState {
	effects := state.CompanyInventory.ActiveEffects(catalog)
	m := state.Metrics
	m.Cash = state.Cash
	m.Productivity = cfg.BaseProductivity + effects.ProductivityDelta - m.TechDebt
	if m.Productivity < 0 {
		m.Productivity = 0
	}
	m.Morale = cfg.BaseMorale + effects.MoraleDelta
	if m.Morale < 0 {
		m.Morale = 0
	}
	m.Reputation += effects.ReputationDelta
	m.TechDebt += effects.TechDebtDelta

	repModifierBPS := 10000 + (m.Reputation * cfg.ReputationBPSPerUnit)
	m.Revenue = (cfg.BaseRevenue * m.Productivity * repModifierBPS) / 10000
	m.BurnRate = cfg.BaseBurn + cfg.PayrollCost + cfg.ToolMaintenance
	if m.BurnRate > 0 {
		m.RunwayMonths = float64(state.Cash) / float64(m.BurnRate)
	} else {
		m.RunwayMonths = 0
	}
	m.Valuation = m.Revenue * cfg.ValuationMultiplier

	state.Metrics = m
	return state
}
