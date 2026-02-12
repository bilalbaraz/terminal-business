package store

import "sort"

type InventoryItemInstance struct {
	ItemID                  ItemID `json:"item_id"`
	Quantity                int    `json:"quantity"`
	AcquiredAtDay           int    `json:"acquired_at_day"`
	RemainingDurabilityDays int    `json:"remaining_durability_days"`
}

type Inventory struct {
	instances []InventoryItemInstance
}

func NewInventory() Inventory {
	return Inventory{instances: []InventoryItemInstance{}}
}

func NewInventoryFromEntries(entries []InventoryItemInstance) Inventory {
	items := make([]InventoryItemInstance, 0, len(entries))
	for _, e := range entries {
		if e.Quantity <= 0 {
			continue
		}
		if e.RemainingDurabilityDays < 0 {
			e.RemainingDurabilityDays = 0
		}
		items = append(items, e)
	}
	sortInstances(items)
	return Inventory{instances: items}
}

func (i Inventory) Entries() []InventoryItemInstance {
	out := make([]InventoryItemInstance, len(i.instances))
	copy(out, i.instances)
	sortInstances(out)
	return out
}

func (i Inventory) Quantity(itemID ItemID) int {
	total := 0
	for _, inst := range i.instances {
		if inst.ItemID == itemID {
			total += inst.Quantity
		}
	}
	return total
}

func (i Inventory) ActiveQuantity(itemID ItemID) int {
	total := 0
	for _, inst := range i.instances {
		if inst.ItemID == itemID && inst.RemainingDurabilityDays > 0 {
			total += inst.Quantity
		}
	}
	return total
}

func (i Inventory) Add(item Item, day int, quantity int) Inventory {
	if quantity <= 0 {
		return i
	}
	next := i.Entries()
	next = append(next, InventoryItemInstance{
		ItemID:                  item.ItemID,
		Quantity:                quantity,
		AcquiredAtDay:           day,
		RemainingDurabilityDays: item.DurabilityDays,
	})
	sortInstances(next)
	return Inventory{instances: next}
}

func (i Inventory) DecrementDurabilityByDay() Inventory {
	next := i.Entries()
	for idx := range next {
		if next[idx].RemainingDurabilityDays > 0 {
			next[idx].RemainingDurabilityDays--
		}
	}
	return Inventory{instances: next}
}

func (i Inventory) ActiveEffects(c Catalog) Effects {
	total := Effects{}
	for _, inst := range i.instances {
		if inst.RemainingDurabilityDays <= 0 {
			continue
		}
		item, ok := c.Item(inst.ItemID)
		if !ok {
			continue
		}
		total.ProductivityDelta += item.Effects.ProductivityDelta * inst.Quantity
		total.MoraleDelta += item.Effects.MoraleDelta * inst.Quantity
		total.TechDebtDelta += item.Effects.TechDebtDelta * inst.Quantity
		total.ReputationDelta += item.Effects.ReputationDelta * inst.Quantity
	}
	return total
}

func sortInstances(items []InventoryItemInstance) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].ItemID != items[j].ItemID {
			return items[i].ItemID < items[j].ItemID
		}
		if items[i].AcquiredAtDay != items[j].AcquiredAtDay {
			return items[i].AcquiredAtDay < items[j].AcquiredAtDay
		}
		if items[i].RemainingDurabilityDays != items[j].RemainingDurabilityDays {
			return items[i].RemainingDurabilityDays < items[j].RemainingDurabilityDays
		}
		return items[i].Quantity < items[j].Quantity
	})
}
