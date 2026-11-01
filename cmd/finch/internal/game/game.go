package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/adm87/finch/content/embedded"
)

type Game struct {
	img                       *ebiten.Image
	x, y                      float64
	screenWidth, screenHeight int
}

func NewGame(screenWidth, screenHeight int) (*Game, error) {
	img, _, err := ebitenutil.NewImageFromFileSystem(embedded.Images, embedded.DefaultImage)
	if err != nil {
		return nil, err
	}
	return &Game{
		img:          img,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}, nil
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.y--
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.y++
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.x--
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.x++
	}

	width, height := g.img.Bounds().Dx(), g.img.Bounds().Dy()

	g.x = min(max(g.x, 0), float64(g.screenWidth-width))
	g.y = min(max(g.y, 0), float64(g.screenHeight-height))

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.x, g.y)
	screen.DrawImage(g.img, op)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return g.screenWidth, g.screenHeight
}
