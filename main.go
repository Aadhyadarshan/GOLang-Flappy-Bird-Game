package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 640
	screenHeight = 630
	birdSize     = 40
	pillarWidth  = 200
	pillarGap    = 200
	pillarSpeed  = 3
)

var (
	birdImage   *ebiten.Image
	pillarImage *ebiten.Image
)

// Pillar represents a pillar object.
type Pillar struct {
	x, y int
}

// Game represents the game state.
type Game struct {
	mu         sync.Mutex
	birdY      float64
	birdDY     float64
	pillars    []*Pillar
	frameCount int
	isGameOver bool
	score      int
	started    bool
}

// Update updates the game state.
func (g *Game) Update() error {
	if !g.isGameOver && !g.started {
		if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			g.started = true
		}
	}

	if g.started && !g.isGameOver {
		// Apply gravity to bird's vertical velocity
		g.birdDY += 0.1

		// Update bird position
		if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			g.birdDY = -2
		}
		g.birdY += g.birdDY

		// Generate new pillars
		g.frameCount++
		if g.frameCount%(screenWidth/pillarSpeed) == 0 {
			g.pillars = append(g.pillars, NewPillar())
		}

		// Update pillar positions
		for _, pillar := range g.pillars {
			pillar.x -= pillarSpeed
			if pillar.x < -pillarWidth {
				g.pillars = g.pillars[1:]
				g.score++
			}

			// Check collision
			if g.birdY < float64(pillar.y) || g.birdY > float64(pillar.y+pillarGap) {
				if pillar.x < birdSize && pillar.x > -pillarWidth {
					g.isGameOver = true
				}
			}
		}
	}

	if g.isGameOver && ebiten.IsKeyPressed(ebiten.KeyEnter) {
		g.Reset()
	}

	return nil
}

// Draw draws the game.
func (g *Game) Draw(screen *ebiten.Image) {
	g.mu.Lock()
	defer g.mu.Unlock()

	screen.Fill(color.RGBA{75, 145, 201, 1})

	// Draw bird
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(50, g.birdY)
	screen.DrawImage(birdImage, op)

	// Draw pillars
	for _, pillar := range g.pillars {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(pillar.x), float64(pillar.y))
		screen.DrawImage(pillarImage, op)

		// Draw second pillar
		op.GeoM.Translate(0, float64(pillarGap)+pillarWidth)
		screen.DrawImage(pillarImage, op)
	}

	// Display score
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Score: %d", g.score))

	if !g.started {
		// Start prompt
		ebitenutil.DebugPrint(screen, "Press Space or Up Arrow to Start")
	}

	if g.isGameOver {
		// Game over screen
		ebitenutil.DebugPrint(screen, "GAME OVER. Press Enter to Restart")
	}
}

// Layout determines the layout of the game.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// NewPillar creates a new pillar object with random height.
func NewPillar() *Pillar {
	x := screenWidth
	y := rand.Intn(screenHeight - pillarGap - pillarWidth)
	if y < 0 {
		y = 0
	}
	return &Pillar{x: x, y: y}
}

// Reset resets the game state.
func (g *Game) Reset() {
	g.birdY = screenHeight / 2
	g.birdDY = 0
	g.pillars = []*Pillar{}
	g.frameCount = 0
	g.isGameOver = false
	g.score = 0
	g.started = false
}

func main() {
	rand.Seed(time.Now().UnixNano())

	game := &Game{
		pillars: []*Pillar{},
	}

	// Load bird image
	birdImg, _, err := ebitenutil.NewImageFromFile("flappybird/bird1.png")
	if err != nil {
		log.Fatal(err)
	}
	birdImage = birdImg

	// Load pillar image
	pillarImg, _, err := ebitenutil.NewImageFromFile("flappybird/pillar.png")
	if err != nil {
		log.Fatal(err)
	}
	pillarImage = pillarImg

	// Serve the frontend
	http.Handle("/", http.FileServer(http.Dir("public")))

	// Start the backend server
	go func() {
		fmt.Println("Backend server started on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Start the game loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
