package core

import (
	"fmt"
	"time"

	"echo-taiga/internal/ai/fear"
	"echo-taiga/internal/audio"
	"echo-taiga/internal/config"
	"echo-taiga/internal/engine"
	"echo-taiga/internal/engine/ecs"
	"echo-taiga/internal/entities/player"
	"echo-taiga/internal/metamorphosis"
	"echo-taiga/internal/render"
	"echo-taiga/internal/symbols"
	"echo-taiga/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game представляет полную игровую структуру
type Game struct {
	config    *config.Config
	ecsWorld  *ecs.World
	engine    *engine.Engine
	world     *world.World
	player    *player.Player
	renderer  *render.Renderer
	audioMgr  *audio.Manager
	fearMgr   *fear.Director
	symbolMgr *symbols.Manager
	metamorph *metamorphosis.MetamorphosisManager

	isRunning      bool
	lastUpdateTime time.Time
}

// NewGame создает новый экземпляр игры
func NewGame(cfg *config.Config) (*Game, error) {
	// Инициализируем ECS мир
	ecsWorld := ecs.NewWorld()

	// Создаем мир с определенным сидом
	seed := cfg.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	gameWorld := world.NewWorld(seed, ecsWorld)

	// Создаем игрока
	playerEntity, err := player.CreatePlayerEntity(ecsWorld, gameWorld)
	if err != nil {
		return nil, fmt.Errorf("failed to create player: %v", err)
	}

	// Создаем менеджер символов
	symbolMgr := symbols.NewSymbolManager()
	err = symbolMgr.Initialize(worldSeed)
	if err := symbolMgr.Initialize(worldSeed); err != nil {
		return nil, err
	}

	// Создаем менеджер страха
	fearMgr := fear.NewFearDirector(ecsWorld, "saves/fear")
	err = fearMgr.Initialize()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize fear manager: %v", err)
	}

	// Создаем аудио менеджер
	audioMgr := audio.NewManager()

	// Создаем рендерер
	renderer, err := render.NewRenderer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %v", err)
	}

	// Создаем движок
	gameEngine := engine.NewEngine(ecsWorld)

	game := &Game{
		config:         cfg,
		ecsWorld:       ecsWorld,
		world:          gameWorld,
		player:         playerEntity,
		renderer:       renderer,
		audioMgr:       audioMgr,
		fearMgr:        fearMgr,
		symbolMgr:      symbolMgr,
		metamorph:      gameWorld.MetamorphManager,
		engine:         gameEngine,
		isRunning:      false,
		lastUpdateTime: time.Now(),
	}

	return game, nil
}

// Update обновляет состояние игры
func (g *Game) Update() error {
	// Вычисляем время между кадрами
	now := time.Now()
	deltaTime := now.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = now

	// Обновляем мир
	g.world.Update(deltaTime)

	// Обновляем движок ECS
	g.ecsWorld.Update(deltaTime)

	// Обновляем менеджер символов
	g.symbolMgr.Update(deltaTime)

	// Обновляем систему метаморфоз
	g.metamorph.Update(deltaTime)

	// Обновляем менеджер страха
	g.fearMgr.Update(deltaTime)

	return nil
}

// Draw отрисовывает игровой мир
func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Render(screen, g.world, g.player)
}

// Layout определяет размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.config.WindowWidth, g.config.WindowHeight
}

// Run запускает основной цикл игры
func (g *Game) Run() error {
	g.isRunning = true

	// Инициализируем Ebiten
	ebiten.SetWindowSize(g.config.WindowWidth, g.config.WindowHeight)
	ebiten.SetWindowTitle(g.config.Title)
	ebiten.SetTPS(g.config.TargetFPS)
	ebiten.SetVsyncEnabled(g.config.EnableVSync)

	// Устанавливаем начальную позицию игрока
	startPosition := ecs.Vector3{X: 0, Y: 0, Z: 0}
	g.world.SetPlayerPosition(startPosition)

	// Запускаем игру
	if err := ebiten.RunGame(g); err != nil {
		return fmt.Errorf("game loop error: %v", err)
	}

	return nil
}

// Stop останавливает игру
func (g *Game) Stop() {
	g.isRunning = false

	// Сохраняем состояние
	g.saveGameState()
}

// saveGameState сохраняет текущее состояние игры
func (g *Game) saveGameState() {
	// Сохраняем состояние мира
	// Сохраняем состояние метаморфоз
	g.metamorph.SaveState()

	// Сохраняем состояние символов
	g.symbolMgr.SaveState()

	// Сохраняем состояние игрока
	// TODO: Реализовать сохранение состояния игрока
}
