package store

import (
	"errors"
	"testing"
)

func TestIsOperationalRequiresDeskChairComputerAlive(t *testing.T) {
	catalog := DefaultCatalog()
	econ := DefaultEconomyConfig()
	base := NewInitialState(100, catalog, econ)

	cases := []struct {
		name  string
		state GameState
		want  bool
	}{
		{name: "no items", state: base, want: false},
		{name: "desk only", state: GameState{CompanyInventory: NewInventoryFromEntries([]InventoryItemInstance{{ItemID: ItemDesk, Quantity: 1, RemainingDurabilityDays: 1}})}, want: false},
		{name: "chair only", state: GameState{CompanyInventory: NewInventoryFromEntries([]InventoryItemInstance{{ItemID: ItemChair, Quantity: 1, RemainingDurabilityDays: 1}})}, want: false},
		{name: "computer only", state: GameState{CompanyInventory: NewInventoryFromEntries([]InventoryItemInstance{{ItemID: ItemComputer, Quantity: 1, RemainingDurabilityDays: 1}})}, want: false},
		{name: "all required", state: GameState{CompanyInventory: NewInventoryFromEntries([]InventoryItemInstance{{ItemID: ItemDesk, Quantity: 1, RemainingDurabilityDays: 1}, {ItemID: ItemChair, Quantity: 1, RemainingDurabilityDays: 1}, {ItemID: ItemComputer, Quantity: 1, RemainingDurabilityDays: 1}})}, want: true},
		{name: "computer broken", state: GameState{CompanyInventory: NewInventoryFromEntries([]InventoryItemInstance{{ItemID: ItemDesk, Quantity: 1, RemainingDurabilityDays: 1}, {ItemID: ItemChair, Quantity: 1, RemainingDurabilityDays: 1}, {ItemID: ItemComputer, Quantity: 1, RemainingDurabilityDays: 0}})}, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsOperational(tc.state); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestDefaultBootstrapConfigAndMinimum(t *testing.T) {
	catalog := DefaultCatalog()
	cfg := DefaultBootstrapConfig(catalog)
	if cfg.BufferCash != 75 {
		t.Fatalf("got buffer %d", cfg.BufferCash)
	}
	if cfg.StartingCash != 825 {
		t.Fatalf("got starting cash %d", cfg.StartingCash)
	}
	min, err := MinimumStartingCash(catalog, cfg.BufferCash)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if min != cfg.StartingCash {
		t.Fatalf("got min %d", min)
	}
}

func TestValidateBootstrapConfig(t *testing.T) {
	catalog := DefaultCatalog()

	if err := ValidateBootstrapConfig(catalog, BootstrapConfig{StartingCash: 825, BufferCash: 75}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := ValidateBootstrapConfig(catalog, BootstrapConfig{StartingCash: -1, BufferCash: 0}); !errors.Is(err, ErrInvalidBootstrapConfig) {
		t.Fatalf("got %v", err)
	}
	if err := ValidateBootstrapConfig(catalog, BootstrapConfig{StartingCash: 700, BufferCash: 75}); !errors.Is(err, ErrInsufficientStartingCashConfig) {
		t.Fatalf("got %v", err)
	}
}

func TestOperationalReadinessCostAndMissingItemPaths(t *testing.T) {
	missing := NewCatalog(CatalogConfig{Items: map[ItemID]Item{ItemDesk: {DisplayName: "Desk", Price: 1}}})
	if _, err := OperationalReadinessCost(missing); !errors.Is(err, ErrOperationalItemNotFound) {
		t.Fatalf("got %v", err)
	}
	missingDesk := NewCatalog(CatalogConfig{Items: map[ItemID]Item{ItemChair: {DisplayName: "Chair", Price: 1}, ItemComputer: {DisplayName: "Computer", Price: 1}}})
	if _, err := OperationalReadinessCost(missingDesk); !errors.Is(err, ErrOperationalItemNotFound) {
		t.Fatalf("got %v", err)
	}
	missingComputer := NewCatalog(CatalogConfig{Items: map[ItemID]Item{ItemDesk: {DisplayName: "Desk", Price: 1}, ItemChair: {DisplayName: "Chair", Price: 1}}})
	if _, err := OperationalReadinessCost(missingComputer); !errors.Is(err, ErrOperationalItemNotFound) {
		t.Fatalf("got %v", err)
	}
	if _, err := MinimumStartingCash(DefaultCatalog(), -1); !errors.Is(err, ErrInvalidBootstrapConfig) {
		t.Fatalf("got %v", err)
	}
	if _, err := MinimumStartingCash(missing, 1); !errors.Is(err, ErrOperationalItemNotFound) {
		t.Fatalf("got %v", err)
	}
	if err := ValidateBootstrapConfig(missing, BootstrapConfig{StartingCash: 10, BufferCash: 1}); !errors.Is(err, ErrOperationalItemNotFound) {
		t.Fatalf("got %v", err)
	}
	if cfg := DefaultBootstrapConfig(missing); cfg != (BootstrapConfig{}) {
		t.Fatalf("expected zero config, got %+v", cfg)
	}
}

func TestBootstrapStateDeterministicAndOperationalGuarantee(t *testing.T) {
	catalog := DefaultCatalog()
	econ := DefaultEconomyConfig()
	cfg := DefaultBootstrapConfig(catalog)

	s1, err := BootstrapState(catalog, econ, cfg)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	s2, err := BootstrapState(catalog, econ, cfg)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if s1.Cash != s2.Cash {
		t.Fatalf("nondeterministic cash: %d != %d", s1.Cash, s2.Cash)
	}
	if s1.Cash < 750 {
		t.Fatalf("starting cash not enough for desk+chair+computer: %d", s1.Cash)
	}
	next, err := ApplyPurchase(s1, catalog, econ, ItemDesk)
	if err != nil {
		t.Fatalf("desk purchase failed: %v", err)
	}
	next, err = ApplyPurchase(next, catalog, econ, ItemChair)
	if err != nil {
		t.Fatalf("chair purchase failed: %v", err)
	}
	next, err = ApplyPurchase(next, catalog, econ, ItemComputer)
	if err != nil {
		t.Fatalf("computer purchase failed: %v", err)
	}
	if !IsOperational(next) {
		t.Fatal("expected operational after desk+chair+computer")
	}

	if _, err := BootstrapState(catalog, econ, BootstrapConfig{StartingCash: 1, BufferCash: 0}); !errors.Is(err, ErrInsufficientStartingCashConfig) {
		t.Fatalf("got %v", err)
	}
}
