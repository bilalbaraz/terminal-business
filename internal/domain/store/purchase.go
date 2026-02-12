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
	if item.Price < 0 || item.DurabilityDays <= 0 {
		return GameState{}, ErrInvalidState
	}
	if state.Cash < item.Price {
		return GameState{}, ErrInsufficientFunds
	}

	target := state.CompanyInventory
	if item.Ownership == OwnershipPlayer {
		target = state.PlayerInventory
	}
	if item.MaxOwned > 0 && target.Quantity(itemID) >= item.MaxOwned {
		return GameState{}, ErrMaxOwned
	}

	next := state
	next.Cash -= item.Price
	if item.Ownership == OwnershipPlayer {
		next.PlayerInventory = next.PlayerInventory.Add(item, state.Day, 1)
	} else {
		next.CompanyInventory = next.CompanyInventory.Add(item, state.Day, 1)
	}
	next = RecomputeMetrics(next, catalog, cfg)
	return next, nil
}
