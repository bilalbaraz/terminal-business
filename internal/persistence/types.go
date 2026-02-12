package persistence

import "time"

type SaveIdentity struct {
	SaveID       string    `json:"save_id"`
	CompanyID    string    `json:"company_id"`
	CompanyName  string    `json:"company_name"`
	CompanyType  string    `json:"company_type"`
	CreatedAt    time.Time `json:"created_at"`
	LastPlayedAt time.Time `json:"last_played_at"`
	Version      int       `json:"version"`
}

type SaveFile struct {
	SaveIdentity        SaveIdentity   `json:"save_identity"`
	SimulationSeed      int64          `json:"simulation_seed"`
	TickCounter         int64          `json:"tick_counter"`
	DomainStateSnapshot map[string]any `json:"domain_state_snapshot"`
	Version             int            `json:"version"`
	Checksum            string         `json:"checksum"`
}

type SaveIndexEntry struct {
	SaveID       string    `json:"save_id"`
	CompanyName  string    `json:"company_name"`
	CompanyType  string    `json:"company_type"`
	LastPlayedAt time.Time `json:"last_played_at"`
	SaveFilePath string    `json:"save_file_path"`
	Version      int       `json:"version"`
}

type SaveIndex struct {
	Version int              `json:"version"`
	Entries []SaveIndexEntry `json:"entries"`
}
