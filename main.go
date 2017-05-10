package main

import (
	"bytes"
	"fmt"
	sf "github.com/manyminds/gosfml"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"time"
)

const (
	height = 17
	width  = 150
)

var (
	reset   = false
	offsetX = 0
	offsetY = 0

	levels   [][][]byte
	tileMap  [][]byte
	curLevel = 0
	mario    *Player
	enemies  []*Enemy
	tileSet  *sf.Texture
)

type Player struct {
	*sf.Sprite

	dx, dy       float32
	rect         sf.FloatRect
	onGround     bool
	currentFrame float32
}

func NewPlayer(image *sf.Texture) *Player {
	p := new(Player)
	p.Sprite, _ = sf.NewSprite(image)
	p.rect = sf.FloatRect{100, 180, 16, 16}
	p.dx = 0.1
	p.dy = 0.1
	return p
}

func (p *Player) Update(time float32) {
	p.rect.Left += p.dx * time
	p.Collision(0)
	if reset {
		return
	}

	if !p.onGround {
		p.dy = p.dy + 0.0005*time
	}
	p.rect.Top += p.dy * time
	p.onGround = false
	p.Collision(1)

	p.currentFrame += time * 0.005
	if p.currentFrame > 3 {
		p.currentFrame -= 3
	}

	if p.dx > 0 {
		p.SetTextureRect(sf.IntRect{112 + 31*int(p.currentFrame), 144, 16, 16})
	}
	if p.dx < 0 {
		p.SetTextureRect(sf.IntRect{112 + 31*int(p.currentFrame) + 16, 144, -16, 16})
	}

	p.SetPosition(sf.Vector2f{p.rect.Left - float32(offsetX), p.rect.Top - float32(offsetY)})
	p.dx = 0
}

func (p *Player) Collision(num int) {
	for i := int(p.rect.Top / 16); float32(i) < ((p.rect.Top + p.rect.Height) / 16); i++ {
		for j := int(p.rect.Left / 16); float32(j) < ((p.rect.Left + p.rect.Width) / 16); j++ {
			if (tileMap[i][j] == 'P') || (tileMap[i][j] == 'k') || (tileMap[i][j] == '0') || (tileMap[i][j] == 'r') || (tileMap[i][j] == 't') {
				if p.dy > 0 && num == 1 {
					p.rect.Top = float32(i*16) - p.rect.Height
					p.dy = 0
					p.onGround = true
				}
				if p.dy < 0 && num == 1 {
					p.rect.Top = float32(i*16) + 16
					p.dy = 0
				}
				if p.dx > 0 && num == 0 {
					p.rect.Left = float32(j*16) - p.rect.Width
				}
				if p.dx < 0 && num == 0 {
					p.rect.Left = float32(j*16) + 16
				}
			}

			// if tileMap[i][j] == 'c' {
			// 	tileMap[i][j] = ' '
			// }

			if tileMap[i][j] == 'x' {
				if curLevel < len(levels)-1 {
					time.Sleep(time.Second * 1)
					curLevel++
					LoadLevel(curLevel)
					return
				}
			}
		}
	}
}

type Enemy struct {
	*sf.Sprite

	dx, dy       float32
	rect         sf.FloatRect
	currentFrame float32
	life         bool
}

func NewEnemy(image *sf.Texture, x, y int) *Enemy {
	e := new(Enemy)
	e.Sprite, _ = sf.NewSprite(image)
	e.rect = sf.FloatRect{float32(x), float32(y), 16, 16}
	e.dx = 0.05
	e.currentFrame = 0
	e.life = true
	return e
}

func (e *Enemy) Update(time float32) {
	e.rect.Left += e.dx * time
	e.Collision()

	e.currentFrame += time * 0.005
	if e.currentFrame > 2 {
		e.currentFrame -= 2
	}

	e.SetTextureRect(sf.IntRect{18 * int(e.currentFrame), 0, 16, 16})
	if !e.life {
		e.SetTextureRect(sf.IntRect{58, 0, 16, 16})
	}
	e.SetPosition(sf.Vector2f{e.rect.Left - float32(offsetX), e.rect.Top - float32(offsetY)})
}

func (e *Enemy) Collision() {
	for i := int(e.rect.Top / 16); float32(i) < (e.rect.Top+e.rect.Height)/16; i++ {
		for j := int(e.rect.Left / 16); float32(j) < (e.rect.Left+e.rect.Width)/16; j++ {
			if (tileMap[i][j] == 'P') || (tileMap[i][j] == '0') || tileMap[i][j] == 'r' {
				if e.dx > 0 {
					e.rect.Left = float32(j*16) - e.rect.Width
					e.dx *= -1
				} else if e.dx < 0 {
					e.rect.Left = float32(j*16) + 16
					e.dx *= -1
				}
			}
		}
	}
}

