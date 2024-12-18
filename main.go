package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"math/cmplx"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"fourier-drawing/fourier"
)

type GameState int
const (
	Preparing GameState = iota
	Start
	Drawing
	Revealing
	Computing
	Fourier
	End
)

type Button struct {
	x, y, width, height float64
	text                []string
	onClick             func(g *Game)
	pressed							bool
}

type ButtonIndex int
const (
	StartButton ButtonIndex = iota
	ClearButton
	FourierButton
)

type Point struct {
	x, y float64
}

// Required from Ebiten.
// Game implements ebiten.Game interface.
type Game struct {
	windowSize  				struct{ width, height int }
	points							[]Point
	state								GameState
	revealIndex 				int
	fourierX						[]complex128
	fourierY						[]complex128
	fourierIndex				int
	fourierPoints				[]Point
	buttons							[]*Button
}

func shiftSequence(sequence []float64, shift float64) {
	for i:=0; i<len(sequence); i++ {
		sequence[i] += shift
	}
}

func drawButton(screen *ebiten.Image, button *Button) {
	borderColor := color.RGBA{255, 255, 255, 255}
	borderWidth := 5.0
	ebitenutil.DrawRect(screen, button.x-borderWidth, button.y-borderWidth, button.width+2*borderWidth, button.height+2*borderWidth, borderColor)

	buttonColor := color.RGBA{0, 0, 0, 255}
	ebitenutil.DrawRect(screen, button.x, button.y, button.width, button.height, buttonColor)

	fontFace := basicfont.Face7x13
	textHeight := 13

	textX := int(button.x)+20
	textY := int(button.y)+20

	d := &font.Drawer{
		Dst:  screen,
		Src:  image.NewUniform(color.White),
		Face: fontFace,
		Dot:  fixed.P(textX, textY),
	}

	for i:=0; i<len(button.text); i++ {
		d.DrawString(button.text[i])
		textY += textHeight
		d.Dot = fixed.P(textX, textY)
	}
}

func drawEmptyCircle(screen *ebiten.Image, cx, cy, r float64, lineColor color.Color) {
	steps := 20
	dAngle := 2*math.Pi/float64(steps)

	point1 := Point{cx+r, cy}
	point2 := Point{0, 0}
	for i:=1; i<=steps; i++ {
		point2.x = cx+r*math.Cos(dAngle*float64(i))
		point2.y = cy+r*math.Sin(dAngle*float64(i))
		ebitenutil.DrawLine(screen, point1.x, point1.y, point2.x, point2.y, lineColor)
		point1 = point2
	}
}

func drawEmptyCircleWithRadius(screen *ebiten.Image, cx, cy, radius, angle float64, lineColor color.Color) (x, y float64) {
	steps := 100
	dAngle := 2*math.Pi/float64(steps)

	point1 := Point{cx+radius, cy}
	point2 := Point{0, 0}
	for i:=1; i<=steps; i++ {
		point2.x = cx+radius*math.Cos(dAngle*float64(i))
		point2.y = cy+radius*math.Sin(dAngle*float64(i))
		ebitenutil.DrawLine(screen, point1.x, point1.y, point2.x, point2.y, color.RGBA{64, 64, 64, 255})
		point1 = point2
	}

	x = cx+radius*math.Cos(angle)
	y = cy-radius*math.Sin(angle)
	ebitenutil.DrawLine(screen, cx, cy, x, y, lineColor)

	return x,y
}

func drawFourierEpicycles(screen *ebiten.Image, fourierSeq []complex128, fourierInd int, startX, startY, phase float64) (x, y float64) {
	N := len(fourierSeq)
	x, y = startX, startY

	for k:=0; k<N; k++ {
		radius := cmplx.Abs(fourierSeq[k])/float64(N)
		arg := 2 * math.Pi * float64(fourierInd) * float64(k) / float64(N) + cmplx.Phase(fourierSeq[k]) + phase;

		x, y = drawEmptyCircleWithRadius(screen, x, y, radius, arg, color.RGBA{150, 150, 150, 255})
	}

	return x, y
}

func (b *Button) CheckIfClicked(g *Game) (pressed bool) {
	pressed = false
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		tempX, tempY := ebiten.CursorPosition()
		mouseX, mouseY := float64(tempX), float64(tempY)
		if mouseX>=b.x && mouseX<=b.x+b.width && mouseY>=b.y && mouseY<=b.y+b.height {
			b.pressed = true
			pressed = true
		}
	} else if b.pressed {
		tempX, tempY := ebiten.CursorPosition()
		mouseX, mouseY := float64(tempX), float64(tempY)
		if mouseX>=b.x && mouseX<=b.x+b.width && mouseY>=b.y && mouseY<=b.y+b.height {
			(*b).onClick(g)
		}
		b.pressed = false
		pressed = true
	}
	return pressed
}

