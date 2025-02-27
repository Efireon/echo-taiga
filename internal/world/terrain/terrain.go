package terrain

import (
	"math"
	"math/rand"

	"echo-taiga/internal/world/biomes"
)

// TerrainData содержит данные о рельефе чанка
type TerrainData struct {
	Width       int
	Height      int
	HeightMap   [][]float64      // Высоты ландшафта
	GroundTypes [][]string       // Типы поверхности (grass, rock, snow, etc.)
	Features    []TerrainFeature // Особенности рельефа (cliffs, rivers, etc.)
}

// TerrainFeature представляет особую черту ландшафта
type TerrainFeature struct {
	Type     string             // cliff, river, lake, etc.
	Position [2]float64         // Позиция в локальных координатах чанка
	Rotation float64            // Поворот в радианах
	Scale    float64            // Масштаб
	Params   map[string]float64 // Дополнительные параметры
}

// Generator отвечает за процедурную генерацию террейна
type Generator struct {
	Seed              int64
	BiomeMap          *biomes.BiomeMap
	NoiseScale        float64
	HeightScale       float64
	Random            *rand.Rand
	FeatureGenerators map[string]func(*TerrainData, int, int, *rand.Rand) // Генераторы особенностей по типам
}

// NewTerrainData создает новый пустой террейн указанных размеров
func NewTerrainData(width, height int) *TerrainData {
	// Инициализируем высоты ландшафта
	heightMap := make([][]float64, width)
	for i := range heightMap {
		heightMap[i] = make([]float64, height)
	}

	// Инициализируем типы поверхности
	groundTypes := make([][]string, width)
	for i := range groundTypes {
		groundTypes[i] = make([]string, height)
		for j := range groundTypes[i] {
			groundTypes[i][j] = "grass" // По умолчанию - трава
		}
	}

	return &TerrainData{
		Width:       width,
		Height:      height,
		HeightMap:   heightMap,
		GroundTypes: groundTypes,
		Features:    make([]TerrainFeature, 0),
	}
}

// GetHeight возвращает высоту в указанной точке
func (t *TerrainData) GetHeight(x, y int) float64 {
	if x < 0 || x >= t.Width || y < 0 || y >= t.Height {
		return 0 // За пределами террейна
	}
	return t.HeightMap[x][y]
}

// GetHeightAt возвращает высоту в указанной позиции с интерполяцией
func (t *TerrainData) GetHeightAt(x, z float64) float64 {
	// Приводим координаты к локальным координатам чанка
	if x < 0 || x >= float64(t.Width) || z < 0 || z >= float64(t.Height) {
		return 0 // За пределами террейна
	}

	// Определяем ближайшие целые координаты
	x0 := int(math.Floor(x))
	z0 := int(math.Floor(z))
	x1 := x0 + 1
	z1 := z0 + 1

	// Ограничиваем координаты размерами чанка
	if x1 >= t.Width {
		x1 = t.Width - 1
	}
	if z1 >= t.Height {
		z1 = t.Height - 1
	}

	// Вычисляем коэффициенты для билинейной интерполяции
	dx := x - float64(x0)
	dz := z - float64(z0)

	// Получаем высоты в четырех ближайших точках
	h00 := t.HeightMap[x0][z0]
	h10 := t.HeightMap[x1][z0]
	h01 := t.HeightMap[x0][z1]
	h11 := t.HeightMap[x1][z1]

	// Выполняем билинейную интерполяцию
	h0 := h00*(1-dx) + h10*dx
	h1 := h01*(1-dx) + h11*dx
	height := h0*(1-dz) + h1*dz

	return height
}

// GetGroundType возвращает тип поверхности в указанной точке
func (t *TerrainData) GetGroundType(x, y int) string {
	if x < 0 || x >= t.Width || y < 0 || y >= t.Height {
		return "grass" // По умолчанию
	}
	return t.GroundTypes[x][y]
}

