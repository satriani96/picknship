// Package store keeps in-memory pick state for the demo.
//
// Production replacement: persist to a database (or call MYOB directly on
// completion). The interface here is intentionally narrow so it can be swapped.
package store

import (
	"sync"

	"github.com/wescoseeds/picknship/internal/data"
	"github.com/wescoseeds/picknship/internal/models"
)

// Pick state for a single order: line UID -> qty picked.
type pickState map[string]float64

// Store is concurrency-safe in-memory pick storage.
type Store struct {
	mu     sync.RWMutex
	picks  map[string]pickState // orderUID -> lineID -> qty picked
	status map[string]string    // orderUID -> "open" | "completed"
}

// New returns an empty store.
func New() *Store {
	return &Store{
		picks:  map[string]pickState{},
		status: map[string]string{},
	}
}

// Get returns the picked qty for a single line. Defaults to 0.
func (s *Store) Get(orderUID, lineID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if ps, ok := s.picks[orderUID]; ok {
		return ps[lineID]
	}
	return 0
}

// GetAll returns picks for an order, with 0s for unset lines.
func (s *Store) GetAll(orderUID string) map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := map[string]float64{}
	if ps, ok := s.picks[orderUID]; ok {
		for k, v := range ps {
			out[k] = v
		}
	}
	return out
}

// Set updates a line's picked qty. Negative values are clamped to zero.
func (s *Store) Set(orderUID, lineID string, qty float64) {
	if qty < 0 {
		qty = 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ps, ok := s.picks[orderUID]
	if !ok {
		ps = pickState{}
		s.picks[orderUID] = ps
	}
	ps[lineID] = qty
}

// Status returns "open" or "completed".
func (s *Store) Status(orderUID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if st, ok := s.status[orderUID]; ok {
		return st
	}
	return "open"
}

// Complete marks an order as completed.
func (s *Store) Complete(orderUID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status[orderUID] = "completed"
}

// PickedOrder returns a copy of the order with QtyOrdered replaced by picked
// qty on each line, and freight unchanged. Used to render the
// invoice/packing-slip after completion. Lines with zero picked qty are kept
// (they appear as back-ordered/zero on the documents) — the demo just shows
// what was picked.
func (s *Store) PickedOrder(o models.Order) models.Order {
	picks := s.GetAll(o.UID)
	out := o
	out.Lines = make([]models.OrderLine, len(o.Lines))
	for i, l := range o.Lines {
		nl := l
		if v, ok := picks[l.LineID]; ok {
			nl.QtyOrdered = v
		} else {
			nl.QtyOrdered = 0
		}
		out.Lines[i] = nl
	}
	return out
}

// Reset clears all state (handy for demoing).
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.picks = map[string]pickState{}
	s.status = map[string]string{}
}

// Orders returns all orders along with simple progress info.
type OrderSummary struct {
	Order        models.Order
	LinesTotal   int
	LinesStarted int
	Status       string
}

// Summaries returns a summary for every demo order.
func (s *Store) Summaries() []OrderSummary {
	out := []OrderSummary{}
	for _, o := range data.All() {
		picks := s.GetAll(o.UID)
		started := 0
		for _, l := range o.Lines {
			if picks[l.LineID] > 0 {
				started++
			}
		}
		out = append(out, OrderSummary{
			Order:        o,
			LinesTotal:   len(o.Lines),
			LinesStarted: started,
			Status:       s.Status(o.UID),
		})
	}
	return out
}
