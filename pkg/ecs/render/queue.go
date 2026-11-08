package render

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

const indexMask = 0xffffffff

// Command holds the information needed to draw an image in a certain layer and depth.
type Command struct {
	Image   *ebiten.Image
	Options ebiten.DrawImageOptions
	Layer   int16
	Depth   int16
}

// Queue holds the draw commands for a frame, sorted by layer and depth.
type Queue struct {
	cmds []Command
	keys []uint64
}

// Push adds a draw command to the queue.
func (q *Queue) Push(c Command) {
	q.keys = append(q.keys, key(c.Layer, c.Depth, len(q.cmds)))
	q.cmds = append(q.cmds, c)
}

func (q *Queue) flush(target *ebiten.Image) {
	slices.Sort(q.keys)
	for _, k := range q.keys {
		c := q.cmds[k&indexMask]
		target.DrawImage(c.Image, &c.Options)
	}
	q.cmds = q.cmds[:0]
	q.keys = q.keys[:0]
}

// key packs layer, depth, and index into a uint64 for sorting.
func key(layer int16, depth int16, index int) uint64 {
	l := uint64(uint16(layer))
	d := uint64(uint16(depth))
	i := uint64(uint32(index))
	return (l << 48) | (d << 32) | i
}
