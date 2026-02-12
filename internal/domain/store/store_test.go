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
	}
}

func TestCatalogConfigDrivenValues(t *testing.T) {
	cfg := CatalogConfig{Items: map[ItemID]Item{
		ItemDesk: {DisplayName: "Desk", Category: "Office", Price: 999, Effects: Effects{ProductivityDelta: 9}},
	}}
	catalog := NewCatalog(cfg)
	item, ok := catalog.Item(ItemDesk)
	if !ok {
		t.Fatal("expected desk")
	}
	if item.Price != 999 || item.Effects.ProductivityDelta != 9 || item.ItemID != ItemDesk {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestInventoryAndDeterministicOrdering(t *testing.T) {
	inv := NewInventoryFromEntries([]InventoryEntry{
		{ItemID: ItemDesk, Quantity: 1},
		{ItemID: ItemChair, Quantity: 2},
		{ItemID: ItemComputer, Quantity: 1},
	})
	if inv.Quantity(ItemChair) != 2 {
		t.Fatalf("got %d", inv.Quantity(ItemChair))
	}
	entries := inv.Entries()
	if len(entries) != 3 {
		t.Fatalf("got %d", len(entries))
	}
	if entries[0].ItemID != ItemChair || entries[1].ItemID != ItemComputer || entries[2].ItemID != ItemDesk {
		t.Fatalf("entries not sorted: %+v", entries)
	}
}

func TestInventoryWithAddedAndDeletePath(t *testing.T) {
	inv := NewInventory()
	inv = inv.WithAdded(ItemDesk, 1)
	if inv.Quantity(ItemDesk) != 1 {
		t.Fatalf("got %d", inv.Quantity(ItemDesk))
	}
	inv = inv.WithAdded(ItemDesk, -1)
	if inv.Quantity(ItemDesk) != 0 {
		t.Fatalf("got %d", inv.Quantity(ItemDesk))
	}
}

func TestAggregateEffects(t *testing.T) {
	catalog := DefaultCatalog()
	inv := NewInventoryFromEntries([]InventoryEntry{{ItemID: ItemDesk, Quantity: 1}, {ItemID: ItemChair, Quantity: 1}})
	effects := AggregateEffects(catalog, inv)
	if effects.ProductivityDelta != 3 {
		t.Fatalf("got %d", effects.ProductivityDelta)
	}
	if effects.MoraleDelta != 3 {
		t.Fatalf("got %d", effects.MoraleDelta)
	}
}

func TestApplyPurchaseTable(t *testing.T) {
	catalog := DefaultCatalog()
	econ := DefaultEconomyConfig()
	base := NewInitialState(1000, catalog, econ)

	tests := []struct {
		name    string
		state   GameState
		itemID  ItemID
		wantErr error
		check   func(t *testing.T, got GameState)
	}{
		{
			name:   "success",
			state:  base,
			itemID: ItemDesk,
			check: func(t *testing.T, got GameState) {
				if got.Cash != base.Cash-150 {
					t.Fatalf("cash got %d", got.Cash)
				}
				if got.Inventory.Quantity(ItemDesk) != 1 {
					t.Fatalf("desk qty got %d", got.Inventory.Quantity(ItemDesk))
				}
				if got.Metrics.Productivity <= base.Metrics.Productivity {
					t.Fatalf("productivity did not increase")
				}
			},
		},
		{name: "unknown item", state: base, itemID: ItemID("unknown"), wantErr: ErrItemNotFound},
		{name: "insufficient funds", state: NewInitialState(10, catalog, econ), itemID: ItemComputer, wantErr: ErrInsufficientFunds},
		{
			name:    "max owned",
			state:   mustPurchase(t, base, catalog, econ, ItemChair),
			itemID:  ItemChair,
			wantErr: ErrMaxOwned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyPurchase(tt.state, catalog, econ, tt.itemID)
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
		ItemDesk: {ItemID: ItemDesk, DisplayName: "Desk", Category: "Office", Price: -1, MaxOwned: 1},
	}}
	catalog := NewCatalog(cfg)
	state := NewInitialState(0, catalog, DefaultEconomyConfig())
	_, err := ApplyPurchase(state, catalog, DefaultEconomyConfig(), ItemDesk)
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("got %v", err)
	}
}

func TestApplyPurchaseUnlimitedOwnership(t *testing.T) {
	catalog := NewCatalog(CatalogConfig{Items: map[ItemID]Item{
		ItemDesk: {ItemID: ItemDesk, DisplayName: "Desk", Category: "Office", Price: 1, MaxOwned: 0},
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
	if next.Inventory.Quantity(ItemDesk) != 2 {
		t.Fatalf("got qty %d", next.Inventory.Quantity(ItemDesk))
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
		Cash:      1000,
		Inventory: NewInventoryFromEntries([]InventoryEntry{{ItemID: ItemComputer, Quantity: 1}}),
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

func TestRecomputeMetricsClamps(t *testing.T) {
	catalog := NewCatalog(CatalogConfig{Items: map[ItemID]Item{
		ItemDesk: {
			ItemID:      ItemDesk,
			DisplayName: "Desk",
			Category:    "Office",
			Price:       1,
			Effects:     Effects{ProductivityDelta: -100, MoraleDelta: -100},
			MaxOwned:    10,
		},
	}})
	state := GameState{
		Cash:      0,
		Inventory: NewInventoryFromEntries([]InventoryEntry{{ItemID: ItemDesk, Quantity: 1}}),
		Metrics:   Metrics{TechDebt: 100},
	}
	cfg := DefaultEconomyConfig()
	cfg.BaseBurn = 0
	next := RecomputeMetrics(state, catalog, cfg)
	if next.Metrics.Productivity != 0 {
		t.Fatalf("got %d", next.Metrics.Productivity)
	}
	if next.Metrics.Morale != 0 {
		t.Fatalf("got %d", next.Metrics.Morale)
	}
	if next.Metrics.RunwayMonths != 0 {
		t.Fatalf("got %f", next.Metrics.RunwayMonths)
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
