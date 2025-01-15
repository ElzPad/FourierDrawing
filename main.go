package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/cmplx"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sqweek/dialog"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"fourier-drawing/fourier"
)

type GameState int
const (
    PREPARING GameState = iota
    START
    DRAWING
    REVEALING
    COMPUTING
    PRERENDERING
    FOURIER
    END
)

type Button struct {
    x, y, width, height float64
    text                []string
    onClick             func(g *Game)
    pressed             bool
}

type ButtonIndex int
const (
    START_BUTTON ButtonIndex = iota
    CLEAR_BUTTON
    SAVE_BUTTON
    LOAD_BUTTON
    FOURIER_BUTTON
)

type Point struct {
    x, y float64
}

/* PRERENDERING DISABLED
type Frame struct {
    drawingLayer *ebiten.Image
    dotsLayer *ebiten.Image
    epicyclesLayer *ebiten.Image
}
*/

// Required from Ebiten.
// Game implements ebiten.Game interface.
type Game struct {
    windowSize                  struct{ width, height int }
    // prerenderedFrames           []Frame        
    // frame                       *ebiten.Image        
    points                      []Point
    state                       GameState
    revealIndex                 int
    prerenderIndex              int
    toggleDots                  bool
    toggleEpicycles             bool
    fourierX                    []fourier.FourierElement
    fourierY                    []fourier.FourierElement
    fourierIndex                int
    fourierPoints               []Point
    buttons                     []*Button
}

const BUFFER_CIRCLES_OPTIONS = 10;
var BufferCircles [BUFFER_CIRCLES_OPTIONS]struct {
    circle      *ebiten.Image
    cx          float64
    cy          float64
    radius      float64
}

func writePointsToFile(points []Point) error {
    filePath, err := dialog.File().Filter("Text files (*.txt)", "txt").Load()
    if err != nil {
        return err
    }

    file, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    for _, point := range points {
        _, err := fmt.Fprintf(file, "%f, %f\n", point.x, point.y)
        if err != nil {
            return err
        }
    }

    return nil
}

func readPointsFromFile() ([]Point) {
    filePath, err := dialog.File().Filter("Text files (*.txt)", "txt").Load()
    if err != nil {
        return nil
    }

    file, err := os.Open(filePath)
    if err != nil {
        return nil
    }
    defer file.Close()

    var points []Point
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Split(line, ",")
        if len(parts) != 2 {
            continue
                }
        x, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
        if err != nil {
            return nil
        }
        y, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
        if err != nil {
            return nil
        }
        points = append(points, Point{x, y})
    }
    if err := scanner.Err(); err != nil {
        return nil
    }
    return points
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

func drawAndInitBufferCircles() {
	for i:=0; i<BUFFER_CIRCLES_OPTIONS; i++ {
		BufferCircles[i].cx = 50.0*float64(i+1)
    	BufferCircles[i].cy = 50.0*float64(i+1)
    	BufferCircles[i].radius = 45.0*float64(i+1)
		BufferCircles[i].circle = ebiten.NewImage(int(BufferCircles[i].cx)*2, int(BufferCircles[i].cy)*2)
		steps := 100
		dAngle := 2*math.Pi/float64(steps)
		point1 := Point{BufferCircles[i].cx+BufferCircles[i].radius, BufferCircles[i].cy}
		point2 := Point{0, 0}
		for j:=1; j<=steps; j++ {
			point2.x = BufferCircles[i].cx+BufferCircles[i].radius*math.Cos(dAngle*float64(j))
			point2.y = BufferCircles[i].cy+BufferCircles[i].radius*math.Sin(dAngle*float64(j))
			ebitenutil.DrawLine(BufferCircles[i].circle, point1.x, point1.y, point2.x, point2.y, color.RGBA{150, 150, 150, 255})
			point1 = point2
		}
	}
}

