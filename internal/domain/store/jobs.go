package store

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
)

type JobStatus string

const (
	JobAvailable JobStatus = "available"
	JobActive    JobStatus = "active"
	JobCompleted JobStatus = "completed"
)

type JobOffer struct {
	JobID           string    `json:"job_id"`
	Title           string    `json:"title"`
	Payout          int       `json:"payout"`
	DurationDays    int       `json:"duration_days"`
	GeneratedAtTick int       `json:"generated_at_tick"`
	Status          JobStatus `json:"status"`
}

type ActiveJob struct {
	JobID          string    `json:"job_id"`
	AssignedTo     string    `json:"assigned_to"`
	AcceptedAtTick int       `json:"accepted_at_tick"`
	DueAtTick      int       `json:"due_at_tick"`
	Payout         int       `json:"payout"`
	Title          string    `json:"title"`
	Status         JobStatus `json:"status"`
}

type CompletedJob struct {
	JobID           string `json:"job_id"`
	AssignedTo      string `json:"assigned_to"`
	CompletedAtTick int    `json:"completed_at_tick"`
	Payout          int    `json:"payout"`
	Title           string `json:"title"`
}

var (
	ErrNotOperational  = errors.New("company is not operational")
	ErrCapacityFull    = errors.New("active job capacity full")
	ErrAssigneeBusy    = errors.New("assignee already has an active job")
	ErrJobNotFound     = errors.New("job offer not found")
	ErrInvalidJobOffer = errors.New("invalid job offer")
)

func Headcount(state GameState) int {
	if state.Headcount <= 0 {
		return 1
	}
	return state.Headcount
}

func MarketOffersForTick(seed int64, tick int, size int) []JobOffer {
	if size <= 0 {
		size = 5
	}
	if size > 10 {
		size = 10
	}
	r := rand.New(rand.NewSource(seedForPurpose(seed, tick, "market")))
	prefix := []string{"Landing Page", "Bug Fix", "Data Cleanup", "Prototype", "API Integration", "Automation", "UI Polish", "Docs Sprint"}
	suffix := []string{"for Startup", "for Local Shop", "for Creator", "for Team", "for Agency", "for Merchant"}
	offers := make([]JobOffer, 0, size)
	for i := 0; i < size; i++ {
		title := fmt.Sprintf("%s %s", prefix[r.Intn(len(prefix))], suffix[r.Intn(len(suffix))])
		duration := 1 + r.Intn(4)
		payout := 80 + r.Intn(420)
		offers = append(offers, JobOffer{
			JobID:           fmt.Sprintf("mkt-%d-%02d", tick, i),
			Title:           title,
			Payout:          payout,
			DurationDays:    duration,
			GeneratedAtTick: tick,
			Status:          JobAvailable,
		})
	}
	return offers
}

func BestROISoonestIndex(offers []JobOffer) int {
	if len(offers) == 0 {
		return -1
	}
	bestIdx := 0
	bestScore := scoreROI(offers[0])
	for i := 1; i < len(offers); i++ {
		s := scoreROI(offers[i])
		if s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	return bestIdx
}

func AcceptJob(state GameState, offers []JobOffer, jobID, assignee string, tick int) (GameState, error) {
	if !IsOperational(state) {
		return GameState{}, ErrNotOperational
	}
	offer, ok := findOffer(offers, jobID)
	if !ok {
		return GameState{}, ErrJobNotFound
	}
	if offer.DurationDays <= 0 || offer.Payout <= 0 {
		return GameState{}, ErrInvalidJobOffer
	}
	if assignee == "" {
		assignee = "founder"
	}
	if len(state.ActiveJobs) >= Headcount(state) {
		return GameState{}, ErrCapacityFull
	}
	for _, active := range state.ActiveJobs {
		if active.AssignedTo == assignee && active.Status == JobActive {
			return GameState{}, ErrAssigneeBusy
		}
	}
	next := state
	next.ActiveJobs = append(next.ActiveJobs, ActiveJob{
		JobID:          offer.JobID,
		AssignedTo:     assignee,
		AcceptedAtTick: tick,
		DueAtTick:      tick + offer.DurationDays,
		Payout:         offer.Payout,
		Title:          offer.Title,
		Status:         JobActive,
	})
	sort.Slice(next.ActiveJobs, func(i, j int) bool {
		if next.ActiveJobs[i].DueAtTick != next.ActiveJobs[j].DueAtTick {
			return next.ActiveJobs[i].DueAtTick < next.ActiveJobs[j].DueAtTick
		}
		if next.ActiveJobs[i].AssignedTo != next.ActiveJobs[j].AssignedTo {
			return next.ActiveJobs[i].AssignedTo < next.ActiveJobs[j].AssignedTo
		}
		return next.ActiveJobs[i].JobID < next.ActiveJobs[j].JobID
	})
	return next, nil
}

func ActiveJobCountdown(currentTick int, job ActiveJob) int {
	return job.DueAtTick - currentTick
}

func CompleteDueJobs(state GameState, tick int) GameState {
	next := state
	stillActive := make([]ActiveJob, 0, len(state.ActiveJobs))
	for _, active := range state.ActiveJobs {
		if active.Status != JobActive {
			continue
		}
		if tick >= active.DueAtTick {
			next.Cash += active.Payout
			next.CompletedJobs = append(next.CompletedJobs, CompletedJob{
				JobID:           active.JobID,
				AssignedTo:      active.AssignedTo,
				CompletedAtTick: tick,
				Payout:          active.Payout,
				Title:           active.Title,
			})
			continue
		}
		stillActive = append(stillActive, active)
	}
	next.ActiveJobs = stillActive
	sort.Slice(next.CompletedJobs, func(i, j int) bool {
		if next.CompletedJobs[i].CompletedAtTick != next.CompletedJobs[j].CompletedAtTick {
			return next.CompletedJobs[i].CompletedAtTick < next.CompletedJobs[j].CompletedAtTick
		}
		return next.CompletedJobs[i].JobID < next.CompletedJobs[j].JobID
	})
	return next
}

func findOffer(offers []JobOffer, jobID string) (JobOffer, bool) {
	for _, offer := range offers {
		if offer.JobID == jobID {
			return offer, true
		}
	}
	return JobOffer{}, false
}

func scoreROI(offer JobOffer) float64 {
	if offer.DurationDays <= 0 {
		return -1
	}
	return float64(offer.Payout) / float64(offer.DurationDays)
}

func seedForPurpose(seed int64, tick int, purpose string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%s:%d:%d", purpose, seed, tick)))
	return int64(h.Sum64())
}
