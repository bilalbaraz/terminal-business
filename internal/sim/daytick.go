package sim

import domain "terminal-business/internal/domain/store"

func AdvanceDay(state domain.GameState, catalog domain.Catalog, econ domain.EconomyConfig) domain.GameState {
	next := state
	next.Day++
	// MVP rule: all company equipment is considered in-use.
	next.CompanyInventory = next.CompanyInventory.DecrementDurabilityByDay()
	next = domain.CompleteDueJobs(next, next.Day)
	return domain.RecomputeMetrics(next, catalog, econ)
}
