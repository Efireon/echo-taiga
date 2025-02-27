package world

import (
	"github.com/yourusername/echo-taiga/internal/world/terrain"
	"github.com/yourusername/echo-taiga/internal/world/biomes"
	"github.com/yourusername/echo-taiga/internal/entities"
	"github.com/yourusername/echo-taiga/internal/metamorphosis"
)

// Chunk представляет одну часть игрового мира
type Chunk struct {
	Position      [2]int
	Terrain       *terrain.TerrainData
	Entities      []entities.Entity
	MetamorphEffects []metamorphosis.Effect
	IsGenerated   bool
	IsActive      bool
}

// World представляет весь игровой мир
type World struct {
	Seed             int64
	Chunks           map[[2]int]*Chunk
	ActiveChunks     map[[2]int]*Chunk
	BiomeMap         *biomes.BiomeMap
	TimeOfDay        float64 // 0.0 - 1.0, где 0.0 - полночь, 0.5 - полдень
	MetamorphManager *metamorphosis.Manager
	EntityManager    *entities.Manager
}

// NewWorld создает новый мир с указанным сидом
func NewWorld(seed int64) *World {
	return &World{
		Seed:         seed,
		Chunks:       make(map[[2]int]*Chunk),
		ActiveChunks: make(map[[2]int]*Chunk),
		TimeOfDay:    0.25, // Начинаем с рассвета
	}
}

// GetChunkAt возвращает чанк в указанной позиции
func (w *World) GetChunkAt(x, y int) *Chunk {
	pos := [2]int{x, y}
	chunk, exists := w.Chunks[pos]
	
	if !exists {
		// Генерируем новый чанк, если он не существует
		chunk = w.generateChunk(x, y)
		w.Chunks[pos] = chunk
	}
	
	return chunk
}

// ActivateChunk активирует чанк для обработки
func (w *World) ActivateChunk(x, y int) {
	chunk := w.GetChunkAt(x, y)
	pos := [2]int{x, y}
	chunk.IsActive = true
	w.ActiveChunks[pos] = chunk
}

// DeactivateChunk деактивирует чанк
func (w *World) DeactivateChunk(x, y int) {
	pos := [2]int{x, y}
	chunk, exists := w.Chunks[pos]
	
	if exists {
		chunk.IsActive = false
		delete(w.ActiveChunks, pos)
	}
}

// Генерирует новый чанк в указанной позиции
func (w *World) generateChunk(x, y int) *Chunk {
	// TODO: реализовать процедурную генерацию чанка
	return &Chunk{
		Position:    [2]int{x, y},
		IsGenerated: true,
		IsActive:    false,
	}
}

// Update обновляет состояние мира
func (w *World) Update(deltaTime float64) {
	// Обновление времени суток
	w.TimeOfDay += deltaTime * 0.001 // Примерный полный цикл за 1000 секунд
	if w.TimeOfDay >= 1.0 {
		w.TimeOfDay -= 1.0
	}
	
	// Обновление активных чанков
	for _, chunk := range w.ActiveChunks {
		w.updateChunk(chunk, deltaTime)
	}
	
	// Обновление менеджера метаморфоз
	w.MetamorphManager.Update(deltaTime)
	
	// Обновление менеджера сущностей
	w.EntityManager.Update(deltaTime)
}

// Обновляет состояние отдельного чанка
func (w *World) updateChunk(chunk *Chunk, deltaTime float64) {
	// TODO: реализовать обновление чанка
}
