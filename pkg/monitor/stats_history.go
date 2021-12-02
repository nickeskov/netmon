package monitor

import (
	"fmt"
	"time"

	"github.com/gammazero/deque"
)

type statsDataSnapshot struct {
	snapshotCreationTime time.Time
	nodes                nodesWithStats
	maxHeight            int
	nodesDownCriterion   bool
	heightCriterion      bool
	stateHashCriterion   bool
}

func (s *statsDataSnapshot) String() string {
	if s == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"(snapshotCreationTime: %s, maxHeight: %d, nodesDownCriterion: %t, heightCriterion: %t, stateHashCriterion: %t)",
		s.snapshotCreationTime,
		s.maxHeight,
		s.nodesDownCriterion,
		s.heightCriterion,
		s.stateHashCriterion,
	)
}

type statsHistoryDeque struct {
	maxLen int
	deque  *deque.Deque
}

func newStatsDeque(maxLen int) statsHistoryDeque {
	return statsHistoryDeque{
		maxLen: maxLen,
		deque:  deque.New(maxLen),
	}
}

// PushFront pushes element to beginning of the deque and returns popped element from back of the deque.
func (d *statsHistoryDeque) PushFront(snapshot *statsDataSnapshot) (back *statsDataSnapshot) {
	if snapshot == nil {
		panic("statsHistoryDeque: push front <nil> data")
	}
	if d.deque.Len() >= d.maxLen {
		back = d.deque.PopBack().(*statsDataSnapshot)
	}
	d.deque.PushFront(snapshot)
	return back
}

func (d *statsHistoryDeque) Front() *statsDataSnapshot {
	return d.deque.Front().(*statsDataSnapshot)
}

func (d *statsHistoryDeque) Back() *statsDataSnapshot {
	return d.deque.Back().(*statsDataSnapshot)
}

func (d *statsHistoryDeque) At(i int) *statsDataSnapshot {
	return d.deque.At(i).(*statsDataSnapshot)
}

func (d *statsHistoryDeque) Len() int {
	return d.deque.Len()
}

func (d *statsHistoryDeque) Clear() {
	d.deque.Clear()
}
