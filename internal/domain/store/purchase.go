package store

import "errors"

var (
	ErrItemNotFound      = errors.New("item not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrMaxOwned          = errors.New("max owned reached")
	ErrInvalidState      = errors.New("invalid state")
)

func ApplyPurchase(state GameState, catalog Catalog, cfg EconomyConfig, itemID ItemID) (GameState, error) {
	item, ok := catalog.Item(itemID)
	if !ok {
		return GameState{}, ErrItemNotFound
	}
	if item.Price < 0 {
		return GameState{}, ErrInvalidState
	}
	if state.Cash < item.Price {
		return GameState{}, ErrInsufficientFunds
	}
	if item.MaxOwned > 0 && state.Inventory.Quantity(itemID) >= item.MaxOwned {
		return GameState{}, ErrMaxOwned
	}

	next := state
	next.Cash -= item.Price
	next.Inventory = next.Inventory.WithAdded(itemID, 1)
	next = RecomputeMetrics(next, catalog, cfg)
	return next, nil
}
