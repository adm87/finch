package game

import (
	"fmt"

	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/cmd/game/internal/player"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/plugins/aseprite"
)

const (
	gymBundle     = "gym"
	playerLibrary = "captain"
	gym04         = "gym04"
)

func (g *Game) MakeTestScene() error {
	w := g.world

	c := camera.NewMain(w)
	camera.SetZoom(c, 2)
	g.view.SetCamera(c)

	if err := g.LoadBundles(gymBundle); err != nil {
		return err
	}
	src := g.bundles[gymBundle].Source

	mapHandle, err := g.tiled.Load(src, gym04)
	if err != nil {
		return fmt.Errorf("load map: %w", err)
	}
	tm, err := g.tiled.Spawn(mapHandle, w)
	if err != nil {
		return fmt.Errorf("spawn map: %w", err)
	}
	mapInfo, err := g.tiled.Map(mapHandle)
	if err != nil {
		return fmt.Errorf("map %q: %w", gym04, err)
	}
	playerDepth, ok := mapInfo.LayerDepth("player")
	if !ok {
		return fmt.Errorf("map %q has no player layer", gym04)
	}

	libraryID, err := g.aseprite.Load(src, playerLibrary)
	if err != nil {
		return fmt.Errorf("load library %q: %w", playerLibrary, err)
	}
	library, err := g.aseprite.Library(libraryID)
	if err != nil {
		return fmt.Errorf("library %q: %w", playerLibrary, err)
	}

	frame, err := g.images.Frame(library.ImageHandle, 0)
	if err != nil {
		return fmt.Errorf("first frame of %q: %w", playerLibrary, err)
	}

	mb, _ := transform.GetBounds(tm)
	mr := mb.WorldRect(transform.TransformComponent.Get(tm))

	ct := transform.TransformComponent.Get(c)
	ct.SetPosition(mr.Center())

	width := float64(frame.Size.X)
	height := float64(frame.Size.Y)

	// Anchor at the character's feet in the first frame.
	visible := frame.Image.Bounds().Size()
	ax := (float64(frame.Offset.X) + float64(visible.X)/2) / width
	ay := float64(frame.Offset.Y+visible.Y) / height

	entry := w.Create(
		rendering.RendererComponent,
		images.ImageComponent,
		aseprite.AnimatorComponent,
		player.Player,
	)

	t := transform.TransformComponent.Get(entry)
	t.SetPosition(mr.Center())

	r := rendering.RendererComponent.Get(entry)
	r.SetRendererID(g.imageRenderer)
	r.SetDepth(playerDepth)

	img := images.ImageComponent.Get(entry)
	img.SetHandle(library.ImageHandle)
	img.SetAnchor(ax, ay)

	anim := aseprite.AnimatorComponent.Get(entry)
	anim.SetLibrary(libraryID)
	anim.SetAnimation("Idle")

	return nil
}
