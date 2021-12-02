package monitor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatsHistoryDeque(t *testing.T) {
	const pushedDataLen = 20

	pushedData := make([]statsDataSnapshot, pushedDataLen)
	for i := range pushedData {
		pushedData[i] = statsDataSnapshot{maxHeight: i + 1}
	}
	d := newStatsDeque(pushedDataLen)
	for i := len(pushedData) - 1; i >= 0; i-- {
		d.PushFront(&pushedData[i])
	}

	const (
		expectedPoppedMaxHeight = 20
		expectedFrontMaxHeight  = 132114234
		expectedBackMaxHeight   = 19
	)

	back := d.PushFront(&statsDataSnapshot{maxHeight: expectedFrontMaxHeight})

	// len hasn't been changed
	require.Equal(t, len(pushedData), d.Len())
	// popped elem is correct
	require.Equal(t, expectedPoppedMaxHeight, back.maxHeight)
	// current FRONT element is correct
	require.Equal(t, expectedFrontMaxHeight, d.Front().maxHeight)
	require.Equal(t, expectedFrontMaxHeight, d.At(0).maxHeight)
	// current BACK element is correct
	require.Equal(t, expectedBackMaxHeight, d.Back().maxHeight)
	require.Equal(t, expectedBackMaxHeight, d.At(d.Len()-1).maxHeight)

	d.Clear()
	require.Equal(t, d.Len(), 0)

	require.Panics(t, func() {
		d.PushFront(nil)
	})
}
