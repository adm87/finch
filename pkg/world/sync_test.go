package world

import (
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/transform"
)

func newSyncWorld(t *testing.T) (*World, *donburi.Entry) {
	t.Helper()
	w := NewWorld(WithPartitionResolutions(ptResolutions...))
	entry := w.Create()
	transform.BoundsComponent.Get(entry).SetRect(ptRect(0, 0, 10, 10))
	return w, entry
}

// found reports whether a 10×10 query at x, y returns entry.
func found(w *World, x, y float64, entry *donburi.Entry) bool {
	for _, e := range w.Query(ptRect(x, y, 10, 10), nil) {
		if e.Entity() == entry.Entity() {
			return true
		}
	}
	return false
}

func TestSyncIndexesNewEntities(t *testing.T) {
	w, entry := newSyncWorld(t)
	require.False(t, found(w, 0, 0, entry))

	require.NoError(t, w.Sync())
	require.True(t, found(w, 0, 0, entry))
}

func TestSyncReindexesMovedEntities(t *testing.T) {
	w, entry := newSyncWorld(t)
	require.NoError(t, w.Sync())

	ptMove(entry, 500, 500)
	require.True(t, found(w, 0, 0, entry))

	require.NoError(t, w.Sync())
	require.False(t, found(w, 0, 0, entry))
	require.True(t, found(w, 500, 500, entry))
}

func TestSyncReindexesChangedBounds(t *testing.T) {
	w, entry := newSyncWorld(t)
	require.NoError(t, w.Sync())

	transform.BoundsComponent.Get(entry).SetRect(ptRect(0, 0, 200, 200))
	require.False(t, found(w, 150, 150, entry))

	require.NoError(t, w.Sync())
	require.True(t, found(w, 150, 150, entry))
}

func TestSyncSkipsUnchangedEntities(t *testing.T) {
	w, entry := newSyncWorld(t)
	require.NoError(t, w.Sync())

	idx, _ := getIndexing(entry)
	require.NoError(t, w.partitions.grids[idx.partition].Delete(idx.key))

	require.NoError(t, w.Sync())
	require.False(t, found(w, 0, 0, entry))

	ptMove(entry, 1, 1)
	require.NoError(t, w.Sync())
	require.True(t, found(w, 0, 0, entry))
}

func TestSyncSkipsUnchangedAfterManualUpdate(t *testing.T) {
	w, entry := newSyncWorld(t)
	require.NoError(t, w.Add(entry))

	idx, _ := getIndexing(entry)
	tr := transform.TransformComponent.Get(entry)
	b := transform.BoundsComponent.Get(entry)
	require.Equal(t, tr.Version(), idx.transformVersion)
	require.Equal(t, b.Version(), idx.boundsVersion)
}

func TestSyncIgnoresDontPartition(t *testing.T) {
	w := NewWorld(WithPartitionResolutions(ptResolutions...))
	entry := w.Create(DontPartition)
	transform.BoundsComponent.Get(entry).SetRect(ptRect(0, 0, 10, 10))

	require.NoError(t, w.Sync())
	require.False(t, found(w, 0, 0, entry))
}

func TestUpdateRemovesEntitiesTaggedDontPartitionLater(t *testing.T) {
	w, entry := newSyncWorld(t)
	require.NoError(t, w.Sync())

	entry.AddComponent(DontPartition)
	require.NoError(t, w.Sync())
	require.True(t, found(w, 0, 0, entry))

	require.NoError(t, w.Update(entry))
	require.False(t, found(w, 0, 0, entry))
}
