package render

import (
	"fmt"
	"image/color"

	"echo-taiga/internal/config"
	"echo-taiga/internal/engine/ecs"
	"echo-taiga/internal/entities/player"
	"echo-taiga/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

// Renderer отвечает за визуализацию игрового мира
type Renderer struct {
	config         *config.Config
	pixelFont      *ebiten.Image
	backgroundTile *ebiten.Image
}

// NewRenderer создает новый рендерер
func NewRenderer(cfg *config.Config) (*Renderer, error) {
	// Создаем базовый рендерер
	renderer := &Renderer{
		config: cfg,
	}

	// Инициализируем ресурсы для рендеринга
	if err := renderer.initializeResources(); err != nil {
		return nil, err
	}

	return renderer, nil
}

// initializeResources инициализирует графические ресурсы
func (r *Renderer) initializeResources() error {
	// Создаем базовый тайл для фона
	backgroundImage := ebiten.NewImage(64, 64)
	backgroundImage.Fill(color.RGBA{R: 100, G: 150, B: 200, A: 255}) // Голубой цвет неба
	r.backgroundTile = backgroundImage

	// TODO: Загрузка шрифта или создание пиксельного шрифта
	// Временно используем встроенный шрифт
	r.pixelFont = ebiten.NewImage(8, 8)
	r.pixelFont.Fill(color.White)

	return nil
}

// Render основной метод отрисовки игрового мира
func (r *Renderer) Render(screen *ebiten.Image, gameWorld *world.World, player *player.Player) {
	// Очищаем экран
	screen.Fill(color.RGBA{R: 50, G: 50, B: 50, A: 255})

	// Рисуем фон
	r.renderBackground(screen, gameWorld)

	// Рисуем чанки
	r.renderChunks(screen, gameWorld)

	// Рисуем сущности
	r.renderEntities(screen, gameWorld)

	// Рисуем игрока
	r.renderPlayer(screen, player)

	// Отрисовка пользовательского интерфейса
	r.renderUI(screen, gameWorld)
}

// renderBackground отрисовывает базовый фон мира
func (r *Renderer) renderBackground(screen *ebiten.Image, gameWorld *world.World) {
	// Вычисляем размеры экрана
	screenWidth, screenHeight := screen.Size()

	// Создаем матрицу для фона с параллаксом
	op := &ebiten.DrawImageOptions{}

	// Смещение фона в зависимости от позиции игрока
	playerX := gameWorld.PlayerPosition.X
	playerZ := gameWorld.PlayerPosition.Z

	// Параллакс-эффект
	op.GeoM.Translate(-playerX/10, -playerZ/10)

	// Заполнение экрана тайлами фона
	for x := -1; x < screenWidth/64+2; x++ {
		for y := -1; y < screenHeight/64+2; y++ {
			op.GeoM.Reset()
			op.GeoM.Translate(float64(x*64), float64(y*64))
			op.GeoM.Translate(-playerX/10, -playerZ/10)
			screen.DrawImage(r.backgroundTile, op)
		}
	}
}

// renderChunks отрисовывает чанки мира
func (r *Renderer) renderChunks(screen *ebiten.Image, gameWorld *world.World) {
	// Определяем текущий чанк игрока
	playerChunkX := int(gameWorld.PlayerPosition.X / world.ChunkSize)
	playerChunkZ := int(gameWorld.PlayerPosition.Z / world.ChunkSize)

	// Радиус отрисовки чанков
	viewRadius := r.config.ViewDistance

	for x := playerChunkX - viewRadius; x <= playerChunkX+viewRadius; x++ {
		for z := playerChunkZ - viewRadius; z <= playerChunkZ+viewRadius; z++ {
			chunk := gameWorld.GetChunkAt(x, z)

			if chunk != nil && chunk.Terrain != nil {
				r.renderChunkTerrain(screen, chunk, gameWorld.PlayerPosition)
			}
		}
	}
}

// renderChunkTerrain отрисовывает ландшафт конкретного чанка
func (r *Renderer) renderChunkTerrain(screen *ebiten.Image, chunk *world.Chunk, playerPos ecs.Vector3) {
	// Создаем изображение для чанка
	chunkImage := ebiten.NewImage(int(world.ChunkSize), int(world.ChunkSize))

	// Цвета по типам поверхности
	groundColors := map[string]color.RGBA{
		"grass": {R: 100, G: 200, B: 100, A: 255},
		"rock":  {R: 150, G: 150, B: 150, A: 255},
		"water": {R: 50, G: 100, B: 250, A: 200},
		"mud":   {R: 139, G: 69, B: 19, A: 255},
		"snow":  {R: 255, G: 255, B: 255, A: 255},
		"void":  {R: 50, G: 50, B: 50, A: 255},
	}

	// Отрисовываем высоты и типы поверхности
	for x := 0; x < int(world.ChunkSize); x++ {
		for z := 0; z < int(world.ChunkSize); z++ {
			height := chunk.Terrain.GetHeight(x, z)
			groundType := chunk.Terrain.GetGroundType(x, z)

			// Выбираем цвет
			color := groundColors[groundType]

			// Модифицируем цвет в зависимости от высоты
			heightFactor := (height - 0) / 50.0 // Нормализация высоты
			color.R = uint8(float64(color.R) * (1 - heightFactor*0.3))
			color.G = uint8(float64(color.G) * (1 - heightFactor*0.3))
			color.B = uint8(float64(color.B) * (1 - heightFactor*0.3))

			// Рисуем пиксель
			chunkImage.Set(x, z, color)
		}
	}

	// Отрисовка чанка на экране
	op := &ebiten.DrawImageOptions{}
	chunkWorldX := float64(chunk.Position[0] * int(world.ChunkSize))
	chunkWorldZ := float64(chunk.Position[1] * int(world.ChunkSize))

	// Смещение относительно позиции игрока
	op.GeoM.Translate(
		chunkWorldX-playerPos.X+float64(r.config.WindowWidth)/2,
		chunkWorldZ-playerPos.Z+float64(r.config.WindowHeight)/2,
	)

	screen.DrawImage(chunkImage, op)
}

// renderEntities отрисовывает сущности в мире
func (r *Renderer) renderEntities(screen *ebiten.Image, gameWorld *world.World) {
	// Получаем все активные чанки
	for _, chunk := range gameWorld.ActiveChunks {
		// Отрисовываем сущности чанка
		for _, entityID := range chunk.Entities {
			entity, exists := gameWorld.ECSWorld.GetEntity(entityID)
			if !exists {
				continue
			}

			// Получаем компоненты трансформации
			transformComp, has := entity.GetComponent(ecs.TransformComponentID)
			if !has {
				continue
			}
			transform := transformComp.(*ecs.TransformComponent)

			// Получаем компоненты рендеринга
			renderComp, has := entity.GetComponent(ecs.RenderComponentID)
			if !has {
				continue
			}
			render := renderComp.(*ecs.RenderComponent)

			// Пропускаем невидимые объекты
			if !render.Visible {
				continue
			}

			// Простейшая отрисовка - цветной прямоугольник
			entityImage := ebiten.NewImage(10, 10)
			entityImage.Fill(render.Color)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(
				transform.Position.X-gameWorld.PlayerPosition.X+float64(r.config.WindowWidth)/2,
				transform.Position.Z-gameWorld.PlayerPosition.Z+float64(r.config.WindowHeight)/2,
			)

			screen.DrawImage(entityImage, op)
		}
	}
}

// renderPlayer отрисовывает игрока
func (r *Renderer) renderPlayer(screen *ebiten.Image, player *player.Player) {
	// Создаем изображение игрока
	playerImage := ebiten.NewImage(20, 20)
	playerImage.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255}) // Красный цвет

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		float64(r.config.WindowWidth)/2-10,
		float64(r.config.WindowHeight)/2-10,
	)

	screen.DrawImage(playerImage, op)
}

// renderUI отрисовывает пользовательский интерфейс
func (r *Renderer) renderUI(screen *ebiten.Image, gameWorld *world.World) {
	// Отрисовка времени суток
	timeText := fmt.Sprintf("Время дня: %.2f", gameWorld.GetGlobalTimeOfDay())

	// Отрисовка уровня аномальности
	anomalyText := fmt.Sprintf("Аномальность: %.2f", gameWorld.GetGlobalAnomalyLevel())

	// Временная отрисовка текста (позже заменить на нормальный шрифт)
	r.drawText(screen, timeText, 10, 30, color.White)
	r.drawText(screen, anomalyText, 10, 50, color.White)
}

// drawText - вспомогательный метод для отрисовки текста
func (r *Renderer) drawText(screen *ebiten.Image, message string, x, y int, c color.Color) {
	// TODO: Реализовать нормальную отрисовку текста с использованием шрифта
	for i, _ := range message {
		charImage := ebiten.NewImage(8, 8)
		charImage.Fill(c)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x+i*10), float64(y))
		screen.DrawImage(charImage, op)
	}
}