func drawEmptyCircle(screen *ebiten.Image, cx, cy, radius float64) {
    index := 0
	for i:=0; i<BUFFER_CIRCLES_OPTIONS; i++ {
		if math.Abs(radius-BufferCircles[i].radius) < math.Abs(radius-BufferCircles[index].radius) {
			index = i
		}
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(radius/BufferCircles[index].radius, radius/BufferCircles[index].radius)
	opts.GeoM.Translate(cx-BufferCircles[index].cx*radius/BufferCircles[index].radius, cy-BufferCircles[index].cy*radius/BufferCircles[index].radius)
    
	screen.DrawImage(BufferCircles[index].circle, opts)
}

func drawEmptyCircleWithRadius(screen1 *ebiten.Image, screen2 *ebiten.Image, cx, cy, radius, angle float64, lineColor color.Color, drawCircle bool) (x, y float64) {
    if (drawCircle) {
		drawEmptyCircle(screen2, cx, cy, radius)
	}
    x = cx+radius*math.Cos(angle)
    y = cy-radius*math.Sin(angle)
    ebitenutil.DrawLine(screen1, cx, cy, x, y, lineColor)

    return x,y
}

func drawFourierEpicycles(screen1 *ebiten.Image, screen2 *ebiten.Image, fourierSeq []fourier.FourierElement, fourierInd int, startX, startY, phase float64, drawCircles bool) (x, y float64) {
    N := len(fourierSeq)
    x, y = startX, startY

    for k:=0; k<N; k++ {
        radius := cmplx.Abs(fourierSeq[k].Val)/float64(N)
        arg := 2 * math.Pi * float64(fourierInd) * float64(fourierSeq[k].Freq) / float64(N) + cmplx.Phase(fourierSeq[k].Val) + phase;

        x, y = drawEmptyCircleWithRadius(screen1, screen2, x, y, radius, arg, color.RGBA{150, 150, 150, 255}, drawCircles)
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

/* PRERENDERING DISABLED
func preRenderFrame(g *Game, frameIndex int) Frame {
    color1 := color.RGBA{64, 64, 64, 64}
    color3 := color.RGBA{255, 255, 255, 255}

    circleWidthBold := 4.0

    drawingImage := ebiten.NewImage(g.windowSize.width, g.windowSize.height)
    dotsImage := ebiten.NewImage(g.windowSize.width, g.windowSize.height)
    epicyclesImage := ebiten.NewImage(g.windowSize.width, g.windowSize.height)
    
    x1, y1 := drawFourierEpicycles(drawingImage, epicyclesImage, g.fourierX, frameIndex, float64(g.windowSize.width)/2 , 100, 0.0, g.toggleEpicycles)
    x2, y2 := drawFourierEpicycles(drawingImage, epicyclesImage, g.fourierY, frameIndex, 200, float64(g.windowSize.height)/2, -math.Pi/2, g.toggleEpicycles)

    vector.DrawFilledCircle(drawingImage, float32(x1), float32(y1), float32(6.0), color.RGBA{255, 0, 0, 100}, false)
    vector.DrawFilledCircle(drawingImage, float32(x2), float32(y2), float32(6.0), color.RGBA{0, 255, 0, 100}, false)
    
    if (y2 >= 200) {
        ebitenutil.DrawLine(drawingImage, x1, y1, x1, float64(g.windowSize.height), color.White)
    } else {
        ebitenutil.DrawLine(drawingImage, x1, 0, x1, y1, color.White)
    }
    if (x1 >= 200) {
        ebitenutil.DrawLine(drawingImage, x2, y2, float64(g.windowSize.width), y2, color.White)
    } else {
        ebitenutil.DrawLine(drawingImage, 0, y2, x2, y2, color.White)
    }

    for i:=1; i<frameIndex; i++ {
        ebitenutil.DrawLine(drawingImage, g.fourierPoints[i-1].x, g.fourierPoints[i-1].y, g.fourierPoints[i].x, g.fourierPoints[i].y, color1)
        ebitenutil.DrawCircle(dotsImage, g.fourierPoints[i].x, g.fourierPoints[i].y, circleWidthBold, color3)
    }

    return Frame{ drawingLayer: drawingImage, dotsLayer: dotsImage, epicyclesLayer: epicyclesImage }
}

func preRenderBatchOfFrames(g *Game) {
    for i:=0; i<3 && g.prerenderIndex<len(g.fourierX); i++ {
        g.prerenderedFrames[g.prerenderIndex] = preRenderFrame(g, g.prerenderIndex)
        g.prerenderIndex++
    }    
}
*/

// Required from Ebiten.
// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
    if (ebiten.IsKeyPressed(ebiten.KeyC)) {
        g.toggleDots = true
    } else if (ebiten.IsKeyPressed(ebiten.KeyV)) {
        g.toggleDots = false
    } else if (ebiten.IsKeyPressed(ebiten.KeyD)) {
        g.toggleEpicycles = true
    } else if (ebiten.IsKeyPressed(ebiten.KeyF)) {
        g.toggleEpicycles = false
    }

    switch g.state {
    case PREPARING:
        drawAndInitBufferCircles()
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
                g.state = DRAWING},
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
                // g.prerenderedFrames = make([]Frame, 0)
                g.prerenderIndex = 0
            },
            false,
        })
        g.buttons = append(g.buttons, &Button{10.0, 815.0, 320.0, 110.0,
            []string{
                "======      ||     ||          || ======",
                "||         || ||    ||        ||  ||    ",
                "||        ||   ||    ||      ||   ||    ",
                "======   =========    ||    ||    ======",
                "     || ||       ||    ||  ||     ||    ",
                "     || ||       ||     ||||      ||    ",
                "======  ||       ||      ||       ======",
            },
            func (g *Game) {
                err := writePointsToFile(g.points)
                if (err != nil) {
                    fmt.Printf("Unable to write points to file.\n")
                }
            },
            false,
        })
        g.buttons = append(g.buttons, &Button{10.0, 950.0, 295.0, 110.0,
            []string{
                "||      =======      ||      =====    ",
                "||      |     |     || ||    ||   ||  ",
                "||     ||     ||   ||   ||   ||    || ",
                "||     ||     ||  =========  ||    || ",
                "||     ||     || ||       || ||    || ",
                "||      |     |  ||       || ||   ||  ",
                "======  =======  ||       || =====    ",
            },
            func (g *Game) {
                g.points = readPointsFromFile()
                // g.prerenderedFrames = make([]Frame, 0)
                g.prerenderIndex = 0
                if (g.points == nil) {
                    fmt.Printf("Unable to read points from file.\n")
                }
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
                if (len(g.points)>0) {
                    g.state = REVEALING
                    g.revealIndex = 0
                    g.fourierIndex = 0
                }
            },
            false,
        })
        // g.prerenderedFrames = make([]Frame, 0)
        g.state = START
    case START:
        g.buttons[START_BUTTON].CheckIfClicked(g)
    case DRAWING:
        buttonPressed := g.buttons[CLEAR_BUTTON].CheckIfClicked(g)
        buttonPressed = buttonPressed || g.buttons[SAVE_BUTTON].CheckIfClicked(g)
        buttonPressed = buttonPressed || g.buttons[LOAD_BUTTON].CheckIfClicked(g)
        buttonPressed = buttonPressed || g.buttons[FOURIER_BUTTON].CheckIfClicked(g)

        if !buttonPressed && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
            x, y := ebiten.CursorPosition()
            dim := len(g.points)
            if (dim==0 || float64(x)!=g.points[dim-1].x || float64(y)!=g.points[dim-1].y) {
                g.points = append(g.points, Point{float64(x), float64(y)})
            }
        }
    case REVEALING:
        if g.revealIndex<len(g.points) && !ebiten.IsKeyPressed(ebiten.KeyS){
            g.revealIndex++
        } else {
            g.state = COMPUTING
        }
    case COMPUTING:
        pointsLen := len(g.points)
        sequenceX := make([]float64, pointsLen)
        sequenceY := make([]float64, pointsLen)
        for i:=0; i<pointsLen; i++ {
            sequenceX[i] = g.points[i].x
            sequenceY[i] = g.points[i].y
        }
        shiftSequence(sequenceX, float64(-g.windowSize.width)/2)
        shiftSequence(sequenceY, float64(-g.windowSize.height)/2)
        g.fourierX = fourier.DiscreteFourierTransform(sequenceX, true)
        g.fourierY = fourier.DiscreteFourierTransform(sequenceY, true)

        sequenceX = fourier.InverseDFT(g.fourierX)
        sequenceY = fourier.InverseDFT(g.fourierY)

        g.fourierPoints = make([]Point, len(sequenceX))
        for i:=0; i<pointsLen; i++ {
            g.fourierPoints[i].x = sequenceX[i]+float64(g.windowSize.width)/2
            g.fourierPoints[i].y = sequenceY[i]+float64(g.windowSize.height)/2
        }

        g.fourierIndex = 0

        g.state = FOURIER