// SetHeight устанавливает высоту в указанной точке
func (t *TerrainData) SetHeight(x, y int, height float64) {
	if x < 0 || x >= t.Width || y < 0 || y >= t.Height {
		return // За пределами террейна
	}
	t.HeightMap[x][y] = height
}

// SetGroundType устанавливает тип поверхности в указанной точке
func (t *TerrainData) SetGroundType(x, y int, groundType string) {
	if x < 0 || x >= t.Width || y < 0 || y >= t.Height {
		return // За пределами террейна
	}
	t.GroundTypes[x][y] = groundType
}

// AddFeature добавляет особую черту ландшафта
func (t *TerrainData) AddFeature(feature TerrainFeature) {
	t.Features = append(t.Features, feature)
}

// ApplyDistortion применяет искажение к высотам ландшафта
func (t *TerrainData) ApplyDistortion(intensity float64) {
	// Создаем временный генератор случайных чисел
	r := rand.New(rand.NewSource(int64(len(t.Features))))

	// Применяем случайные искажения к высотам
	for x := 0; x < t.Width; x++ {
		for y := 0; y < t.Height; y++ {
			// Добавляем случайное отклонение, пропорциональное интенсивности
			noise := (r.Float64()*2 - 1) * intensity * 5.0
			t.HeightMap[x][y] += noise
		}
	}

	// С определенной вероятностью добавляем особенности ландшафта
	if r.Float64() < intensity {
		// Добавляем кратер или холм
		centerX := r.Float64() * float64(t.Width)
		centerY := r.Float64() * float64(t.Height)
		radius := 5.0 + r.Float64()*15.0
		depth := (r.Float64()*2 - 1) * intensity * 10.0 // Отрицательные значения - впадины, положительные - холмы

		for x := 0; x < t.Width; x++ {
			for y := 0; y < t.Height; y++ {
				// Расстояние до центра особенности
				distance := math.Sqrt(math.Pow(float64(x)-centerX, 2) + math.Pow(float64(y)-centerY, 2))
				if distance < radius {
					// Плавно уменьшаем эффект от центра к краям
					factor := 1.0 - (distance / radius)
					factor = factor * factor // Квадратичное затухание для более плавного перехода
					t.HeightMap[x][y] += depth * factor
				}
			}
		}
	}
}

// NewGenerator создает новый генератор террейна
func NewGenerator(seed int64, biomeMap *biomes.BiomeMap) *Generator {
	gen := &Generator{
		Seed:              seed,
		BiomeMap:          biomeMap,
		NoiseScale:        0.01, // Масштаб шума Перлина
		HeightScale:       50.0, // Масштаб высот
		Random:            rand.New(rand.NewSource(seed)),
		FeatureGenerators: make(map[string]func(*TerrainData, int, int, *rand.Rand)),
	}

	// Регистрируем генераторы особенностей ландшафта
	gen.registerFeatureGenerators()

	return gen
}

