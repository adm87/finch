package world

import (
	"slices"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/transform"
)

func TestPartitionCells(t *testing.T) {
	w := NewWorld(WithPartitionResolutions(ptResolutions...))

	small := w.Create()
	transform.BoundsComponent.Get(small).SetRect(ptRect(0, 0, 10, 10))

	large := w.Create()
	transform.BoundsComponent.Get(large).SetRect(ptRect(0, 0, 50, 50))
	transform.TransformComponent.Get(large).SetPosition(200, 0)

	hidden := w.Create(DontPartition)
	transform.BoundsComponent.Get(hidden).SetRect(ptRect(0, 0, 10, 10))
	transform.TransformComponent.Get(hidden).SetPosition(400, 0)

	require.NoError(t, w.Sync())

	cells := slices.Collect(w.PartitionCells(ptRect(-10, -10, 500, 100)))
	require.ElementsMatch(t, []PartitionCell{
		{Bounds: ptRect(0, 0, 16, 16), Partition: 0, Entities: 1},
		{Bounds: ptRect(192, 0, 64, 64), Partition: 2, Entities: 1},
	}, cells)

	require.Empty(t, slices.Collect(w.PartitionCells(ptRect(1000, 1000, 10, 10))))
}