// Required from Ebiten.
// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	switch g.state {
	case Preparing:
		g.buttons = append(g.buttons, &Button{775.0, 450.0, 350.0, 110.0,
			[]string{
				"====== ========     ||      =====   ==========",
				"||        ||       || ||    ||   ||     ||",
				"||        ||      ||   ||   ||    ||    ||",
				"======    ||     =========  ======      ||",
				"     ||   ||    ||       || ||   ||     ||",
				"     ||   ||    ||       || ||    ||    ||",
				"======    ||    ||       || ||     ||   ||",
			},
			func (g *Game) {
				g.state = Drawing},
			false,
		})
		g.buttons = append(g.buttons, &Button{1570.0, 10.0, 330.0, 110.0,
			[]string{
				"======  ||     ======     ||      =====   ",
				"||      ||     ||        || ||    ||   || ",
				"||      ||     ||       ||   ||   ||    ||",
				"||      ||     ======  =========  ======  ",
				"||      ||     ||     ||       || ||   || ",
				"||      ||     ||     ||       || ||    ||",
				"======  ====== ====== ||       || ||     |",
			},
			func (g *Game) {
				g.points = make([]Point, 0)
			},
			false,
		})
		g.buttons = append(g.buttons, &Button{1485.0, 950.0, 415.0, 110.0,
			[]string{
				"======   =======  ||     || =====    || ====== =====   ",
				"||       |     |  ||     || ||   ||  || ||     ||   || ",
				"||      ||     || ||     || ||    || || ||     ||    ||",
				"======  ||     || ||     || ======   || ====== ======  ",
				"||      ||     || ||     || ||   ||  || ||     ||   || ",
				"||       |     |  ||     || ||    || || ||     ||    ||",
				"||       =======   =======  ||     | || ====== ||     |",
			},
			func (g *Game) {
				g.state = Revealing
				g.revealIndex = 0
				g.fourierIndex = 0
			},
			false,
		})
		g.state = Start
	case Start:
		g.buttons[StartButton].CheckIfClicked(g)
	case Drawing:
		buttonPressed := g.buttons[ClearButton].CheckIfClicked(g)
		buttonPressed = buttonPressed || g.buttons[FourierButton].CheckIfClicked(g)

		if !buttonPressed && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			dim := len(g.points)
			if (dim==0 || float64(x)!=g.points[dim-1].x || float64(y)!=g.points[dim-1].y) {
				g.points = append(g.points, Point{float64(x), float64(y)})
			}
		}
	case Revealing:
		if  g.revealIndex<len(g.points) {
			g.revealIndex++
		} else {
			g.state = Computing
		}
	case Computing:
		pointsLen := len(g.points)
		sequenceX := make([]float64, pointsLen)
		sequenceY := make([]float64, pointsLen)
		for i:=0; i<pointsLen; i++ {
			sequenceX[i] = g.points[i].x
			sequenceY[i] = g.points[i].y
		}
		shiftSequence(sequenceX, float64(-g.windowSize.width)/2)
		shiftSequence(sequenceY, float64(-g.windowSize.height)/2)
		g.fourierX = fourier.DiscreteFourierTransform(sequenceX)
		g.fourierY = fourier.DiscreteFourierTransform(sequenceY)

		g.fourierIndex = 0
		g.fourierPoints = make([]Point, 0)
		g.state = Fourier
	case Fourier:
		if g.fourierIndex<len(g.fourierX)-1  {
				g.fourierIndex++
		} else {
			g.state = Drawing
		}
	}

	return nil
}

// Required from Ebiten.
// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	lineColor := color.White

	switch g.state {
	case Start:
		drawButton(screen, g.buttons[StartButton])
	case Drawing:
		for i:=1; i<len(g.points)-1; i++ {
			ebitenutil.DrawLine(screen, g.points[i-1].x, g.points[i-1].y, g.points[i].x, g.points[i].y, lineColor)
		}
		drawButton(screen, g.buttons[ClearButton])
		drawButton(screen, g.buttons[FourierButton])
	case Revealing:
		for i:=1; i<g.revealIndex; i++ {
			ebitenutil.DrawLine(screen, g.points[i-1].x, g.points[i-1].y, g.points[i].x, g.points[i].y, lineColor)
		}
	case Fourier:
		x1, y1:= drawFourierEpicycles(screen, g.fourierX, g.fourierIndex, float64(g.windowSize.width)/2 , 200, 0.0)
		x2, y2 := drawFourierEpicycles(screen, g.fourierY, g.fourierIndex, 200, float64(g.windowSize.height)/2, -math.Pi/2)
		g.fourierPoints = append(g.fourierPoints, Point{x1,y2})

		vector.DrawFilledCircle(screen, float32(x1), float32(y1), float32(6.0), color.RGBA{255, 0, 0, 100}, false)
		vector.DrawFilledCircle(screen, float32(x2), float32(y2), float32(6.0), color.RGBA{0, 255, 0, 100}, false)
		ebitenutil.DrawLine(screen, x1, y1, x1, float64(g.windowSize.height), lineColor)
		ebitenutil.DrawLine(screen, x2, y2, float64(g.windowSize.width), y2, lineColor)

		for i:=1; i<len(g.fourierPoints); i++ {
			ebitenutil.DrawLine(screen, g.fourierPoints[i-1].x, g.fourierPoints[i-1].y, g.fourierPoints[i].x, g.fourierPoints[i].y, lineColor)
		}
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
	game.state = Preparing
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