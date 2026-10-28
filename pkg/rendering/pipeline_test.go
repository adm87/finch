package rendering_test

import (
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/transform"
)

type drawn struct {
	name  string
	pass  int
	depth int32
}

// layeredAdapter splits each entity into one job per depth, like a tile map
// with several layers, and records what it draws.
type layeredAdapter struct {
	drawn  *[]drawn
	names  map[donburi.Entity]string
	depths []int32
	bounds []geom.Rect // per pass, empty uses the entity's
}

func (a *layeredAdapter) Jobs(c *rendering.Candidate, jobs []rendering.Job) []rendering.Job {
	for pass, depth := range a.depths {
		job := c.Job()
		job.Pass, job.Depth = pass, depth
		if pass < len(a.bounds) && !a.bounds[pass].Empty() {
			job.Bounds = a.bounds[pass]
		}
		jobs = append(jobs, job)
	}
	return jobs
}

func (a *layeredAdapter) Render(_ *ebiten.Image, job *rendering.Job, _ geom.Rect, _ geom.Matrix) {
	*a.drawn = append(*a.drawn, drawn{name: a.names[job.Entry.Entity()], pass: job.Pass, depth: job.Depth})
}

var view = geom.Rect{Width: 100, Height: 100}

func newEntry(t *testing.T, w donburi.World, id rendering.AdapterID, rect geom.Rect) *donburi.Entry {
	t.Helper()
	e := w.Entry(w.Create(rendering.RendererComponent, transform.TransformComponent, transform.BoundsComponent))
	r, _ := rendering.GetRenderer(e)
	r.SetRendererID(id)
	b, _ := transform.GetBounds(e)
	b.SetRect(rect)
	return e
}

func TestPipelineInterleavesJobsFromOneEntity(t *testing.T) {
	w := donburi.NewWorld()
	p := rendering.NewPipeline()
	var got []drawn
	names := map[donburi.Entity]string{}

	mapID, err := p.RegisterRenderer(&layeredAdapter{drawn: &got, names: names, depths: []int32{0, 20}})
	require.NoError(t, err)
	spriteID, err := p.RegisterRenderer(&layeredAdapter{drawn: &got, names: names, depths: []int32{10}})
	require.NoError(t, err)

	tiles := newEntry(t, w, mapID, view)
	player := newEntry(t, w, spriteID, geom.Rect{X: 10, Y: 10, Width: 8, Height: 8})
	names[tiles.Entity()], names[player.Entity()] = "map", "player"

	require.NoError(t, p.Draw(nil, []*donburi.Entry{tiles, player}, view, geom.Matrix{}))
	require.Equal(t, []drawn{
		{name: "map", pass: 0, depth: 0},
		{name: "player", pass: 0, depth: 10},
		{name: "map", pass: 1, depth: 20},
	}, got)
}

func TestPipelineCullsEachJob(t *testing.T) {
	w := donburi.NewWorld()
	p := rendering.NewPipeline()
	var got []drawn

	id, err := p.RegisterRenderer(&layeredAdapter{
		drawn:  &got,
		names:  map[donburi.Entity]string{},
		depths: []int32{0, 1},
		bounds: []geom.Rect{{}, {X: 500, Y: 500, Width: 10, Height: 10}},
	})
	require.NoError(t, err)

	e := newEntry(t, w, id, view)
	require.NoError(t, p.Draw(nil, []*donburi.Entry{e}, view, geom.Matrix{}))
	require.Equal(t, []drawn{{pass: 0, depth: 0}}, got)
}

func TestPipelineSkipsHiddenAndReportsUnknownAdapters(t *testing.T) {
	w := donburi.NewWorld()
	p := rendering.NewPipeline()
	var got []drawn

	id, err := p.RegisterRenderer(&layeredAdapter{drawn: &got, names: map[donburi.Entity]string{}, depths: []int32{0}})
	require.NoError(t, err)

	hidden := newEntry(t, w, id, view)
	r, _ := rendering.GetRenderer(hidden)
	r.SetVisible(false)
	offscreen := newEntry(t, w, id, geom.Rect{X: 500, Y: 500, Width: 10, Height: 10})
	unknown := newEntry(t, w, id+1, view)

	require.Error(t, p.Draw(nil, []*donburi.Entry{hidden, offscreen, unknown}, view, geom.Matrix{}))
	require.Empty(t, got)
}