// registerFeatureGenerators регистрирует функции для генерации особенностей ландшафта
func (g *Generator) registerFeatureGenerators() {
	// Генератор реки
	g.FeatureGenerators["river"] = func(terrain *TerrainData, chunkX, chunkY int, r *rand.Rand) {
		// Определяем начальную и конечную точки реки в пределах чанка
		startX := r.Float64() * float64(terrain.Width)
		startY := 0.0
		endX := r.Float64() * float64(terrain.Width)
		endY := float64(terrain.Height)

		// Ширина реки
		width := 3.0 + r.Float64()*5.0

		// Глубина реки
		depth := 2.0 + r.Float64()*3.0

		// Создаем извилистую реку с помощью кривой Безье
		controlX := startX + (endX-startX)*0.5 + (r.Float64()*2-1)*20.0
		controlY := (startY + endY) * 0.5

		// Генерируем точки по длине реки
		for t := 0.0; t <= 1.0; t += 0.02 {
			// Вычисляем точку на кривой Безье
			px := (1-t)*(1-t)*startX + 2*(1-t)*t*controlX + t*t*endX
			py := (1-t)*(1-t)*startY + 2*(1-t)*t*controlY + t*t*endY

			// Преобразуем в целые координаты
			x := int(px)
			y := int(py)

			// Проверяем, что точка внутри террейна
			if x < 0 || x >= terrain.Width || y < 0 || y >= terrain.Height {
				continue
			}

			// Создаем углубление для реки
			for dx := -int(width); dx <= int(width); dx++ {
				for dy := -int(width); dy <= int(width); dy++ {
					nx := x + dx
					ny := y + dy

					// Проверяем, что точка внутри террейна
					if nx < 0 || nx >= terrain.Width || ny < 0 || ny >= terrain.Height {
						continue
					}

					// Вычисляем расстояние до центра реки
					distance := math.Sqrt(float64(dx*dx + dy*dy))
					if distance <= width {
						// Вычисляем глубину в зависимости от расстояния до центра
						depthFactor := 1.0 - (distance / width)

						// Уменьшаем высоту для создания русла реки
						currentHeight := terrain.HeightMap[nx][ny]
						newHeight := currentHeight - depth*depthFactor
						terrain.HeightMap[nx][ny] = newHeight

						// Устанавливаем тип поверхности как "вода"
						if distance < width*0.7 {
							terrain.GroundTypes[nx][ny] = "water"
						} else {
							// Берега реки
							terrain.GroundTypes[nx][ny] = "mud"
						}
					}
				}
			}
		}

		// Добавляем особенность в список
		terrain.AddFeature(TerrainFeature{
			Type:     "river",
			Position: [2]float64{startX, startY}, // Начальная точка реки
			Rotation: math.Atan2(endY-startY, endX-startX),
			Scale:    width,
			Params: map[string]float64{
				"start_x": startX,
				"start_y": startY,
				"end_x":   endX,
				"end_y":   endY,
				"width":   width,
				"depth":   depth,
			},
		})
	}

	// Генератор скал
	g.FeatureGenerators["cliff"] = func(terrain *TerrainData, chunkX, chunkY int, r *rand.Rand) {
		// Определяем центр скалы в пределах чанка
		centerX := r.Float64() * float64(terrain.Width)
		centerY := r.Float64() * float64(terrain.Height)

		// Размер скалы
		radius := 5.0 + r.Float64()*10.0

		// Высота скалы
		height := 10.0 + r.Float64()*20.0

		// Создаем скалу
		for x := 0; x < terrain.Width; x++ {
			for y := 0; y < terrain.Height; y++ {
				// Расстояние до центра скалы
				distance := math.Sqrt(math.Pow(float64(x)-centerX, 2) + math.Pow(float64(y)-centerY, 2))

				if distance < radius {
					// Вычисляем высоту в зависимости от расстояния до центра
					// Чем ближе к центру, тем выше
					heightFactor := 1.0 - (distance / radius)
					heightFactor = heightFactor * heightFactor * heightFactor // Кубическое затухание для крутых скал

					// Повышаем высоту
					currentHeight := terrain.HeightMap[x][y]
					newHeight := currentHeight + height*heightFactor
					terrain.HeightMap[x][y] = newHeight

					// Устанавливаем тип поверхности как "скала"
					if heightFactor > 0.5 {
						terrain.GroundTypes[x][y] = "rock"
					}
				}
			}
		}

		// Добавляем особенность в список
		terrain.AddFeature(TerrainFeature{
			Type:     "cliff",
			Position: [2]float64{centerX, centerY},
			Rotation: 0,
			Scale:    radius,
			Params: map[string]float64{
				"center_x": centerX,
				"center_y": centerY,
				"radius":   radius,
				"height":   height,
			},
		})
	}

	// Генератор озера
	g.FeatureGenerators["lake"] = func(terrain *TerrainData, chunkX, chunkY int, r *rand.Rand) {
		// Определяем центр озера в пределах чанка
		centerX := r.Float64() * float64(terrain.Width)
		centerY := r.Float64() * float64(terrain.Height)

		// Размер озера (эллипс)
		radiusX := 10.0 + r.Float64()*15.0
		radiusY := 10.0 + r.Float64()*15.0

		// Глубина озера
		depth := 5.0 + r.Float64()*5.0

		// Вращение эллипса
		rotation := r.Float64() * math.Pi
		cosRotation := math.Cos(rotation)
		sinRotation := math.Sin(rotation)

		// Создаем озеро
		for x := 0; x < terrain.Width; x++ {
			for y := 0; y < terrain.Height; y++ {
				// Преобразуем координаты с учетом вращения
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				rotatedX := dx*cosRotation - dy*sinRotation
				rotatedY := dx*sinRotation + dy*cosRotation

				// Проверяем, попадает ли точка внутрь эллипса
				normalizedX := rotatedX / radiusX
				normalizedY := rotatedY / radiusY
				distance := normalizedX*normalizedX + normalizedY*normalizedY

				if distance < 1.0 {
					// Вычисляем глубину в зависимости от расстояния до центра
					depthFactor := 1.0 - math.Sqrt(distance)

					// Понижаем высоту для создания чаши озера
					baseHeight := terrain.HeightMap[x][y]
					newHeight := baseHeight - depth*depthFactor
					terrain.HeightMap[x][y] = newHeight

					// Устанавливаем тип поверхности в зависимости от расстояния
					if distance < 0.8 {
						terrain.GroundTypes[x][y] = "water"
					} else {
						// Берега озера
						terrain.GroundTypes[x][y] = "mud"
					}
				}
			}
		}

		// Добавляем особенность в список
		terrain.AddFeature(TerrainFeature{
			Type:     "lake",
			Position: [2]float64{centerX, centerY},
			Rotation: rotation,
			Scale:    (radiusX + radiusY) / 2.0,
			Params: map[string]float64{
				"center_x": centerX,
				"center_y": centerY,
				"radius_x": radiusX,
				"radius_y": radiusY,
				"rotation": rotation,
				"depth":    depth,
			},
		})
	}

	// Генератор поляны
	g.FeatureGenerators["clearing"] = func(terrain *TerrainData, chunkX, chunkY int, r *rand.Rand) {
		// Определяем центр поляны в пределах чанка
		centerX := r.Float64() * float64(terrain.Width)
		centerY := r.Float64() * float64(terrain.Height)

		// Размер поляны
		radius := 8.0 + r.Float64()*10.0

		// Сглаживание высот для создания ровной поверхности
		for x := 0; x < terrain.Width; x++ {
			for y := 0; y < terrain.Height; y++ {
				// Расстояние до центра поляны
				distance := math.Sqrt(math.Pow(float64(x)-centerX, 2) + math.Pow(float64(y)-centerY, 2))

				if distance < radius {
					// Вычисляем фактор сглаживания в зависимости от расстояния до центра
					smoothFactor := 1.0 - (distance / radius)
					smoothFactor = smoothFactor * smoothFactor // Квадратичное затухание для плавного перехода

					// Вычисляем среднюю высоту поляны
					baseHeight := terrain.HeightMap[x][y]

					// Сглаживаем высоту
					terrain.HeightMap[x][y] = baseHeight*(1.0-smoothFactor) + baseHeight*smoothFactor

					// Устанавливаем тип поверхности как "поляна"
					if distance < radius*0.8 {
						terrain.GroundTypes[x][y] = "grass"
					}
				}
			}
		}

		// Добавляем особенность в список
		terrain.AddFeature(TerrainFeature{
			Type:     "clearing",
			Position: [2]float64{centerX, centerY},
			Rotation: 0,
			Scale:    radius,
			Params: map[string]float64{
				"center_x": centerX,
				"center_y": centerY,
				"radius":   radius,
			},
		})
	}
}