func LoadLevel(n int) {
	fmt.Println(n, " ", len(levels))
	tileMap = levels[n]

	mario = NewPlayer(tileSet)
	fmt.Println(mario.GetPosition())
	enemies = []*Enemy{}
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if tileMap[i][j] == 'e' {
				enemies = append(enemies, NewEnemy(tileSet, j*16, i*16))
			}
		}
	}
}

func main() {
	runtime.LockOSThread()

	var err error
	tileSet, err = sf.NewTextureFromFile("res/tileset.png", &sf.IntRect{})
	if err != nil {
		panic(err)
	}

	files, _ := ioutil.ReadDir("levels")
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".txt" {
			filename := "levels/" + f.Name()
			text, err := ioutil.ReadFile(filename)

			if err != nil {
				panic(err)
			}
			levels = append(levels, bytes.Split(text, []byte("\n")))
		}
	}

	LoadLevel(curLevel)

	window := sf.NewRenderWindow(sf.VideoMode{400, 250, 32}, "mario", sf.StyleDefault, sf.DefaultContextSettings())
	window.SetVSyncEnabled(true)

	tile, err := sf.NewSprite(tileSet)
	if err != nil {
		panic(err)
	}

	buffer, err := sf.NewSoundBufferFromFile("res/jump.ogg")
	if err != nil {
		panic(err)
	}
	sound := sf.NewSound(buffer)

	music, err := sf.NewMusicFromFile("res/theme.ogg")
	if err != nil {
		panic(err)
	}
	music.SetLoop(true)
	music.Play()

	for window.IsOpen() {
		for event := window.PollEvent(); event != nil; event = window.PollEvent() {
			switch event.(type) {
			case sf.EventClosed:
				window.Close()
			}
		}

		if sf.KeyboardIsKeyPressed(sf.KeyLeft) {
			mario.dx = -0.1
		}
		if sf.KeyboardIsKeyPressed(sf.KeyRight) {
			mario.dx = 0.1
		}
		if sf.KeyboardIsKeyPressed(sf.KeyUp) {
			if mario.onGround {
				mario.dy = -0.27
				mario.onGround = false
				sound.Play()
			}
		}

		music.GetStatus()

		mario.Update(30)
		for _, enemy := range enemies {
			enemy.Update(30)

			if intersects, _ := mario.rect.Intersects(enemy.rect); intersects {
				if enemy.life {
					if mario.dy > 0 {
						enemy.dx = 0
						mario.dy = -0.2
						enemy.life = false
					} else {
						mario.SetColor(sf.ColorRed())
					}
				}
			}
		}

		if mario.rect.Left > 200 {
			offsetX = int(mario.rect.Left - 200)
		}

		window.Clear(sf.Color{107, 140, 255, 255})

		for i := 0; i < height; i++ {
			for j := 0; j < width; j++ {
				tileValue := tileMap[i][j]
				if tileValue == 'P' {
					tile.SetTextureRect(sf.IntRect{143 - 16*3, 112, 16, 16})
				} else if tileValue == 'k' {
					tile.SetTextureRect(sf.IntRect{143, 112, 16, 16})
				} else if tileValue == 'c' {
					tile.SetTextureRect(sf.IntRect{143 - 16, 112, 16, 16})
				} else if tileValue == 't' {
					tile.SetTextureRect(sf.IntRect{0, 47, 32, 95 - 47})
				} else if tileValue == 'g' {
					tile.SetTextureRect(sf.IntRect{0, 16*9 - 5, 3 * 16, 16*2 + 5})
				} else if tileValue == 'G' {
					tile.SetTextureRect(sf.IntRect{145, 222, 222 - 145, 255 - 222})
				} else if tileValue == 'd' {
					tile.SetTextureRect(sf.IntRect{0, 106, 74, 127 - 106})
				} else if tileValue == 'w' {
					tile.SetTextureRect(sf.IntRect{99, 224, 140 - 99, 255 - 224})
				} else if tileValue == 'C' {
					tile.SetTextureRect(sf.IntRect{96, 0, 106, 112})
				} else if tileValue == 'r' {
					tile.SetTextureRect(sf.IntRect{143 - 32, 112, 16, 16})
				} else if (tileValue == ' ') || (tileValue == '0') || (tileValue == 'e') || (tileValue == 'x') {
					continue
				}
				tile.SetPosition(sf.Vector2f{float32(j*16 - offsetX), float32(i*16 - offsetY)})
				window.Draw(tile, sf.DefaultRenderStates())
			}
		}

		window.Draw(mario, sf.DefaultRenderStates())
		for _, enemy := range enemies {
			window.Draw(enemy, sf.DefaultRenderStates())
		}

		window.Display()
	}
}
