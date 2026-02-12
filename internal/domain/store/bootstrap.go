package store

import "errors"

var (
	ErrOperationalItemNotFound        = errors.New("operational item not found in catalog")
	ErrInsufficientStartingCashConfig = errors.New("starting cash is below operational minimum")
	ErrInvalidBootstrapConfig         = errors.New("invalid bootstrap config")
)

type BootstrapConfig struct {
	StartingCash int `json:"starting_cash"`
	BufferCash   int `json:"buffer_cash"`
}

func IsOperational(state GameState) bool {
	return state.Inventory.Quantity(ItemDesk) >= 1 && state.Inventory.Quantity(ItemChair) >= 1
}

func DefaultBootstrapConfig(catalog Catalog) BootstrapConfig {
	rulesCost, err := OperationalReadinessCost(catalog)
	if err != nil {
		return BootstrapConfig{}
	}
	buffer := rulesCost / 10
	return BootstrapConfig{
		StartingCash: rulesCost + buffer,
		BufferCash:   buffer,
	}
}

func OperationalReadinessCost(catalog Catalog) (int, error) {
	desk, ok := catalog.Item(ItemDesk)
	if !ok {
		return 0, ErrOperationalItemNotFound
	}
	chair, ok := catalog.Item(ItemChair)
	if !ok {
		return 0, ErrOperationalItemNotFound
	}
	return desk.Price + chair.Price, nil
}

func MinimumStartingCash(catalog Catalog, bufferCash int) (int, error) {
	if bufferCash < 0 {
		return 0, ErrInvalidBootstrapConfig
	}
	cost, err := OperationalReadinessCost(catalog)
	if err != nil {
		return 0, err
	}
	return cost + bufferCash, nil
}

func ValidateBootstrapConfig(catalog Catalog, cfg BootstrapConfig) error {
	if cfg.StartingCash < 0 || cfg.BufferCash < 0 {
		return ErrInvalidBootstrapConfig
	}
	min, err := MinimumStartingCash(catalog, cfg.BufferCash)
	if err != nil {
		return err
	}
	if cfg.StartingCash < min {
		return ErrInsufficientStartingCashConfig
	}
	return nil
}

func BootstrapState(catalog Catalog, economy EconomyConfig, cfg BootstrapConfig) (GameState, error) {
	if err := ValidateBootstrapConfig(catalog, cfg); err != nil {
		return GameState{}, err
	}
	return NewInitialState(cfg.StartingCash, catalog, economy), nil
}
