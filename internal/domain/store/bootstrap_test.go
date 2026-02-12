package store

import (
	"errors"
	"testing"
)

func TestIsOperational(t *testing.T) {
	catalog := DefaultCatalog()
	economy := DefaultEconomyConfig()

	cases := []struct {
		name  string
		state GameState
		want  bool
	}{
		{name: "no items", state: NewInitialState(100, catalog, economy), want: false},
		{name: "desk only", state: GameState{Cash: 100, Inventory: NewInventoryFromEntries([]InventoryEntry{{ItemID: ItemDesk, Quantity: 1}})}, want: false},
		{name: "chair only", state: GameState{Cash: 100, Inventory: NewInventoryFromEntries([]InventoryEntry{{ItemID: ItemChair, Quantity: 1}})}, want: false},
		{name: "desk and chair", state: GameState{Cash: 100, Inventory: NewInventoryFromEntries([]InventoryEntry{{ItemID: ItemDesk, Quantity: 1}, {ItemID: ItemChair, Quantity: 1}})}, want: true},
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
	if cfg.BufferCash != 25 {
		t.Fatalf("got buffer %d", cfg.BufferCash)
	}
	if cfg.StartingCash != 275 {
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

	if err := ValidateBootstrapConfig(catalog, BootstrapConfig{StartingCash: 275, BufferCash: 25}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := ValidateBootstrapConfig(catalog, BootstrapConfig{StartingCash: -1, BufferCash: 0}); !errors.Is(err, ErrInvalidBootstrapConfig) {
		t.Fatalf("got %v", err)
	}
	if err := ValidateBootstrapConfig(catalog, BootstrapConfig{StartingCash: 200, BufferCash: 25}); !errors.Is(err, ErrInsufficientStartingCashConfig) {
		t.Fatalf("got %v", err)
	}
}

func TestOperationalReadinessCostAndMissingItemPaths(t *testing.T) {
	missing := NewCatalog(CatalogConfig{Items: map[ItemID]Item{ItemDesk: {DisplayName: "Desk", Price: 1}}})
	if _, err := OperationalReadinessCost(missing); !errors.Is(err, ErrOperationalItemNotFound) {
		t.Fatalf("got %v", err)
	}
	missingDesk := NewCatalog(CatalogConfig{Items: map[ItemID]Item{ItemChair: {DisplayName: "Chair", Price: 1}}})
	if _, err := OperationalReadinessCost(missingDesk); !errors.Is(err, ErrOperationalItemNotFound) {
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
	if s1.Cash < 250 {
		t.Fatalf("starting cash not enough for desk+chair: %d", s1.Cash)
	}
	next, err := ApplyPurchase(s1, catalog, econ, ItemDesk)
	if err != nil {
		t.Fatalf("desk purchase failed: %v", err)
	}
	next, err = ApplyPurchase(next, catalog, econ, ItemChair)
	if err != nil {
		t.Fatalf("chair purchase failed: %v", err)
	}
	if !IsOperational(next) {
		t.Fatal("expected operational after desk+chair")
	}

	if _, err := BootstrapState(catalog, econ, BootstrapConfig{StartingCash: 1, BufferCash: 0}); !errors.Is(err, ErrInsufficientStartingCashConfig) {
		t.Fatalf("got %v", err)
	}
}
