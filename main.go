package main

import (
	"image/color"
	"log"
	
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Game implements ebiten.Game interface.
type Game struct {
	points			[]struct{ x, y float64}
	revealing		bool
	revealIndex int
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.points = append(g.points, struct{ x, y float64 }{float64(x), float64(y)})
	}

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	lineColor := color.White

	// Draw lines
	for i:=1; i<len(g.points); i++ {
		ebitenutil.DrawLine(screen, g.points[i-1].x, g.points[i-1].y, g.points[i].x, g.points[i].y, lineColor)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return 640, 480
}

func main() {
	game := &Game{}

	// Set the Ebiten game parameters.
	ebiten.SetWindowTitle("Fourier Board")
	ebiten.SetWindowSize(640, 480)

	// Run the game.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}