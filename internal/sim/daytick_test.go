package sim

import (
	"testing"

	domain "terminal-business/internal/domain/store"
)

func TestAdvanceDayDeterministicDurabilityDecay(t *testing.T) {
	catalog := domain.DefaultCatalog()
	econ := domain.DefaultEconomyConfig()
	state := domain.GameState{
		Day:             0,
		Cash:            1000,
		PlayerInventory: domain.NewInventory(),
		CompanyInventory: domain.NewInventoryFromEntries([]domain.InventoryItemInstance{
			{ItemID: domain.ItemDesk, Quantity: 1, AcquiredAtDay: 0, RemainingDurabilityDays: 2},
			{ItemID: domain.ItemChair, Quantity: 1, AcquiredAtDay: 0, RemainingDurabilityDays: 1},
			{ItemID: domain.ItemComputer, Quantity: 1, AcquiredAtDay: 0, RemainingDurabilityDays: 1},
		}),
	}
	state = domain.RecomputeMetrics(state, catalog, econ)
	a := AdvanceDay(state, catalog, econ)
	b := AdvanceDay(state, catalog, econ)
	if a.Day != 1 || b.Day != 1 {
		t.Fatalf("day mismatch %d %d", a.Day, b.Day)
	}
	aEntries := a.CompanyInventory.Entries()
	bEntries := b.CompanyInventory.Entries()
	for i := range aEntries {
		if aEntries[i] != bEntries[i] {
			t.Fatalf("nondeterministic entries: %+v %+v", aEntries, bEntries)
		}
	}
	if a.CompanyInventory.ActiveQuantity(domain.ItemChair) != 0 {
		t.Fatalf("chair should be broken")
	}
	if a.CompanyInventory.ActiveQuantity(domain.ItemComputer) != 0 {
		t.Fatalf("computer should be broken")
	}
}

func TestAdvanceDayBrokenItemsStopAffectingMetrics(t *testing.T) {
	catalog := domain.DefaultCatalog()
	econ := domain.DefaultEconomyConfig()
	state := domain.GameState{
		Day:             0,
		Cash:            1000,
		PlayerInventory: domain.NewInventory(),
		CompanyInventory: domain.NewInventoryFromEntries([]domain.InventoryItemInstance{
			{ItemID: domain.ItemComputer, Quantity: 1, AcquiredAtDay: 0, RemainingDurabilityDays: 1},
		}),
	}
	state = domain.RecomputeMetrics(state, catalog, econ)
	if state.Metrics.Productivity <= econ.BaseProductivity {
		t.Fatalf("expected productivity bonus before break")
	}
	next := AdvanceDay(state, catalog, econ)
	if next.Metrics.Productivity != econ.BaseProductivity {
		t.Fatalf("expected bonus removed after break, got %d", next.Metrics.Productivity)
	}
}