// GenerateChunkTerrain генерирует террейн для чанка
func (g *Generator) GenerateChunkTerrain(chunkX, chunkY, chunkSize int) *TerrainData {
	// Создаем новый террейн
	terrain := NewTerrainData(chunkSize, chunkSize)

	// Сид для этого чанка
	chunkSeed := g.Seed + int64(chunkX*10000) + int64(chunkY)
	r := rand.New(rand.NewSource(chunkSeed))

	// Определяем тип биома для чанка
	biomeType := g.BiomeMap.GetBiomeAt(float64(chunkX), float64(chunkY))

	// Генерируем высоты в зависимости от биома
	g.generateHeights(terrain, chunkX, chunkY, biomeType, r)

	// Определяем типы поверхности
	g.assignGroundTypes(terrain, biomeType, r)

	// Добавляем особенности ландшафта
	g.addTerrainFeatures(terrain, chunkX, chunkY, biomeType, r)

	return terrain
}

// generateHeights генерирует высоты для террейна
func (g *Generator) generateHeights(terrain *TerrainData, chunkX, chunkY int, biomeType string, r *rand.Rand) {
	// Базовое смещение чанка в мировых координатах
	worldOffsetX := chunkX * terrain.Width
	worldOffsetY := chunkY * terrain.Height

	// Настройки шума в зависимости от биома
	noiseScale := g.NoiseScale
	heightScale := g.HeightScale
	octaves := 4
	persistence := 0.5
	lacunarity := 2.0

	// Настраиваем параметры в зависимости от биома
	switch biomeType {
	case "taiga":
		heightScale *= 1.2
		persistence = 0.6
	case "marsh":
		heightScale *= 0.7
		persistence = 0.4
	case "rocky":
		heightScale *= 1.5
		persistence = 0.7
		lacunarity = 2.5
	}

	// Генерируем высоты с помощью шума Перлина
	for x := 0; x < terrain.Width; x++ {
		for y := 0; y < terrain.Height; y++ {
			// Мировые координаты
			worldX := float64(worldOffsetX+x) * noiseScale
			worldY := float64(worldOffsetY+y) * noiseScale

			// Суммируем октавы шума
			amplitude := 1.0
			frequency := 1.0
			height := 0.0

			for i := 0; i < octaves; i++ {
				// Используем разные сиды для каждой октавы
				octaveSeed := g.Seed + int64(i*1000)

				// Получаем значение шума Перлина
				noiseValue := g.perlinNoise(worldX*frequency, worldY*frequency, octaveSeed)

				height += noiseValue * amplitude

				amplitude *= persistence
				frequency *= lacunarity
			}

			// Нормализуем и масштабируем высоту
			height = (height + 1.0) / 2.0 * heightScale

			// Добавляем случайные вариации
			height += (r.Float64()*2.0 - 1.0) * 0.5

			// Устанавливаем высоту в террейне
			terrain.SetHeight(x, y, height)
		}
	}
}

