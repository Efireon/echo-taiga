package biomes

import (
	"math"
	"math/rand"
)

// BiomeType и определения констант биомов обычно уже определены в вашем коде
// Если нет, то можно использовать код из предыдущего примера

// BiomeMap отвечает за хранение и управление биомами в игровом мире
type BiomeMap struct {
	biomeManager *BiomeManager
	chunkSize    int
	biomeChunks  map[ChunkCoord]*BiomeChunk
	seed         int64
	rng          *rand.Rand
}

// ChunkCoord представляет координаты чанка в мире
type ChunkCoord struct {
	X, Z int
}

// BiomeChunk содержит информацию о биомах в одном чанке мира
type BiomeChunk struct {
	Biomes        [][]BiomeType
	NoiseValues   map[string][][]float64
	MetamorpLevel [][]int
}

// NewBiomeMap создает новую карту биомов с указанным размером чанка и сидом
func NewBiomeMap(chunkSize int, seed int64) *BiomeMap {
	return &BiomeMap{
		biomeManager: NewBiomeManager(seed),
		chunkSize:    chunkSize,
		biomeChunks:  make(map[ChunkCoord]*BiomeChunk),
		seed:         seed,
		rng:          rand.New(rand.NewSource(seed)),
	}
}

// GetBiomeAt возвращает тип биома на указанных мировых координатах
func (bm *BiomeMap) GetBiomeAt(x, z float64) BiomeType {
	// Преобразуем мировые координаты в координаты чанка
	chunkX := int(math.Floor(x / float64(bm.chunkSize)))
	chunkZ := int(math.Floor(z / float64(bm.chunkSize)))

	// Координаты внутри чанка
	localX := int(x) % bm.chunkSize
	if localX < 0 {
		localX += bm.chunkSize
	}

	localZ := int(z) % bm.chunkSize
	if localZ < 0 {
		localZ += bm.chunkSize
	}

	// Получаем или генерируем чанк
	chunkCoord := ChunkCoord{X: chunkX, Z: chunkZ}
	chunk, exists := bm.biomeChunks[chunkCoord]

	if !exists {
		chunk = bm.generateBiomeChunk(chunkCoord)
		bm.biomeChunks[chunkCoord] = chunk
	}

	// Возвращаем биом на указанных координатах
	return chunk.Biomes[localX][localZ]
}

// GetNoiseValueAt возвращает значение шума определенного типа на указанных координатах
func (bm *BiomeMap) GetNoiseValueAt(noiseType string, x, z float64) float64 {
	// Преобразуем мировые координаты в координаты чанка
	chunkX := int(math.Floor(x / float64(bm.chunkSize)))
	chunkZ := int(math.Floor(z / float64(bm.chunkSize)))

	// Координаты внутри чанка
	localX := int(x) % bm.chunkSize
	if localX < 0 {
		localX += bm.chunkSize
	}

	localZ := int(z) % bm.chunkSize
	if localZ < 0 {
		localZ += bm.chunkSize
	}

	// Получаем или генерируем чанк
	chunkCoord := ChunkCoord{X: chunkX, Z: chunkZ}
	chunk, exists := bm.biomeChunks[chunkCoord]

	if !exists {
		chunk = bm.generateBiomeChunk(chunkCoord)
		bm.biomeChunks[chunkCoord] = chunk
	}

	// Возвращаем значение шума на указанных координатах
	if values, ok := chunk.NoiseValues[noiseType]; ok {
		return values[localX][localZ]
	}

	return 0.0
}

// GetMetamorphLevel возвращает уровень метаморфозы на указанных координатах
func (bm *BiomeMap) GetMetamorphLevel(x, z float64) int {
	// Преобразуем мировые координаты в координаты чанка
	chunkX := int(math.Floor(x / float64(bm.chunkSize)))
	chunkZ := int(math.Floor(z / float64(bm.chunkSize)))

	// Координаты внутри чанка
	localX := int(x) % bm.chunkSize
	if localX < 0 {
		localX += bm.chunkSize
	}

	localZ := int(z) % bm.chunkSize
	if localZ < 0 {
		localZ += bm.chunkSize
	}

	// Получаем или генерируем чанк
	chunkCoord := ChunkCoord{X: chunkX, Z: chunkZ}
	chunk, exists := bm.biomeChunks[chunkCoord]

	if !exists {
		chunk = bm.generateBiomeChunk(chunkCoord)
		bm.biomeChunks[chunkCoord] = chunk
	}

	// Возвращаем уровень метаморфозы на указанных координатах
	return chunk.MetamorpLevel[localX][localZ]
}