/* PRERENDERING DISABLED
        if g.prerenderIndex == len(g.fourierX) {
            g.state = FOURIER
        } else {
            expandedFramesSlice := make([]Frame, len(g.fourierX))
            copy(expandedFramesSlice, g.prerenderedFrames)
            g.prerenderedFrames = expandedFramesSlice
            g.state = PRERENDERING
        }
        
    case PRERENDERING:
        if g.prerenderIndex<len(g.fourierX)  {
            preRenderBatchOfFrames(g)
        } else {
            g.fourierIndex = 0
            g.state = FOURIER
        }
*/

    case FOURIER:
        if g.fourierIndex<len(g.fourierX)-1  {
            g.fourierIndex++
        } else {
            g.state = DRAWING
        }
    }

    return nil
}

// Required from Ebiten.
// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
    screen.Fill(color.Black)

    color1 := color.RGBA{64, 64, 64, 64}
    color2 := color.RGBA{192, 192, 192, 255}

    circleWidth := 3.0

    switch g.state {
    case START:
        drawButton(screen, g.buttons[START_BUTTON])
    case DRAWING:
        for i:=1; i<len(g.points); i++ {
            ebitenutil.DrawLine(screen, g.points[i-1].x, g.points[i-1].y, g.points[i].x, g.points[i].y, color1)
            if (g.toggleDots) {
                ebitenutil.DrawCircle(screen, g.points[i].x, g.points[i].y, circleWidth, color2)
            }
        }
        drawButton(screen, g.buttons[CLEAR_BUTTON])
        drawButton(screen, g.buttons[SAVE_BUTTON])
        drawButton(screen, g.buttons[LOAD_BUTTON])
        drawButton(screen, g.buttons[FOURIER_BUTTON])
    case REVEALING:
        text.Draw(screen, "Click S to skip", basicfont.Face7x13, 940, 20, color.White)
        for i:=1; i<g.revealIndex; i++ {
            ebitenutil.DrawLine(screen, g.points[i-1].x, g.points[i-1].y, g.points[i].x, g.points[i].y, color1)
            if (g.toggleDots) {
                ebitenutil.DrawCircle(screen, g.points[i].x, g.points[i].y, circleWidth, color2)
            }
        }
    case PRERENDERING:
        textOnScreen := fmt.Sprintf("Prerendering: %.2f%%", float64(g.prerenderIndex)/float64(len(g.fourierX))*100)
        text.Draw(screen, textOnScreen, basicfont.Face7x13, 900, 530, color.White)
        ebitenutil.DrawRect(screen, 760, 560, float64(g.prerenderIndex)/float64(len(g.fourierX))*400, 40, color.White)
    case FOURIER:
        color3 := color.RGBA{255, 255, 255, 255}
        circleWidthBold := 4.0
        x1, y1 := drawFourierEpicycles(screen, screen, g.fourierX, g.fourierIndex, float64(g.windowSize.width)/2 , 100, 0.0, g.toggleEpicycles)
        x2, y2 := drawFourierEpicycles(screen, screen, g.fourierY, g.fourierIndex, 200, float64(g.windowSize.height)/2, -math.Pi/2, g.toggleEpicycles)

        vector.DrawFilledCircle(screen, float32(x1), float32(y1), float32(6.0), color.RGBA{255, 0, 0, 100}, false)
        vector.DrawFilledCircle(screen, float32(x2), float32(y2), float32(6.0), color.RGBA{0, 255, 0, 100}, false)
        
        if (y2 >= 200) {
            ebitenutil.DrawLine(screen, x1, y1, x1, float64(g.windowSize.height), color.White)
        } else {
            ebitenutil.DrawLine(screen, x1, 0, x1, y1, color.White)
        }
        if (x1 >= 200) {
            ebitenutil.DrawLine(screen, x2, y2, float64(g.windowSize.width), y2, color.White)
        } else {
            ebitenutil.DrawLine(screen, 0, y2, x2, y2, color.White)
        }

        for i:=1; i<g.fourierIndex; i++ {
            ebitenutil.DrawLine(screen, g.fourierPoints[i-1].x, g.fourierPoints[i-1].y, g.fourierPoints[i].x, g.fourierPoints[i].y, color1)
            if (g.toggleDots) {
				ebitenutil.DrawCircle(screen, g.fourierPoints[i].x, g.fourierPoints[i].y, circleWidthBold, color3)
			}
		}
	}
    if (g.state!=PREPARING && g.state!=START) {
        if (g.toggleDots) {
            text.Draw(screen, "Points visualization: enabled      - Click V to disable", basicfont.Face7x13, 20, 20, color.White)
        } else {
            text.Draw(screen, "Points visualization: disabled     - Click C to enable", basicfont.Face7x13, 20, 20, color.White)
        }

        if (g.toggleEpicycles) {
            text.Draw(screen, "Epicycles visualization: enabled   - Click F to disable", basicfont.Face7x13, 20, 40, color.White)
        } else {
            text.Draw(screen, "Epicycles visualization: disabled  - Click D to enable", basicfont.Face7x13, 20, 40, color.White)
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
    game.state = PREPARING
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