// assignGroundTypes определяет типы поверхности на основе высот и биома
func (g *Generator) assignGroundTypes(terrain *TerrainData, biomeType string, r *rand.Rand) {
	// Находим минимальную и максимальную высоты
	minHeight := math.MaxFloat64
	maxHeight := -math.MaxFloat64

	for x := 0; x < terrain.Width; x++ {
		for y := 0; y < terrain.Height; y++ {
			height := terrain.GetHeight(x, y)
			if height < minHeight {
				minHeight = height
			}
			if height > maxHeight {
				maxHeight = height
			}
		}
	}

	// Нормализуем высоты для определения типов поверхности
	heightRange := maxHeight - minHeight
	if heightRange < 0.001 {
		heightRange = 0.001 // Избегаем деления на ноль
	}

	// Устанавливаем типы поверхности в зависимости от высоты и биома
	for x := 0; x < terrain.Width; x++ {
		for y := 0; y < terrain.Height; y++ {
			height := terrain.GetHeight(x, y)
			normalizedHeight := (height - minHeight) / heightRange

			// Случайный фактор для вариации границ
			randomFactor := r.Float64()*0.1 - 0.05

			// Определяем тип поверхности в зависимости от биома и нормализованной высоты
			switch biomeType {
			case "taiga":
				if normalizedHeight < 0.2+randomFactor {
					terrain.SetGroundType(x, y, "mud")
				} else if normalizedHeight < 0.8+randomFactor {
					terrain.SetGroundType(x, y, "grass")
				} else {
					terrain.SetGroundType(x, y, "rock")
				}
			case "marsh":
				if normalizedHeight < 0.3+randomFactor {
					terrain.SetGroundType(x, y, "water")
				} else if normalizedHeight < 0.6+randomFactor {
					terrain.SetGroundType(x, y, "mud")
				} else {
					terrain.SetGroundType(x, y, "grass")
				}
			case "rocky":
				if normalizedHeight < 0.3+randomFactor {
					terrain.SetGroundType(x, y, "grass")
				} else if normalizedHeight < 0.7+randomFactor {
					terrain.SetGroundType(x, y, "rock")
				} else {
					terrain.SetGroundType(x, y, "snow")
				}
			default: // Для любых других биомов
				if normalizedHeight < 0.3+randomFactor {
					terrain.SetGroundType(x, y, "mud")
				} else if normalizedHeight < 0.7+randomFactor {
					terrain.SetGroundType(x, y, "grass")
				} else {
					terrain.SetGroundType(x, y, "rock")
				}
			}
		}
	}
}