// ApplyMetamorphosis применяет метаморфозу указанного порядка в указанной области
func (bm *BiomeMap) ApplyMetamorphosis(centerX, centerZ float64, radius float64, order int) {
	// Определяем затронутые чанки
	minChunkX := int(math.Floor((centerX - radius) / float64(bm.chunkSize)))
	maxChunkX := int(math.Floor((centerX + radius) / float64(bm.chunkSize)))
	minChunkZ := int(math.Floor((centerZ - radius) / float64(bm.chunkSize)))
	maxChunkZ := int(math.Floor((centerZ + radius) / float64(bm.chunkSize)))

	// Для каждого затронутого чанка
	for chunkX := minChunkX; chunkX <= maxChunkX; chunkX++ {
		for chunkZ := minChunkZ; chunkZ <= maxChunkZ; chunkZ++ {
			chunkCoord := ChunkCoord{X: chunkX, Z: chunkZ}

			// Получаем или генерируем чанк
			chunk, exists := bm.biomeChunks[chunkCoord]
			if !exists {
				chunk = bm.generateBiomeChunk(chunkCoord)
				bm.biomeChunks[chunkCoord] = chunk
			}

			// Для каждой клетки в чанке
			for x := 0; x < bm.chunkSize; x++ {
				for z := 0; z < bm.chunkSize; z++ {
					// Вычисляем мировые координаты
					worldX := float64(chunkX*bm.chunkSize + x)
					worldZ := float64(chunkZ*bm.chunkSize + z)

					// Вычисляем расстояние до центра метаморфозы
					dx := worldX - centerX
					dz := worldZ - centerZ
					dist := math.Sqrt(dx*dx + dz*dz)

					// Если клетка в пределах радиуса метаморфозы
					if dist <= radius {
						// Применяем метаморфозу
						if chunk.MetamorpLevel[x][z] < order {
							chunk.MetamorpLevel[x][z] = order

							// Обновляем биом, если нужно
							if order >= 4 {
								chunk.Biomes[x][z] = BiomeDistorted
							} else {
								// Возможно изменить биом на основе порядка метаморфозы
								// Или оставить как есть для более сложной логики
							}

							// Можно также обновить шумовые значения, если требуется
						}
					}
				}
			}
		}
	}
}

// generateBiomeChunk генерирует новый чанк биомов на указанных координатах
func (bm *BiomeMap) generateBiomeChunk(coord ChunkCoord) *BiomeChunk {
	chunk := &BiomeChunk{
		Biomes:        make([][]BiomeType, bm.chunkSize),
		NoiseValues:   make(map[string][][]float64),
		MetamorpLevel: make([][]int, bm.chunkSize),
	}

	// Инициализируем массивы
	for x := 0; x < bm.chunkSize; x++ {
		chunk.Biomes[x] = make([]BiomeType, bm.chunkSize)
		chunk.MetamorpLevel[x] = make([]int, bm.chunkSize)
	}

	// Создаем карты шума
	noiseTypes := []string{"elevation", "humidity", "temperature", "anomaly"}
	for _, noiseType := range noiseTypes {
		chunk.NoiseValues[noiseType] = make([][]float64, bm.chunkSize)
		for x := 0; x < bm.chunkSize; x++ {
			chunk.NoiseValues[noiseType][x] = make([]float64, bm.chunkSize)
		}
	}

	// Генерируем значения шума
	// Это упрощенная версия, в реальности используйте более сложные шумовые функции
	chunkSeed := bm.seed + int64(coord.X*10000+coord.Z)
	noise := rand.New(rand.NewSource(chunkSeed))

	// Генерируем шумы
	for x := 0; x < bm.chunkSize; x++ {
		for z := 0; z < bm.chunkSize; z++ {
			worldX := float64(coord.X*bm.chunkSize + x)
			worldZ := float64(coord.Z*bm.chunkSize + z)

			// Генерируем базовые шумовые значения
			// В реальной реализации будет персистентный шум Перлина
			chunk.NoiseValues["elevation"][x][z] = simplexNoise(worldX*0.01, worldZ*0.01, chunkSeed)
			chunk.NoiseValues["humidity"][x][z] = simplexNoise(worldX*0.02, worldZ*0.02, chunkSeed+10)
			chunk.NoiseValues["temperature"][x][z] = simplexNoise(worldX*0.005, worldZ*0.005, chunkSeed+20)
			chunk.NoiseValues["anomaly"][x][z] = simplexNoise(worldX*0.03, worldZ*0.03, chunkSeed+30)

			// Определяем биом на основе шумовых значений
			noiseValues := map[string]float64{
				"elevation":   chunk.NoiseValues["elevation"][x][z],
				"humidity":    chunk.NoiseValues["humidity"][x][z],
				"temperature": chunk.NoiseValues["temperature"][x][z],
				"anomaly":     chunk.NoiseValues["anomaly"][x][z],
			}

			// Используем BiomeManager для определения типа биома
			chunk.Biomes[x][z] = bm.biomeManager.GetBiomeAtPosition(worldX, worldZ, noiseValues)

			// Инициализируем уровень метаморфозы как 0
			chunk.MetamorpLevel[x][z] = 0

			// Если уровень аномалии высок, возможно начальное искажение
			if chunk.NoiseValues["anomaly"][x][z] > 0.8 {
				chunk.MetamorpLevel[x][z] = int(noise.Float64() * 3) // от 0 до 2
			}
		}
	}

	return chunk
}

// Очень простая реализация шума для примера
// В реальности рекомендуется использовать полноценную библиотеку шума или реализацию Perlin/Simplex noise
func simplexNoise(x, y float64, seed int64) float64 {
	// Это очень упрощенная версия, не настоящий Simplex noise
	// Просто для демонстрации
	return math.Sin(x+float64(seed)*0.1)*math.Cos(y+float64(seed)*0.1)*0.5 + 0.5
}

// GetBiome возвращает подробную информацию о биоме указанного типа
func (bm *BiomeMap) GetBiome(biomeType BiomeType) *Biome {
	return bm.biomeManager.GetBiome(biomeType)
}
