package store

type Inventory struct {
	quantities map[ItemID]int
}

type InventoryEntry struct {
	ItemID   ItemID `json:"item_id"`
	Quantity int    `json:"quantity"`
}

func NewInventory() Inventory {
	return Inventory{quantities: map[ItemID]int{}}
}

func NewInventoryFromEntries(entries []InventoryEntry) Inventory {
	inv := NewInventory()
	for _, e := range entries {
		if e.Quantity > 0 {
			inv.quantities[e.ItemID] = e.Quantity
		}
	}
	return inv
}

func (i Inventory) Quantity(id ItemID) int {
	return i.quantities[id]
}

func (i Inventory) WithAdded(id ItemID, delta int) Inventory {
	next := NewInventoryFromEntries(i.Entries())
	next.quantities[id] += delta
	if next.quantities[id] <= 0 {
		delete(next.quantities, id)
	}
	return next
}

func (i Inventory) Entries() []InventoryEntry {
	ids := make([]ItemID, 0, len(i.quantities))
	for id := range i.quantities {
		ids = append(ids, id)
	}
	for a := 0; a < len(ids)-1; a++ {
		for b := a + 1; b < len(ids); b++ {
			if ids[a] > ids[b] {
				ids[a], ids[b] = ids[b], ids[a]
			}
		}
	}
	entries := make([]InventoryEntry, 0, len(ids))
	for _, id := range ids {
		entries = append(entries, InventoryEntry{ItemID: id, Quantity: i.quantities[id]})
	}
	return entries
}

func AggregateEffects(c Catalog, inv Inventory) Effects {
	total := Effects{}
	for _, item := range c.OrderedItems() {
		qty := inv.Quantity(item.ItemID)
		total.ProductivityDelta += item.Effects.ProductivityDelta * qty
		total.MoraleDelta += item.Effects.MoraleDelta * qty
		total.TechDebtDelta += item.Effects.TechDebtDelta * qty
		total.ReputationDelta += item.Effects.ReputationDelta * qty
	}
	return total
}
