package core

import (
	"echo-taiga/internal/ai/director"
	"echo-taiga/internal/audio"
	"echo-taiga/internal/config"
	"echo-taiga/internal/engine"
	"echo-taiga/internal/entities/player"
	"echo-taiga/internal/render"
	"echo-taiga/internal/world"
)

// Game представляет основной класс игры
type Game struct {
	config     *config.Config
	engine     *engine.Engine
	world      *world.World
	player     *player.Player
	renderer   *render.Renderer
	audioMgr   *audio.Manager
	aiDirector *director.AIDirector
	isRunning  bool
}

// NewGame создает новый экземпляр игры
func NewGame(cfg *config.Config) (*Game, error) {
	// TODO: инициализация компонентов игры
	return &Game{
		config:    cfg,
		isRunning: false,
	}, nil
}

// Run запускает основной цикл игры
func (g *Game) Run() error {
	g.isRunning = true

	// TODO: реализовать игровой цикл

	return nil
}

// Stop останавливает игру
func (g *Game) Stop() {
	g.isRunning = false
}
