package store

import "sort"

type ItemID string

const (
	ItemDesk     ItemID = "desk"
	ItemChair    ItemID = "chair"
	ItemComputer ItemID = "computer"
)

type Effects struct {
	ProductivityDelta int `json:"productivity_delta"`
	MoraleDelta       int `json:"morale_delta"`
	TechDebtDelta     int `json:"tech_debt_delta"`
	ReputationDelta   int `json:"reputation_delta"`
}

type Item struct {
	ItemID      ItemID  `json:"item_id"`
	DisplayName string  `json:"display_name"`
	Category    string  `json:"category"`
	Price       int     `json:"price"`
	Effects     Effects `json:"effects"`
	MaxOwned    int     `json:"max_owned"`
}

type CatalogConfig struct {
	Items map[ItemID]Item
}

type Catalog struct {
	items map[ItemID]Item
	order []ItemID
}

func DefaultCatalogConfig() CatalogConfig {
	return CatalogConfig{Items: map[ItemID]Item{
		ItemDesk: {
			ItemID:      ItemDesk,
			DisplayName: "Desk",
			Category:    "Office",
			Price:       150,
			Effects: Effects{
				ProductivityDelta: 2,
				MoraleDelta:       1,
			},
			MaxOwned: 1,
		},
		ItemChair: {
			ItemID:      ItemChair,
			DisplayName: "Chair",
			Category:    "Office",
			Price:       100,
			Effects: Effects{
				ProductivityDelta: 1,
				MoraleDelta:       2,
			},
			MaxOwned: 1,
		},
		ItemComputer: {
			ItemID:      ItemComputer,
			DisplayName: "Computer",
			Category:    "Office",
			Price:       500,
			Effects: Effects{
				ProductivityDelta: 5,
			},
			MaxOwned: 1,
		},
	}}
}

func NewCatalog(cfg CatalogConfig) Catalog {
	items := map[ItemID]Item{}
	for id, item := range cfg.Items {
		item.ItemID = id
		items[id] = item
	}
	order := make([]ItemID, 0, len(items))
	for id := range items {
		order = append(order, id)
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	return Catalog{items: items, order: order}
}

func DefaultCatalog() Catalog {
	return NewCatalog(DefaultCatalogConfig())
}

func (c Catalog) Item(id ItemID) (Item, bool) {
	item, ok := c.items[id]
	return item, ok
}

func (c Catalog) OrderedItems() []Item {
	items := make([]Item, 0, len(c.order))
	for _, id := range c.order {
		items = append(items, c.items[id])
	}
	return items
}