// addTerrainFeatures добавляет особенности ландшафта
func (g *Generator) addTerrainFeatures(terrain *TerrainData, chunkX, chunkY int, biomeType string, r *rand.Rand) {
	// Определяем, какие особенности добавлять в зависимости от биома
	var featureTypes []string

	switch biomeType {
	case "taiga":
		featureTypes = []string{"clearing"}
		// С небольшой вероятностью добавляем озеро
		if r.Float64() < 0.2 {
			featureTypes = append(featureTypes, "lake")
		}
	case "marsh":
		featureTypes = []string{"lake", "river"}
	case "rocky":
		featureTypes = []string{"cliff"}
		// С небольшой вероятностью добавляем озеро
		if r.Float64() < 0.1 {
			featureTypes = append(featureTypes, "lake")
		}
	default:
		featureTypes = []string{"clearing"}
	}

	// Проверяем наличие реки в соседних чанках, чтобы продолжить ее
	// TODO: Реализовать проверку соседних чанков для связности рек

	// Добавляем 1-3 случайных особенности
	featureCount := 1 + r.Intn(2) // 1-2 особенности

	for i := 0; i < featureCount; i++ {
		// Случайно выбираем тип особенности
		if len(featureTypes) == 0 {
			continue
		}

		featureTypeIndex := r.Intn(len(featureTypes))
		featureType := featureTypes[featureTypeIndex]

		// Получаем генератор для этого типа особенности
		generator, exists := g.FeatureGenerators[featureType]
		if !exists {
			continue
		}

		// Вызываем генератор для создания особенности
		generator(terrain, chunkX, chunkY, r)
	}
}

// perlinNoise генерирует значение шума Перлина в указанной точке
func (g *Generator) perlinNoise(x, y float64, seed int64) float64 {
	// В реальном приложении здесь была бы реализация шума Перлина
	// Или использование библиотеки для шума

	// Для простоты используем упрощенную версию на основе синусоид
	// Это не настоящий шум Перлина, но даст схожие результаты для примера
	r := rand.New(rand.NewSource(seed))

	// Генерируем случайные смещения для каждого измерения
	offsetX := r.Float64() * 1000
	offsetY := r.Float64() * 1000

	// Используем синусоиды для создания "шумного" значения
	value := math.Sin(x*10+offsetX) * math.Cos(y*10+offsetY) * 0.5
	value += math.Sin((x+y)*20+offsetX+offsetY) * 0.25
	value += math.Sin(x*5) * math.Sin(y*5) * 0.25

	return value
}
