package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type GameState int
const (
	Start GameState = iota
	Drawing
	Revealing
	Computing
	Fourier
	End
)

// Required from Ebiten.
// Game implements ebiten.Game interface.
type Game struct {
	windowSize  struct{ width, height int }
	points			[]struct{ x, y float64}
	state				GameState
	revealIndex int
}

// Required from Ebiten.
// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	switch g.state {
	case Start:
		g.state = Drawing
	case Drawing:
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			dim := len(g.points)
			if (dim==0 || float64(x)!=g.points[dim-1].x || float64(y)!=g.points[dim-1].y) {
				g.points = append(g.points, struct{ x, y float64 }{float64(x), float64(y)})
			}
		}
		if ebiten.IsKeyPressed(ebiten.KeyN) {
			g.state = Revealing
			g.revealIndex = 1
		}
	case Revealing:
		if  g.revealIndex<len(g.points)-1 {
			g.revealIndex++
		} else {
			g.state = Computing
		}
	case Computing:
		g.state = Drawing
	}

	return nil
}

func drawEmptyCircle(screen *ebiten.Image, cx, cy, r float64, lineColor color.Color) {
	steps := 20
	dAngle := 2*math.Pi/float64(steps)

	point1 := struct{ x,y float64 }{cx+r, cy}
	point2 := struct{ x,y float64 }{0, 0}
	for i:=1; i<=steps; i++ {
		point2.x = cx+r*math.Cos(dAngle*float64(i))
		point2.y = cy+r*math.Sin(dAngle*float64(i))
		ebitenutil.DrawLine(screen, point1.x, point1.y, point2.x, point2.y, lineColor)
		point1 = point2
	}
}

// Required from Ebiten.
// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	lineColor := color.White

	// Draw lines
	upperLimit := 0
	if g.state == Revealing {
		upperLimit = g.revealIndex
	}	else {
		upperLimit = len(g.points)
	}

	for i:=1; i<upperLimit; i++ {
		ebitenutil.DrawLine(screen, g.points[i-1].x, g.points[i-1].y, g.points[i].x, g.points[i].y, lineColor)
		drawEmptyCircle(screen, g.points[i].x, g.points[i].y, 10, lineColor)
	}
}

// Required from Ebiten.
// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
  return g.windowSize.width, g.windowSize.height
}

func main() {
	game := &Game{}
	game.state = Start
	game.windowSize = struct{ width, height int }{1920, 1080}

	// Set the Ebiten game parameters.
	ebiten.SetWindowTitle("Fourier Board")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(game.windowSize.width, game.windowSize.height)

	// Run the game.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}