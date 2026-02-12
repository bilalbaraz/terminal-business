package store

import (
	"errors"
	"testing"
)

func TestMarketOffersDeterministicAndTickVariant(t *testing.T) {
	offersA := MarketOffersForTick(42, 3, 6)
	offersB := MarketOffersForTick(42, 3, 6)
	offersC := MarketOffersForTick(42, 4, 6)
	if len(offersA) != 6 || len(offersB) != 6 {
		t.Fatalf("unexpected size")
	}
	for i := range offersA {
		if offersA[i] != offersB[i] {
			t.Fatalf("non deterministic offers at %d", i)
		}
	}
	same := true
	for i := range offersA {
		if offersA[i] != offersC[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("expected different tick to vary offers")
	}
}

func TestMarketOfferSizeBoundsAndBestROI(t *testing.T) {
	if len(MarketOffersForTick(1, 1, 0)) != 5 {
		t.Fatal("expected min default size")
	}
	if len(MarketOffersForTick(1, 1, 99)) != 10 {
		t.Fatal("expected max capped size")
	}
	idx := BestROISoonestIndex([]JobOffer{
		{Payout: 100, DurationDays: 5},
		{Payout: 99, DurationDays: 1},
		{Payout: 120, DurationDays: 2},
	})
	if idx != 1 {
		t.Fatalf("best idx got %d", idx)
	}
	if BestROISoonestIndex(nil) != -1 {
		t.Fatal("expected -1 on empty")
	}
}

func TestAcceptJobGatingAndCapacity(t *testing.T) {
	catalog := DefaultCatalog()
	state := NewInitialState(1000, catalog, DefaultEconomyConfig())
	offers := []JobOffer{{JobID: "j1", Title: "A", Payout: 100, DurationDays: 2}}
	if _, err := AcceptJob(state, offers, "j1", "founder", 1); !errors.Is(err, ErrNotOperational) {
		t.Fatalf("got %v", err)
	}

	state.CompanyInventory = NewInventoryFromEntries([]InventoryItemInstance{
		{ItemID: ItemDesk, Quantity: 1, RemainingDurabilityDays: 1},
		{ItemID: ItemChair, Quantity: 1, RemainingDurabilityDays: 1},
		{ItemID: ItemComputer, Quantity: 1, RemainingDurabilityDays: 1},
	})
	next, err := AcceptJob(state, offers, "j1", "founder", 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(next.ActiveJobs) != 1 {
		t.Fatalf("got %d", len(next.ActiveJobs))
	}
	if next.ActiveJobs[0].DueAtTick != 3 {
		t.Fatalf("due tick got %d", next.ActiveJobs[0].DueAtTick)
	}
	if ActiveJobCountdown(1, next.ActiveJobs[0]) != 2 {
		t.Fatalf("countdown got %d", ActiveJobCountdown(1, next.ActiveJobs[0]))
	}

	if _, err := AcceptJob(next, offers, "j1", "founder", 1); !errors.Is(err, ErrCapacityFull) {
		t.Fatalf("got %v", err)
	}

	next.Headcount = 2
	next.ActiveJobs = []ActiveJob{{JobID: "x", AssignedTo: "founder", DueAtTick: 9, Status: JobActive}}
	if _, err := AcceptJob(next, []JobOffer{{JobID: "j2", Title: "B", Payout: 120, DurationDays: 2}}, "j2", "founder", 1); !errors.Is(err, ErrAssigneeBusy) {
		t.Fatalf("got %v", err)
	}
	if _, err := AcceptJob(next, []JobOffer{{JobID: "j2", Title: "B", Payout: 120, DurationDays: 2}}, "j2", "", 1); !errors.Is(err, ErrAssigneeBusy) {
		t.Fatalf("got %v", err)
	}
}

func TestAcceptJobValidationBranchesAndSorting(t *testing.T) {
	catalog := DefaultCatalog()
	state := NewInitialState(1000, catalog, DefaultEconomyConfig())
	state.CompanyInventory = NewInventoryFromEntries([]InventoryItemInstance{
		{ItemID: ItemDesk, Quantity: 1, RemainingDurabilityDays: 1},
		{ItemID: ItemChair, Quantity: 1, RemainingDurabilityDays: 1},
		{ItemID: ItemComputer, Quantity: 1, RemainingDurabilityDays: 1},
	})
	if _, err := AcceptJob(state, nil, "missing", "a", 1); !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("got %v", err)
	}
	if _, err := AcceptJob(state, []JobOffer{{JobID: "j", Payout: 0, DurationDays: 1}}, "j", "a", 1); !errors.Is(err, ErrInvalidJobOffer) {
		t.Fatalf("got %v", err)
	}
	if _, err := AcceptJob(state, []JobOffer{{JobID: "j", Payout: 1, DurationDays: 0}}, "j", "a", 1); !errors.Is(err, ErrInvalidJobOffer) {
		t.Fatalf("got %v", err)
	}

	next := state
	next.Headcount = 3
	next.ActiveJobs = []ActiveJob{{JobID: "b", AssignedTo: "z", DueAtTick: 5, Status: JobActive}}
	next, err := AcceptJob(next, []JobOffer{{JobID: "a", Title: "A", Payout: 10, DurationDays: 1}}, "a", "a", 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	next, err = AcceptJob(next, []JobOffer{{JobID: "c", Title: "C", Payout: 10, DurationDays: 3}}, "c", "b", 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if next.ActiveJobs[0].DueAtTick > next.ActiveJobs[1].DueAtTick {
		t.Fatal("expected sorted active jobs by due tick")
	}

	// Cover sort tie-breakers by assignee then job id.
	tie := state
	tie.Headcount = 5
	tie.ActiveJobs = []ActiveJob{
		{JobID: "z", AssignedTo: "a", DueAtTick: 2, Status: JobCompleted},
		{JobID: "m", AssignedTo: "a", DueAtTick: 2, Status: JobCompleted},
	}
	tie, err = AcceptJob(tie, []JobOffer{{JobID: "a", Title: "A", Payout: 10, DurationDays: 1}}, "a", "b", 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tie.ActiveJobs[0].AssignedTo != "a" || tie.ActiveJobs[1].AssignedTo != "a" {
		t.Fatalf("expected assignee tie sort first %+v", tie.ActiveJobs)
	}
	if tie.ActiveJobs[0].JobID != "m" || tie.ActiveJobs[1].JobID != "z" {
		t.Fatalf("expected job id tie sort %+v", tie.ActiveJobs)
	}
}

func TestCompleteDueJobsPayoutOnceAndIdempotent(t *testing.T) {
	state := GameState{
		Cash: 100,
		ActiveJobs: []ActiveJob{
			{JobID: "a", AssignedTo: "founder", DueAtTick: 2, Payout: 50, Title: "A", Status: JobActive},
			{JobID: "b", AssignedTo: "founder2", DueAtTick: 5, Payout: 70, Title: "B", Status: JobActive},
			{JobID: "c", AssignedTo: "founder3", DueAtTick: 1, Payout: 10, Title: "C", Status: JobCompleted},
		},
	}
	next := CompleteDueJobs(state, 2)
	if next.Cash != 150 {
		t.Fatalf("cash got %d", next.Cash)
	}
	if len(next.ActiveJobs) != 1 || next.ActiveJobs[0].JobID != "b" {
		t.Fatalf("unexpected active jobs %+v", next.ActiveJobs)
	}
	if len(next.CompletedJobs) != 1 || next.CompletedJobs[0].JobID != "a" {
		t.Fatalf("unexpected completed jobs %+v", next.CompletedJobs)
	}
	again := CompleteDueJobs(next, 2)
	if again.Cash != next.Cash {
		t.Fatalf("cash changed on idempotent pass: %d -> %d", next.Cash, again.Cash)
	}

	// Cover completed history sort tie-breakers.
	unsorted := GameState{
		CompletedJobs: []CompletedJob{
			{JobID: "z", CompletedAtTick: 3},
			{JobID: "a", CompletedAtTick: 3},
			{JobID: "m", CompletedAtTick: 2},
		},
	}
	sorted := CompleteDueJobs(unsorted, 1)
	if sorted.CompletedJobs[0].JobID != "m" || sorted.CompletedJobs[1].JobID != "a" || sorted.CompletedJobs[2].JobID != "z" {
		t.Fatalf("unexpected completed sort %+v", sorted.CompletedJobs)
	}
}

func TestHeadcountAndSeedHelpers(t *testing.T) {
	if Headcount(GameState{}) != 1 {
		t.Fatal("default headcount should be 1")
	}
	if Headcount(GameState{Headcount: 3}) != 3 {
		t.Fatal("headcount should use explicit value")
	}
	if seedForPurpose(1, 2, "x") == seedForPurpose(1, 3, "x") {
		t.Fatal("seed should differ with tick")
	}
	if scoreROI(JobOffer{Payout: 100, DurationDays: 4}) != 25 {
		t.Fatal("roi mismatch")
	}
	if scoreROI(JobOffer{Payout: 100, DurationDays: 0}) != -1 {
		t.Fatal("expected invalid roi")
	}
	if _, ok := findOffer([]JobOffer{{JobID: "a"}}, "a"); !ok {
		t.Fatal("expected offer found")
	}
	if _, ok := findOffer([]JobOffer{{JobID: "a"}}, "b"); ok {
		t.Fatal("expected offer missing")
	}
}
