package store

import (
	"errors"
	"testing"
)

func TestDefaultCatalogContainsRequiredStableItems(t *testing.T) {
	catalog := DefaultCatalog()
	items := catalog.OrderedItems()
	if len(items) != 3 {
		t.Fatalf("got %d items", len(items))
	}
	ids := []ItemID{items[0].ItemID, items[1].ItemID, items[2].ItemID}
	if ids[0] != ItemChair || ids[1] != ItemComputer || ids[2] != ItemDesk {
		t.Fatalf("unexpected stable order: %+v", ids)
	}
	for _, id := range []ItemID{ItemDesk, ItemChair, ItemComputer} {
		item, ok := catalog.Item(id)
		if !ok {
			t.Fatalf("missing item %s", id)
		}
		if item.ItemID != id {
			t.Fatalf("stable id mismatch for %s", id)
		}
		if item.MaxOwned != 1 {
			t.Fatalf("expected max owned 1 for %s", id)
		}
		if item.DurabilityDays <= 0 {
			t.Fatalf("expected durability for %s", id)
		}
		if item.Ownership != OwnershipCompany {
			t.Fatalf("expected company ownership for %s", id)
		}
	}
}

func TestCatalogConfigDrivenValues(t *testing.T) {
	cfg := CatalogConfig{Items: map[ItemID]Item{
		ItemDesk: {
			DisplayName:    "Desk",
			Category:       "Office",
			Ownership:      OwnershipPlayer,
			Price:          999,
			DurabilityDays: 7,
			Effects:        Effects{ProductivityDelta: 9},
		},
	}}
	catalog := NewCatalog(cfg)
	item, ok := catalog.Item(ItemDesk)
	if !ok {
		t.Fatal("expected desk")
	}
	if item.Price != 999 || item.Effects.ProductivityDelta != 9 || item.ItemID != ItemDesk || item.Ownership != OwnershipPlayer {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestInventoryDeterministicOrderingAndQuantity(t *testing.T) {
	inv := NewInventoryFromEntries([]InventoryItemInstance{
		{ItemID: ItemDesk, Quantity: 1, AcquiredAtDay: 4, RemainingDurabilityDays: 10},
		{ItemID: ItemChair, Quantity: 2, AcquiredAtDay: 1, RemainingDurabilityDays: 2},
		{ItemID: ItemComputer, Quantity: 1, AcquiredAtDay: 3, RemainingDurabilityDays: 0},
		{ItemID: ItemDesk, Quantity: 1, AcquiredAtDay: 2, RemainingDurabilityDays: -5},
		{ItemID: ItemDesk, Quantity: 1, AcquiredAtDay: 2, RemainingDurabilityDays: 9},
		{ItemID: ItemDesk, Quantity: 0, AcquiredAtDay: 0, RemainingDurabilityDays: 1},
	})
	entries := inv.Entries()
	if len(entries) != 5 {
		t.Fatalf("got %d", len(entries))
	}
	if entries[0].ItemID != ItemChair || entries[1].ItemID != ItemComputer || entries[2].ItemID != ItemDesk {
		t.Fatalf("entries not sorted: %+v", entries)
	}
	if entries[2].AcquiredAtDay != 2 || entries[2].RemainingDurabilityDays != 0 {
		t.Fatalf("expected normalized durability and acquired sort: %+v", entries[2])
	}
	if entries[3].AcquiredAtDay != 2 || entries[3].RemainingDurabilityDays != 9 {
		t.Fatalf("expected durability tie sort: %+v", entries[3])
	}
	if inv.Quantity(ItemChair) != 2 {
		t.Fatalf("qty got %d", inv.Quantity(ItemChair))
	}
	if inv.ActiveQuantity(ItemComputer) != 0 {
		t.Fatalf("active qty got %d", inv.ActiveQuantity(ItemComputer))
	}
}

func TestInventoryAddAndDurabilityDecay(t *testing.T) {
	catalog := DefaultCatalog()
	desk, _ := catalog.Item(ItemDesk)
	inv := NewInventory().Add(desk, 1, 1)
	if inv.Quantity(ItemDesk) != 1 {
		t.Fatalf("qty got %d", inv.Quantity(ItemDesk))
	}
	inv = inv.DecrementDurabilityByDay()
	entries := inv.Entries()
	if entries[0].RemainingDurabilityDays != desk.DurabilityDays-1 {
		t.Fatalf("durability got %d", entries[0].RemainingDurabilityDays)
	}
	unchanged := inv.Add(desk, 2, 0)
	if len(unchanged.Entries()) != len(inv.Entries()) {
		t.Fatal("expected no change when adding non-positive quantity")
	}
}

func TestActiveEffectsExcludesBroken(t *testing.T) {
	catalog := DefaultCatalog()
	inv := NewInventoryFromEntries([]InventoryItemInstance{
		{ItemID: ItemDesk, Quantity: 1, AcquiredAtDay: 1, RemainingDurabilityDays: 10},
		{ItemID: ItemChair, Quantity: 1, AcquiredAtDay: 1, RemainingDurabilityDays: 0},
	})
	effects := inv.ActiveEffects(catalog)
	if effects.ProductivityDelta != 2 {
		t.Fatalf("prod got %d", effects.ProductivityDelta)
	}
	if effects.MoraleDelta != 1 {
		t.Fatalf("morale got %d", effects.MoraleDelta)
	}

	unknown := NewInventoryFromEntries([]InventoryItemInstance{{ItemID: ItemID("unknown"), Quantity: 1, RemainingDurabilityDays: 5}})
	effects = unknown.ActiveEffects(catalog)
	if effects != (Effects{}) {
		t.Fatalf("expected zero effects for unknown item, got %+v", effects)
	}
}

func TestApplyPurchaseTable(t *testing.T) {
	catalog := DefaultCatalog()
	econ := DefaultEconomyConfig()
	base := NewInitialState(1000, catalog, econ)

	playerCatalog := NewCatalog(CatalogConfig{Items: map[ItemID]Item{
		ItemDesk: {
			DisplayName:    "Notebook",
			Category:       "Personal",
			Ownership:      OwnershipPlayer,
			Price:          10,
			DurabilityDays: 20,
			MaxOwned:       2,
		},
	}})

	tests := []struct {
		name    string
		state   GameState
		catalog Catalog
		itemID  ItemID
		wantErr error
		check   func(t *testing.T, got GameState)
	}{
		{
			name:    "company purchase routes to company inventory",
			state:   base,
			catalog: catalog,
			itemID:  ItemDesk,
			check: func(t *testing.T, got GameState) {
				if got.Cash != base.Cash-150 {
					t.Fatalf("cash got %d", got.Cash)
				}
				if got.CompanyInventory.Quantity(ItemDesk) != 1 {
					t.Fatalf("company desk qty got %d", got.CompanyInventory.Quantity(ItemDesk))
				}
				if got.PlayerInventory.Quantity(ItemDesk) != 0 {
					t.Fatalf("player desk qty got %d", got.PlayerInventory.Quantity(ItemDesk))
				}
			},
		},
		{
			name:    "player ownership routes to player inventory",
			state:   base,
			catalog: playerCatalog,
			itemID:  ItemDesk,
			check: func(t *testing.T, got GameState) {
				if got.PlayerInventory.Quantity(ItemDesk) != 1 {
					t.Fatalf("player qty got %d", got.PlayerInventory.Quantity(ItemDesk))
				}
				if got.CompanyInventory.Quantity(ItemDesk) != 0 {
					t.Fatalf("company qty got %d", got.CompanyInventory.Quantity(ItemDesk))
				}
			},
		},
		{name: "unknown item", state: base, catalog: catalog, itemID: ItemID("unknown"), wantErr: ErrItemNotFound},
		{name: "insufficient funds", state: NewInitialState(10, catalog, econ), catalog: catalog, itemID: ItemComputer, wantErr: ErrInsufficientFunds},
		{
			name:    "max owned",
			state:   mustPurchase(t, base, catalog, econ, ItemChair),
			catalog: catalog,
			itemID:  ItemChair,
			wantErr: ErrMaxOwned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyPurchase(tt.state, tt.catalog, econ, tt.itemID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err got %v want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestApplyPurchaseInvalidStateBranch(t *testing.T) {
	cfg := CatalogConfig{Items: map[ItemID]Item{
		ItemDesk: {
			ItemID:         ItemDesk,
			DisplayName:    "Desk",
			Category:       "Office",
			Ownership:      OwnershipCompany,
			Price:          -1,
			DurabilityDays: 0,
			MaxOwned:       1,
		},
	}}
	catalog := NewCatalog(cfg)
	state := NewInitialState(10, catalog, DefaultEconomyConfig())
	_, err := ApplyPurchase(state, catalog, DefaultEconomyConfig(), ItemDesk)
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("got %v", err)
	}
}

func TestApplyPurchaseUnlimitedOwnership(t *testing.T) {
	catalog := NewCatalog(CatalogConfig{Items: map[ItemID]Item{
		ItemDesk: {
			ItemID:         ItemDesk,
			DisplayName:    "Desk",
			Category:       "Office",
			Ownership:      OwnershipCompany,
			Price:          1,
			DurabilityDays: 5,
			MaxOwned:       0,
		},
	}})
	state := NewInitialState(10, catalog, DefaultEconomyConfig())
	next, err := ApplyPurchase(state, catalog, DefaultEconomyConfig(), ItemDesk)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	next, err = ApplyPurchase(next, catalog, DefaultEconomyConfig(), ItemDesk)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if next.CompanyInventory.Quantity(ItemDesk) != 2 {
		t.Fatalf("got qty %d", next.CompanyInventory.Quantity(ItemDesk))
	}
}

func TestRecomputeMetricsFormulaFlow(t *testing.T) {
	catalog := DefaultCatalog()
	cfg := DefaultEconomyConfig()
	cfg.BaseBurn = 120
	cfg.PayrollCost = 30
	cfg.ToolMaintenance = 10
	cfg.BaseRevenue = 50
	cfg.BaseProductivity = 10
	cfg.BaseMorale = 10
	cfg.ValuationMultiplier = 25
	cfg.ReputationBPSPerUnit = 10

	state := GameState{
		Cash:             1000,
		PlayerInventory:  NewInventory(),
		CompanyInventory: NewInventoryFromEntries([]InventoryItemInstance{{ItemID: ItemComputer, Quantity: 1, RemainingDurabilityDays: 120}}),
		Metrics: Metrics{
			Reputation: 2,
			TechDebt:   1,
		},
	}
	next := RecomputeMetrics(state, catalog, cfg)
	if next.Metrics.Productivity != 14 {
		t.Fatalf("got productivity %d", next.Metrics.Productivity)
	}
	if next.Metrics.Morale != 10 {
		t.Fatalf("got morale %d", next.Metrics.Morale)
	}
	if next.Metrics.Revenue != 701 {
		t.Fatalf("got revenue %d", next.Metrics.Revenue)
	}
	if next.Metrics.BurnRate != 160 {
		t.Fatalf("got burn %d", next.Metrics.BurnRate)
	}
	if next.Metrics.RunwayMonths != 6.25 {
		t.Fatalf("got runway %f", next.Metrics.RunwayMonths)
	}
	if next.Metrics.Valuation != 17525 {
		t.Fatalf("got valuation %d", next.Metrics.Valuation)
	}
}

func TestRecomputeMetricsClampsAndBrokenEffectsExcluded(t *testing.T) {
	catalog := NewCatalog(CatalogConfig{Items: map[ItemID]Item{
		ItemDesk: {
			ItemID:         ItemDesk,
			DisplayName:    "Desk",
			Category:       "Office",
			Ownership:      OwnershipCompany,
			Price:          1,
			DurabilityDays: 10,
			Effects:        Effects{ProductivityDelta: -100, MoraleDelta: -100},
			MaxOwned:       10,
		},
	}})
	state := GameState{
		Cash:             0,
		PlayerInventory:  NewInventory(),
		CompanyInventory: NewInventoryFromEntries([]InventoryItemInstance{{ItemID: ItemDesk, Quantity: 1, RemainingDurabilityDays: 0}}),
		Metrics:          Metrics{TechDebt: 100},
	}
	cfg := DefaultEconomyConfig()
	cfg.BaseBurn = 0
	next := RecomputeMetrics(state, catalog, cfg)
	if next.Metrics.Productivity != 0 {
		t.Fatalf("got %d", next.Metrics.Productivity)
	}
	if next.Metrics.Morale != 10 {
		t.Fatalf("got %d", next.Metrics.Morale)
	}
	if next.Metrics.RunwayMonths != 0 {
		t.Fatalf("got %f", next.Metrics.RunwayMonths)
	}
}

func TestRecomputeMetricsMoraleClampAndSortTieBreakers(t *testing.T) {
	catalog := NewCatalog(CatalogConfig{Items: map[ItemID]Item{
		ItemDesk: {
			ItemID:         ItemDesk,
			DisplayName:    "Desk",
			Category:       "Office",
			Ownership:      OwnershipCompany,
			Price:          1,
			DurabilityDays: 10,
			Effects:        Effects{MoraleDelta: -100},
			MaxOwned:       10,
		},
	}})
	state := GameState{
		Cash:            0,
		PlayerInventory: NewInventory(),
		CompanyInventory: NewInventoryFromEntries([]InventoryItemInstance{
			{ItemID: ItemDesk, Quantity: 2, AcquiredAtDay: 1, RemainingDurabilityDays: 5},
			{ItemID: ItemDesk, Quantity: 1, AcquiredAtDay: 1, RemainingDurabilityDays: 5},
		}),
	}
	entries := state.CompanyInventory.Entries()
	if entries[0].Quantity != 1 || entries[1].Quantity != 2 {
		t.Fatalf("expected quantity tie-break sorting, got %+v", entries)
	}
	cfg := DefaultEconomyConfig()
	cfg.BaseMorale = 0
	next := RecomputeMetrics(state, catalog, cfg)
	if next.Metrics.Morale != 0 {
		t.Fatalf("expected morale clamp to 0, got %d", next.Metrics.Morale)
	}
}

func mustPurchase(t *testing.T, s GameState, c Catalog, cfg EconomyConfig, item ItemID) GameState {
	t.Helper()
	next, err := ApplyPurchase(s, c, cfg, item)
	if err != nil {
		t.Fatalf("purchase failed: %v", err)
	}
	return next
}
