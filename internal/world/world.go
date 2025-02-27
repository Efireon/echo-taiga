package world

import (
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"time"

	"echo-taiga/internal/engine/ecs"
	"echo-taiga/internal/metamorphosis"
	"echo-taiga/internal/world/biomes"
	"echo-taiga/internal/world/terrain"
)

// ChunkSize определяет размер чанка в игровых единицах
const ChunkSize = 64

// ViewDistance определяет, сколько чанков вокруг игрока должны быть активны
const ViewDistance = 3

// Chunk представляет одну часть игрового мира
type Chunk struct {
	Position         [2]int
	Terrain          *terrain.TerrainData
	Entities         []ecs.EntityID
	MetamorphEffects []string // ID эффектов метаморфоза
	IsGenerated      bool
	IsActive         bool
	AnomalyLevel     float64 // Уровень аномальности чанка (0-1)
	LastVisited      int64   // Время последнего посещения игроком
	BiomeType        string  // Тип биома в этом чанке
}

// World представляет весь игровой мир
type World struct {
	Seed               int64
	Chunks             map[[2]int]*Chunk
	ActiveChunks       map[[2]int]*Chunk
	BiomeMap           *biomes.BiomeMap
	TimeOfDay          float64 // 0.0 - 1.0, где 0.0 - полночь, 0.5 - полдень
	MetamorphManager   *metamorphosis.MetamorphosisManager
	ECSWorld           *ecs.World
	PlayerPosition     ecs.Vector3
	GlobalAnomalyLevel float64                   // Общий уровень аномальности мира
	WeatherCondition   string                    // Текущие погодные условия
	ChunkEntities      map[[2]int][]ecs.EntityID // Кэш сущностей по чанкам
	TerrainGenerator   *terrain.Generator
}

// NewWorld создает новый мир с указанным сидом
func NewWorld(seed int64, ecsWorld *ecs.World) *World {
	world := &World{
		Seed:               seed,
		Chunks:             make(map[[2]int]*Chunk),
		ActiveChunks:       make(map[[2]int]*Chunk),
		TimeOfDay:          0.25, // Начинаем с рассвета
		ECSWorld:           ecsWorld,
		ChunkEntities:      make(map[[2]int][]ecs.EntityID),
		GlobalAnomalyLevel: 0.1, // Начальный низкий уровень аномальности
		WeatherCondition:   "clear",
	}

	// Инициализируем биомы
	world.BiomeMap = biomes.NewBiomeMap(seed)

	// Инициализируем генератор террейна
	world.TerrainGenerator = terrain.NewGenerator(seed, world.BiomeMap)

	// Инициализируем менеджер метаморфоз
	world.MetamorphManager = metamorphosis.NewMetamorphosisManager(ecsWorld, "saves/metamorphosis")
	err := world.MetamorphManager.Init()
	if err != nil {
		// Логировать ошибку, но продолжить работу
		println("Failed to initialize metamorphosis manager:", err.Error())
	}

	return world
}

func createNightCreature(world *ecs.World, position ecs.Vector3, anomalyLevel float64) *ecs.Entity {
	// Преобразуйте значение здоровья
	healthComp := ecs.NewHealthComponent(float64(50 + int(anomalyLevel*100)))
	// Остальной код без изменений
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

// GetChunkAtPosition возвращает чанк, содержащий указанную мировую позицию
func (w *World) GetChunkAtPosition(worldX, worldZ float64) *Chunk {
	// Преобразуем мировые координаты в координаты чанка
	chunkX := int(math.Floor(worldX / ChunkSize))
	chunkZ := int(math.Floor(worldZ / ChunkSize))

	return w.GetChunkAt(chunkX, chunkZ)
}

// ActivateChunk активирует чанк для обработки
func (w *World) ActivateChunk(x, y int) {
	chunk := w.GetChunkAt(x, y)
	pos := [2]int{x, y}

	if !chunk.IsActive {
		chunk.IsActive = true
		w.ActiveChunks[pos] = chunk

		// Загружаем сущности из чанка, если они были сохранены
		if entities, exists := w.ChunkEntities[pos]; exists {
			for _, entityID := range entities {
				// Получаем сущность из мира ECS
				entity, exists := w.ECSWorld.GetEntity(entityID)
				if exists {
					// Реактивируем сущность
					if entity.HasComponent(ecs.AIComponentID) {
						aiComp, _ := entity.GetComponent(ecs.AIComponentID)
						ai := aiComp.(*ecs.AIComponent)
						ai.SetState("idle") // Сбрасываем состояние ИИ
					}
				}
			}
		}

		// Применяем эффекты метаморфоза к чанку
		w.applyChunkMetamorphoses(chunk)
	}
}

// DeactivateChunk деактивирует чанк
func (w *World) DeactivateChunk(x, y int) {
	pos := [2]int{x, y}
	chunk, exists := w.Chunks[pos]

	if exists && chunk.IsActive {
		chunk.IsActive = false
		delete(w.ActiveChunks, pos)

		// Сохраняем сущности чанка и удаляем их из активного мира
		w.storeChunkEntities(chunk)
	}
}

// storeChunkEntities сохраняет ссылки на сущности чанка и деактивирует их
func (w *World) storeChunkEntities(chunk *Chunk) {
	// Сохраняем текущий список сущностей в кэше
	w.ChunkEntities[[2]int{chunk.Position[0], chunk.Position[1]}] = chunk.Entities

	// Очищаем список сущностей в чанке
	chunk.Entities = []ecs.EntityID{}
}

// UpdateActiveChunks обновляет список активных чанков вокруг игрока
func (w *World) UpdateActiveChunks() {
	// Определяем, в каком чанке находится игрок
	playerChunkX := int(math.Floor(w.PlayerPosition.X / ChunkSize))
	playerChunkZ := int(math.Floor(w.PlayerPosition.Z / ChunkSize))

	// Создаем временную карту для отслеживания чанков, которые должны быть активны
	shouldBeActive := make(map[[2]int]bool)

	// Активируем чанки в радиусе просмотра
	for x := playerChunkX - ViewDistance; x <= playerChunkX+ViewDistance; x++ {
		for z := playerChunkZ - ViewDistance; z <= playerChunkZ+ViewDistance; z++ {
			// Проверяем, находится ли чанк в радиусе просмотра (круглая область)
			distX := x - playerChunkX
			distZ := z - playerChunkZ
			distSquared := distX*distX + distZ*distZ

			if distSquared <= ViewDistance*ViewDistance {
				pos := [2]int{x, z}
				shouldBeActive[pos] = true
				w.ActivateChunk(x, z)
			}
		}
	}

	// Деактивируем чанки, которые вышли из зоны видимости
	for pos := range w.ActiveChunks {
		if !shouldBeActive[pos] {
			w.DeactivateChunk(pos[0], pos[1])
		}
	}
}

// Генерирует новый чанк в указанной позиции
func (w *World) generateChunk(x, y int) *Chunk {
	// Создаем новый чанк
	chunk := &Chunk{
		Position:    [2]int{x, y},
		IsGenerated: true,
		IsActive:    false,
		Entities:    make([]ecs.EntityID, 0),
	}

	// Определяем тип биома для этого чанка
	biomeType := w.BiomeMap.GetBiomeAt(float64(x), float64(y))
	chunk.BiomeType = biomeType

	// Генерируем террейн для чанка
	chunk.Terrain = w.TerrainGenerator.GenerateChunkTerrain(x, y, ChunkSize)

	// Устанавливаем начальный уровень аномальности на основе удаленности от центра
	distFromCenter := math.Sqrt(float64(x*x + y*y))
	chunk.AnomalyLevel = math.Min(1.0, w.GlobalAnomalyLevel+(distFromCenter/10.0)*0.05)

	// Добавляем базовые сущности в зависимости от биома
	w.populateChunkWithEntities(chunk)

	return chunk
}

// populateChunkWithEntities добавляет базовые сущности в чанк в зависимости от биома
func (w *World) populateChunkWithEntities(chunk *Chunk) {
	// Инициализируем генератор псевдослучайных чисел с предсказуемым сидом для этого чанка
	r := rand.New(rand.NewSource(w.Seed + int64(chunk.Position[0]*10000) + int64(chunk.Position[1])))

	// Рассчитываем мировые координаты угла чанка
	worldX := float64(chunk.Position[0] * ChunkSize)
	worldZ := float64(chunk.Position[1] * ChunkSize)

	// Добавляем различные объекты в зависимости от биома
	switch chunk.BiomeType {
	case "taiga":
		// Добавляем деревья
		treeCount := 10 + r.Intn(15) // 10-24 деревьев на чанк
		for i := 0; i < treeCount; i++ {
			// Случайная позиция внутри чанка
			x := worldX + r.Float64()*ChunkSize
			z := worldZ + r.Float64()*ChunkSize

			// Получаем высоту местности в этой точке
			y := chunk.Terrain.GetHeightAt(x-worldX, z-worldZ)

			// Создаем дерево
			treeEntity := createTree(w.ECSWorld, ecs.Vector3{X: x, Y: y, Z: z}, r.Float64())

			// Добавляем в список сущностей чанка
			chunk.Entities = append(chunk.Entities, treeEntity.ID)
		}

		// Добавляем камни
		rockCount := 5 + r.Intn(10) // 5-14 камней на чанк
		for i := 0; i < rockCount; i++ {
			x := worldX + r.Float64()*ChunkSize
			z := worldZ + r.Float64()*ChunkSize
			y := chunk.Terrain.GetHeightAt(x-worldX, z-worldZ)

			rockEntity := createRock(w.ECSWorld, ecs.Vector3{X: x, Y: y, Z: z}, r.Float64())
			chunk.Entities = append(chunk.Entities, rockEntity.ID)
		}

		// Добавляем кусты
		bushCount := 8 + r.Intn(12) // 8-19 кустов на чанк
		for i := 0; i < bushCount; i++ {
			x := worldX + r.Float64()*ChunkSize
			z := worldZ + r.Float64()*ChunkSize
			y := chunk.Terrain.GetHeightAt(x-worldX, z-worldZ)

			bushEntity := createBush(w.ECSWorld, ecs.Vector3{X: x, Y: y, Z: z}, r.Float64())
			chunk.Entities = append(chunk.Entities, bushEntity.ID)
		}

		// С малой вероятностью добавляем особые объекты
		if r.Float64() < 0.2 { // 20% шанс
			// Создаем поляну или особое место
			x := worldX + ChunkSize*0.5 + (r.Float64()-0.5)*ChunkSize*0.5
			z := worldZ + ChunkSize*0.5 + (r.Float64()-0.5)*ChunkSize*0.5
			y := chunk.Terrain.GetHeightAt(x-worldX, z-worldZ)

			clearingEntity := createClearing(w.ECSWorld, ecs.Vector3{X: x, Y: y, Z: z}, r.Float64())
			chunk.Entities = append(chunk.Entities, clearingEntity.ID)
		}

		// С очень малой вероятностью добавляем символ
		if r.Float64() < 0.05 { // 5% шанс
			x := worldX + r.Float64()*ChunkSize
			z := worldZ + r.Float64()*ChunkSize
			y := chunk.Terrain.GetHeightAt(x-worldX, z-worldZ)

			symbolEntity := createSymbol(w.ECSWorld, ecs.Vector3{X: x, Y: y, Z: z}, r.Float64())
			chunk.Entities = append(chunk.Entities, symbolEntity.ID)
		}

	case "marsh":
		// Болотный биом - меньше деревьев, больше воды и особых объектов
		// TODO: Реализовать генерацию для болота

	case "rocky":
		// Скалистый биом - много камней, мало растительности
		// TODO: Реализовать генерацию для скалистой местности
	}

	// Добавляем животных с малой вероятностью
	if r.Float64() < 0.3 { // 30% шанс для чанка иметь животных
		animalCount := 1 + r.Intn(3) // 1-3 животных
		for i := 0; i < animalCount; i++ {
			x := worldX + r.Float64()*ChunkSize
			z := worldZ + r.Float64()*ChunkSize
			y := chunk.Terrain.GetHeightAt(x-worldX, z-worldZ)

			animalEntity := createAnimal(w.ECSWorld, ecs.Vector3{X: x, Y: y, Z: z}, r.Float64())
			chunk.Entities = append(chunk.Entities, animalEntity.ID)
		}
	}
}

// applyChunkMetamorphoses применяет эффекты метаморфоза к чанку
func (w *World) applyChunkMetamorphoses(chunk *Chunk) {
	// Применяем активные метаморфозы из менеджера к чанку
	effects := w.MetamorphManager.GetActiveEffects()

	for _, effect := range effects {
		// Проверяем, применим ли эффект к данному чанку
		if isChunkAffectedByEffect(chunk, effect) {
			// Проверяем, не применен ли уже эффект
			alreadyApplied := false
			for _, appliedID := range chunk.MetamorphEffects {
				if appliedID == effect.ID {
					alreadyApplied = true
					break
				}
			}

			if !alreadyApplied {
				// Добавляем эффект в список примененных
				chunk.MetamorphEffects = append(chunk.MetamorphEffects, effect.ID)

				// Применяем эффект к террейну и сущностям чанка
				applyMetamorphEffectToChunk(w.ECSWorld, chunk, effect)
			}
		}
	}

	// Обновляем уровень аномальности чанка
	chunk.AnomalyLevel = calculateChunkAnomalyLevel(chunk, w.GlobalAnomalyLevel)
}

// isChunkAffectedByEffect определяет, влияет ли эффект метаморфоза на данный чанк
func isChunkAffectedByEffect(chunk *Chunk, effect *metamorphosis.MetamorphEffect) bool {
	// Если у эффекта нет области воздействия, то считаем что он глобальный
	if effect.AffectedArea == nil {
		return true
	}

	// Получаем мировые координаты центра чанка
	chunkCenterX := float64(chunk.Position[0]*ChunkSize) + ChunkSize/2
	chunkCenterZ := float64(chunk.Position[1]*ChunkSize) + ChunkSize/2

	// Вычисляем расстояние от центра чанка до центра эффекта
	effectCenter := effect.AffectedArea.Center
	distance := math.Sqrt(
		math.Pow(chunkCenterX-effectCenter.X, 2) +
			math.Pow(chunkCenterZ-effectCenter.Z, 2),
	)

	// Проверяем, находится ли чанк в радиусе действия эффекта
	switch effect.AffectedArea.Type {
	case "sphere", "cylinder":
		// Учитываем размер чанка (диагональ / 2)
		chunkRadius := ChunkSize * math.Sqrt(2) / 2
		return distance <= effect.AffectedArea.Radius+chunkRadius
	case "box":
		// Проверяем, пересекается ли AABB чанка с AABB эффекта
		halfSize := effect.AffectedArea.Size.Multiply(0.5)
		chunkMin := ecs.Vector3{X: chunkCenterX - ChunkSize/2, Z: chunkCenterZ - ChunkSize/2}
		chunkMax := ecs.Vector3{X: chunkCenterX + ChunkSize/2, Z: chunkCenterZ + ChunkSize/2}
		effectMin := ecs.Vector3{X: effectCenter.X - halfSize.X, Z: effectCenter.Z - halfSize.Z}
		effectMax := ecs.Vector3{X: effectCenter.X + halfSize.X, Z: effectCenter.Z + halfSize.Z}

		return !(chunkMin.X > effectMax.X || chunkMax.X < effectMin.X ||
			chunkMin.Z > effectMax.Z || chunkMax.Z < effectMin.Z)
	}

	return false
}

// applyMetamorphEffectToChunk применяет эффект метаморфоза к чанку
func applyMetamorphEffectToChunk(world *ecs.World, chunk *Chunk, effect *metamorphosis.MetamorphEffect) {
	// Применяем эффект в зависимости от его типа и порядка
	switch effect.Order {
	case metamorphosis.OrderFirst:
		// Визуальные изменения - текстуры, цвета, звуки
		for _, entityID := range chunk.Entities {
			entity, exists := world.GetEntity(entityID)
			if !exists {
				continue
			}

			// Если есть компонент рендера, изменяем его
			if entity.HasComponent(ecs.RenderComponentID) {
				renderComp, _ := entity.GetComponent(ecs.RenderComponentID)
				render := renderComp.(*ecs.RenderComponent)

				// Добавляем визуальные эффекты
				for _, visualEffect := range effect.VisualEffects {
					if !containsString(render.Effects, visualEffect) {
						render.Effects = append(render.Effects, visualEffect)
					}
				}

				// Увеличиваем искажение
				if effect.Category == "visual" {
					render.Distortion += 0.1 * effect.Intensity
					if render.Distortion > 1.0 {
						render.Distortion = 1.0
					}
				}
			}

			// Если есть компонент звука, изменяем его
			if entity.HasComponent(ecs.SoundEmitterComponentID) {
				soundComp, _ := entity.GetComponent(ecs.SoundEmitterComponentID)
				sound := soundComp.(*ecs.SoundEmitterComponent)

				// Добавляем звуковые эффекты
				for _, soundEffect := range effect.SoundEffects {
					if sound.Sounds[soundEffect] != "" {
						sound.SoundID = sound.Sounds[soundEffect]
						sound.Play()
					}
				}
			}
		}

	case metamorphosis.OrderSecond:
		// Структурные изменения - форма объектов, локальные аномалии
		for _, entityID := range chunk.Entities {
			entity, exists := world.GetEntity(entityID)
			if !exists {
				continue
			}

			// Если есть компонент трансформации, изменяем его
			if entity.HasComponent(ecs.TransformComponentID) {
				transformComp, _ := entity.GetComponent(ecs.TransformComponentID)
				transform := transformComp.(*ecs.TransformComponent)

				// Случайные изменения масштаба и поворота
				r := rand.New(rand.NewSource(int64(entityID)))

				// Изменяем масштаб
				scaleChange := 0.2 * effect.Intensity * (r.Float64()*2 - 1)
				transform.Scale = transform.Scale.Add(ecs.Vector3{X: scaleChange, Y: scaleChange, Z: scaleChange})

				// Изменяем вращение
				rotChange := 0.3 * effect.Intensity * (r.Float64()*2 - 1)
				transform.Rotation = transform.Rotation.Add(ecs.Vector3{X: 0, Y: rotChange, Z: 0})
			}

			// Если есть метаморфный компонент, увеличиваем его аномальность
			if entity.HasComponent(ecs.MetamorphicComponentID) {
				metaComp, _ := entity.GetComponent(ecs.MetamorphicComponentID)
				meta := metaComp.(*ecs.MetamorphicComponent)

				meta.AbnormalityIndex += 0.1 * effect.Intensity
				if meta.AbnormalityIndex > 1.0 {
					meta.AbnormalityIndex = 1.0
				}
			}
		}

		// Изменяем террейн - создаем аномалии рельефа
		if effect.Category == "environment" {
			// Например, создаем холмы или впадины
			chunk.Terrain.ApplyDistortion(effect.Intensity)
		}

	case metamorphosis.OrderThird:
		// Функциональные изменения - новые свойства, поведение
		for _, entityID := range chunk.Entities {
			entity, exists := world.GetEntity(entityID)
			if !exists {
				continue
			}

			// Изменяем поведение ИИ, если есть
			if entity.HasComponent(ecs.AIComponentID) {
				aiComp, _ := entity.GetComponent(ecs.AIComponentID)
				ai := aiComp.(*ecs.AIComponent)

				// Увеличиваем агрессивность в зависимости от интенсивности
				if effect.Intensity > 0.5 && ai.AIType == "neutral" {
					ai.AIType = "aggressive"
					ai.DetectionRange *= 1.5
				}

				// Увеличиваем урон
				ai.AttackDamage *= 1.0 + (effect.Intensity * 0.3)
			}

			// Добавляем новые способности и свойства
			if entity.HasTag("animal") && effect.Intensity > 0.7 {
				// Создаем мутировавшее животное с новыми способностями
				entity.AddTag("mutated")

				// Добавляем свечение
				if entity.HasComponent(ecs.RenderComponentID) {
					renderComp, _ := entity.GetComponent(ecs.RenderComponentID)
					render := renderComp.(*ecs.RenderComponent)
					render.Effects = append(render.Effects, "glow")
				}

				// Добавляем компонент здоровья или усиливаем его
				if entity.HasComponent(ecs.HealthComponentID) {
					healthComp, _ := entity.GetComponent(ecs.HealthComponentID)
					health := healthComp.(*ecs.HealthComponent)
					health.MaxHealth *= 1.5
					health.CurrentHealth = health.MaxHealth
				}
			}
		}

		// Создаем новые сущности или изменяем физику
		if effect.Category == "reality" {
			// TODO: Реализовать изменения физики для всего чанка
		}

	case metamorphosis.OrderFourth:
		// Системные изменения - физика, пространство, правила
		// На этом уровне изменения затрагивают все аспекты чанка

		// Изменяем физику всех сущностей
		for _, entityID := range chunk.Entities {
			entity, exists := world.GetEntity(entityID)
			if !exists {
				continue
			}

			if entity.HasComponent(ecs.PhysicsComponentID) {
				physicsComp, _ := entity.GetComponent(ecs.PhysicsComponentID)
				physics := physicsComp.(*ecs.PhysicsComponent)

				// Изменяем гравитацию
				if effect.WorldChanges["physics.gravity"] != 0 {
					physics.Gravity = effect.WorldChanges["physics.gravity"]
				}

				// Изменяем трение
				if effect.WorldChanges["physics.friction"] != 0 {
					physics.Friction = effect.WorldChanges["physics.friction"]
				}
			}

			// Радикально изменяем внешний вид
			if entity.HasComponent(ecs.RenderComponentID) {
				renderComp, _ := entity.GetComponent(ecs.RenderComponentID)
				render := renderComp.(*ecs.RenderComponent)

				// Сильные визуальные искажения
				if effect.WorldChanges["rendering.distortion"] != 0 {
					render.Distortion = effect.WorldChanges["rendering.distortion"]
				}

				// Добавляем эффекты
				for _, visualEffect := range effect.VisualEffects {
					if !containsString(render.Effects, visualEffect) {
						render.Effects = append(render.Effects, visualEffect)
					}
				}

				// Изменяем цвет
				if effect.Category == "reality" {
					// Радикальная смена цвета
					render.Color.R = uint8(rand.Intn(255))
					render.Color.G = uint8(rand.Intn(255))
					render.Color.B = uint8(rand.Intn(255))
				}
			}
		}

		// Изменяем террейн более радикально
		chunk.Terrain.ApplyDistortion(effect.Intensity * 2)
		// Можно создать порталы, разломы и т.д.

	case metamorphosis.OrderFifth:
		// Фундаментальные изменения - новые механики, цели
		// Этот уровень может полностью изменить игровой процесс

		// Радикальное преобразование всего чанка
		if effect.Category == "reality" && effect.WorldChanges["reality.consistency"] < 0 {
			// Чанк становится "нереальным" - может содержать невозможные объекты
			chunk.BiomeType = "void" // Изменяем биом на "пустоту"

			// Заменяем все обычные объекты на искаженные версии
			var newEntities []ecs.EntityID
			for _, entityID := range chunk.Entities {
				entity, exists := world.GetEntity(entityID)
				if !exists {
					continue
				}

				// Удаляем старую сущность
				world.RemoveEntity(entityID)

				// Создаем искаженную версию
				newEntity := createDistortedEntity(world, entity)
				newEntities = append(newEntities, newEntity.ID)
			}

			// Обновляем список сущностей
			chunk.Entities = newEntities

			// Полностью преобразуем террейн
			chunk.Terrain = createVoidTerrain(chunk.Position[0], chunk.Position[1])
		}

		// Активируем новые механики
		if effect.WorldChanges["mechanics.core"] > 0 {
			// TODO: Реализовать активацию новых механик
		}
	}
}

// calculateChunkAnomalyLevel вычисляет уровень аномальности чанка
func calculateChunkAnomalyLevel(chunk *Chunk, globalLevel float64) float64 {
	baseLevel := globalLevel

	// Учитываем количество метаморфоз
	metaBonus := float64(len(chunk.MetamorphEffects)) * 0.05

	// Учитываем расстояние от центра мира
	distFromCenter := math.Sqrt(float64(chunk.Position[0]*chunk.Position[0] + chunk.Position[1]*chunk.Position[1]))
	distFactor := math.Min(0.3, distFromCenter/30.0*0.3)

	return math.Min(1.0, baseLevel+metaBonus+distFactor)
}

// Update обновляет состояние мира
func (w *World) Update(deltaTime float64) {
	// Обновление времени суток
	w.TimeOfDay += deltaTime * 0.001 // Примерный полный цикл за 1000 секунд
	if w.TimeOfDay >= 1.0 {
		w.TimeOfDay -= 1.0
	}

	// Обновление активных чанков вокруг игрока
	w.UpdateActiveChunks()

	// Обновление каждого активного чанка
	for _, chunk := range w.ActiveChunks {
		w.updateChunk(chunk, deltaTime)
	}

	// Обновление менеджера метаморфоз
	w.MetamorphManager.Update(deltaTime)
}

// SetPlayerPosition устанавливает текущую позицию игрока
func (w *World) SetPlayerPosition(position ecs.Vector3) {
	w.PlayerPosition = position
}

// GetGlobalTimeOfDay возвращает текущее время суток
func (w *World) GetGlobalTimeOfDay() float64 {
	return w.TimeOfDay
}

// SetGlobalAnomalyLevel устанавливает глобальный уровень аномальности
func (w *World) SetGlobalAnomalyLevel(level float64) {
	w.GlobalAnomalyLevel = math.Max(0.0, math.Min(1.0, level))
}

// GetGlobalAnomalyLevel возвращает глобальный уровень аномальности
func (w *World) GetGlobalAnomalyLevel() float64 {
	return w.GlobalAnomalyLevel
}

// SetWeatherCondition устанавливает текущие погодные условия
func (w *World) SetWeatherCondition(weather string) {
	w.WeatherCondition = weather

	// Обновляем менеджер метаморфоз
	if w.MetamorphManager != nil {
		w.MetamorphManager.SetWeather(weather)
	}
}

// Обновляет состояние отдельного чанка
func (w *World) updateChunk(chunk *Chunk, deltaTime float64) {
	// Обновляем время последнего посещения, если игрок в этом чанке
	chunkX := int(math.Floor(w.PlayerPosition.X / ChunkSize))
	chunkZ := int(math.Floor(w.PlayerPosition.Z / ChunkSize))

	if chunkX == chunk.Position[0] && chunkZ == chunk.Position[1] {
		chunk.LastVisited = int64(time.Now().Unix())
	}

	// Обрабатываем эффекты метаморфоза
	for _, effectID := range chunk.MetamorphEffects {
		effect, exists := w.MetamorphManager.GetEffect(effectID)
		if !exists {
			continue
		}

		// Обновляем эффект, если он имеет функцию обновления
		if effect.OnUpdate != nil {
			for _, entityID := range chunk.Entities {
				entity, exists := w.ECSWorld.GetEntity(entityID)
				if !exists {
					continue
				}

				// Вызываем функцию обновления для каждой сущности в чанке
				effect.OnUpdate(w.ECSWorld, entity, deltaTime)
			}
		}
	}

	// Проверяем условия для спавна новых сущностей
	// Например, с малой вероятностью спавним существ ночью
	if w.TimeOfDay < 0.25 || w.TimeOfDay > 0.75 { // Ночь
		// Изредка спавним ночных существ
		if rand.Float64() < 0.01*deltaTime { // Корректируем по deltaTime для независимости от FPS
			w.spawnNightCreature(chunk)
		}
	}

	// С вероятностью, зависящей от уровня аномальности, спавним аномалии
	if rand.Float64() < chunk.AnomalyLevel*0.005*deltaTime {
		w.spawnAnomaly(chunk)
	}
}

// spawnNightCreature создает ночное существо в чанке
func (w *World) spawnNightCreature(chunk *Chunk) {
	// Случайная позиция в чанке
	worldX := float64(chunk.Position[0] * ChunkSize)
	worldZ := float64(chunk.Position[1] * ChunkSize)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	x := worldX + r.Float64()*ChunkSize
	z := worldZ + r.Float64()*ChunkSize
	y := chunk.Terrain.GetHeightAt(x-worldX, z-worldZ)

	// Создаем ночное существо
	creatureEntity := createNightCreature(w.ECSWorld, ecs.Vector3{X: x, Y: y, Z: z}, chunk.AnomalyLevel)

	// Добавляем в список сущностей чанка
	chunk.Entities = append(chunk.Entities, creatureEntity.ID)
}

// spawnAnomaly создает аномалию в чанке
func (w *World) spawnAnomaly(chunk *Chunk) {
	// Случайная позиция в чанке
	worldX := float64(chunk.Position[0] * ChunkSize)
	worldZ := float64(chunk.Position[1] * ChunkSize)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	x := worldX + r.Float64()*ChunkSize
	z := worldZ + r.Float64()*ChunkSize
	y := chunk.Terrain.GetHeightAt(x-worldX, z-worldZ)

	// Создаем аномалию
	anomalyType := "minor"
	if chunk.AnomalyLevel > 0.5 {
		anomalyType = "medium"
	}
	if chunk.AnomalyLevel > 0.8 {
		anomalyType = "major"
	}

	anomalyEntity := createAnomaly(w.ECSWorld, ecs.Vector3{X: x, Y: y, Z: z}, anomalyType)

	// Добавляем в список сущностей чанка
	chunk.Entities = append(chunk.Entities, anomalyEntity.ID)
}

// Вспомогательные функции для создания различных сущностей

// createTree создает дерево в указанной позиции
func createTree(world *ecs.World, position ecs.Vector3, randomFactor float64) *ecs.Entity {
	tree := ecs.NewEntity()

	// Добавляем базовые компоненты
	tree.AddComponent(ecs.NewTransformComponent(position))

	// Выбираем случайную модель и текстуру дерева
	modelTypes := []string{"pine", "spruce", "fir"}
	modelType := modelTypes[int(randomFactor*float64(len(modelTypes)))]

	renderComp := ecs.NewRenderComponent("tree_"+modelType, "tree_"+modelType+"_texture")
	tree.AddComponent(renderComp)

	// Добавляем физический компонент
	physicsComp := ecs.NewPhysicsComponent(0, true)                // Неподвижный объект
	physicsComp.ColliderSize = ecs.Vector3{X: 1.0, Y: 5.0, Z: 1.0} // Размер коллайдера
	tree.AddComponent(physicsComp)

	// Добавляем метаморфный компонент с высокой стабильностью
	metaComp := ecs.NewMetamorphicComponent(0.8) // Достаточно стабильны, но могут меняться
	tree.AddComponent(metaComp)

	// Добавляем теги
	tree.AddTag("tree")
	tree.AddTag("vegetation")
	tree.AddTag("environment")

	// Случайно варьируем размер
	transformComp, _ := tree.GetComponent(ecs.TransformComponentID)
	transform := transformComp.(*ecs.TransformComponent)

	baseScale := 0.8 + randomFactor*0.4 // 0.8 - 1.2
	transform.Scale = ecs.Vector3{X: baseScale, Y: baseScale + randomFactor*0.3, Z: baseScale}

	// Добавляем сущность в мир
	world.AddEntity(tree)

	return tree
}

// createRock создает камень в указанной позиции
func createRock(world *ecs.World, position ecs.Vector3, randomFactor float64) *ecs.Entity {
	rock := ecs.NewEntity()

	// Добавляем базовые компоненты
	rock.AddComponent(ecs.NewTransformComponent(position))

	// Выбираем случайную модель и текстуру камня
	modelTypes := []string{"boulder", "rock", "stone"}
	modelType := modelTypes[int(randomFactor*float64(len(modelTypes)))]

	renderComp := ecs.NewRenderComponent("rock_"+modelType, "rock_"+modelType+"_texture")
	rock.AddComponent(renderComp)

	// Добавляем физический компонент
	physicsComp := ecs.NewPhysicsComponent(100, true)              // Тяжелый неподвижный объект
	physicsComp.ColliderSize = ecs.Vector3{X: 1.0, Y: 1.0, Z: 1.0} // Размер коллайдера
	tree.AddComponent(physicsComp)

	// Добавляем метаморфный компонент с очень высокой стабильностью
	metaComp := ecs.NewMetamorphicComponent(0.9) // Очень стабильны, редко меняются
	rock.AddComponent(metaComp)

	// Добавляем теги
	rock.AddTag("rock")
	rock.AddTag("environment")

	// Случайно варьируем размер и поворот
	transformComp, _ := rock.GetComponent(ecs.TransformComponentID)
	transform := transformComp.(*ecs.TransformComponent)

	baseScale := 0.5 + randomFactor*1.0 // 0.5 - 1.5
	transform.Scale = ecs.Vector3{X: baseScale, Y: baseScale * 0.8, Z: baseScale}
	transform.Rotation = ecs.Vector3{X: 0, Y: randomFactor * 6.28, Z: 0} // Случайный поворот вокруг Y

	// Добавляем сущность в мир
	world.AddEntity(rock)

	return rock
}

// createBush создает куст в указанной позиции
func createBush(world *ecs.World, position ecs.Vector3, randomFactor float64) *ecs.Entity {
	bush := ecs.NewEntity()

	// Добавляем базовые компоненты
	bush.AddComponent(ecs.NewTransformComponent(position))

	// Выбираем случайную модель и текстуру куста
	renderComp := ecs.NewRenderComponent("bush", "bush_texture")
	bush.AddComponent(renderComp)

	// Добавляем физический компонент
	physicsComp := ecs.NewPhysicsComponent(10, true)               // Легкий неподвижный объект
	physicsComp.ColliderSize = ecs.Vector3{X: 0.8, Y: 0.8, Z: 0.8} // Размер коллайдера
	bush.AddComponent(physicsComp)

	// Добавляем метаморфный компонент со средней стабильностью
	metaComp := ecs.NewMetamorphicComponent(0.6) // Средне стабильны, могут легко меняться
	bush.AddComponent(metaComp)

	// Добавляем теги
	bush.AddTag("bush")
	bush.AddTag("vegetation")
	bush.AddTag("environment")

	// Случайно варьируем размер
	transformComp, _ := bush.GetComponent(ecs.TransformComponentID)
	transform := transformComp.(*ecs.TransformComponent)

	baseScale := 0.7 + randomFactor*0.6 // 0.7 - 1.3
	transform.Scale = ecs.Vector3{X: baseScale, Y: baseScale, Z: baseScale}

	// Добавляем сущность в мир
	world.AddEntity(bush)

	return bush
}

// createClearing создает поляну или особое место
func createClearing(world *ecs.World, position ecs.Vector3, randomFactor float64) *ecs.Entity {
	clearing := ecs.NewEntity()

	// Добавляем базовые компоненты
	clearing.AddComponent(ecs.NewTransformComponent(position))

	// Визуальный компонент (может быть невидимым или иметь особую текстуру)
	renderComp := ecs.NewRenderComponent("clearing", "clearing_texture")
	renderComp.Visible = false // Поляна сама по себе невидима, но влияет на окружение
	clearing.AddComponent(renderComp)

	// Добавляем интерактивный компонент
	interactComp := ecs.NewInteractableComponent("examine", "Исследовать поляну", 5.0)
	clearing.AddComponent(interactComp)

	// Добавляем метаморфный компонент с низкой стабильностью
	metaComp := ecs.NewMetamorphicComponent(0.3) // Нестабильны, легко меняются
	clearing.AddComponent(metaComp)

	// Добавляем теги
	clearing.AddTag("clearing")
	clearing.AddTag("ritual_site")
	clearing.AddTag("environment")
	clearing.AddTag("special")

	// Добавляем компонент света для создания особой атмосферы
	lightComp := ecs.NewLightComponent(color.RGBA{R: 200, G: 220, B: 255, A: 255}, 0.5, 10.0)
	lightComp.Flickering = true
	clearing.AddComponent(lightComp)

	// Добавляем сущность в мир
	world.AddEntity(clearing)

	return clearing
}

// createSymbol создает символ, который игрок может обнаружить
func createSymbol(world *ecs.World, position ecs.Vector3, randomFactor float64) *ecs.Entity {
	symbol := ecs.NewEntity()

	// Добавляем базовые компоненты
	symbol.AddComponent(ecs.NewTransformComponent(position))

	// Создаем символ с разными параметрами в зависимости от случайного фактора
	symbolType := "basic"
	if randomFactor > 0.7 {
		symbolType = "complex"
	}

	// Визуальный компонент
	renderComp := ecs.NewRenderComponent("symbol_"+symbolType, "symbol_"+symbolType+"_texture")
	renderComp.Effects = append(renderComp.Effects, "glow") // Символы слегка светятся
	symbol.AddComponent(renderComp)

	// Компонент символа
	symbolComp := ecs.NewSymbolComponent(generateSymbolID(randomFactor), symbolType, 0.3+randomFactor*0.7, 0.2+randomFactor*0.8)
	symbolComp.DiscoveryRadius = 3.0 // Радиус, в котором игрок может обнаружить символ
	symbol.AddComponent(symbolComp)

	// Добавляем интерактивный компонент
	interactComp := ecs.NewInteractableComponent("examine", "Изучить символ", 2.0)
	symbol.AddComponent(interactComp)

	// Добавляем теги
	symbol.AddTag("symbol")
	symbol.AddTag("interactive")
	symbol.AddTag("special")

	// Добавляем слабый источник света
	lightComp := ecs.NewLightComponent(color.RGBA{R: 100, G: 100, B: 255, A: 255}, 0.3, 5.0)
	lightComp.Flickering = true
	symbol.AddComponent(lightComp)

	// Добавляем сущность в мир
	world.AddEntity(symbol)

	return symbol
}

// createAnimal создает животное
func createAnimal(world *ecs.World, position ecs.Vector3, randomFactor float64) *ecs.Entity {
	animal := ecs.NewEntity()

	// Добавляем базовые компоненты
	animal.AddComponent(ecs.NewTransformComponent(position))

	// Определяем тип животного
	animalTypes := []string{"deer", "wolf", "rabbit", "fox"}
	animalType := animalTypes[int(randomFactor*float64(len(animalTypes)))]

	// Визуальный компонент
	renderComp := ecs.NewRenderComponent("animal_"+animalType, "animal_"+animalType+"_texture")
	animal.AddComponent(renderComp)

	// Физический компонент
	physicsComp := ecs.NewPhysicsComponent(50, false) // Подвижный объект
	animal.AddComponent(physicsComp)

	// Компонент здоровья
	healthComp := ecs.NewHealthComponent(100)
	animal.AddComponent(healthComp)

	// Компонент ИИ с разным поведением в зависимости от типа
	aiType := "neutral"
	detectionRange := 10.0

	if animalType == "wolf" {
		aiType = "aggressive"
		detectionRange = 15.0
	} else if animalType == "rabbit" {
		aiType = "scared"
		detectionRange = 12.0
	}

	aiComp := ecs.NewAIComponent(aiType, detectionRange)
	animal.AddComponent(aiComp)

	// Добавляем метаморфный компонент
	metaComp := ecs.NewMetamorphicComponent(0.5) // Средняя стабильность
	animal.AddComponent(metaComp)

	// Добавляем теги
	animal.AddTag("animal")
	animal.AddTag(animalType)
	animal.AddTag("living")

	// Если это хищник, добавляем тег
	if animalType == "wolf" || animalType == "fox" {
		animal.AddTag("predator")
	} else {
		animal.AddTag("prey")
	}

	// Добавляем сущность в мир
	world.AddEntity(animal)

	return animal
}

// createNightCreature создает ночное существо
func createNightCreature(world *ecs.World, position ecs.Vector3, anomalyLevel float64) *ecs.Entity {
	creature := ecs.NewEntity()

	// Добавляем базовые компоненты
	creature.AddComponent(ecs.NewTransformComponent(position))

	// Определяем тип существа в зависимости от уровня аномальности
	creatureType := "shadow"
	if anomalyLevel > 0.5 {
		creatureType = "wraith"
	}
	if anomalyLevel > 0.8 {
		creatureType = "nightmare"
	}

	// Визуальный компонент
	renderComp := ecs.NewRenderComponent("creature_"+creatureType, "creature_"+creatureType+"_texture")
	renderComp.Effects = append(renderComp.Effects, "glow", "transparency")
	creature.AddComponent(renderComp)

	// Физический компонент
	physicsComp := ecs.NewPhysicsComponent(0, false) // Бесплотное существо
	physicsComp.IsTrigger = true                     // Проходит сквозь физические объекты
	creature.AddComponent(physicsComp)

	// Компонент здоровья
	healthComp := ecs.NewHealthComponent(50 + int(anomalyLevel*100))
	creature.AddComponent(healthComp)

	// Компонент ИИ
	aiComp := ecs.NewAIComponent("aggressive", 20.0)
	aiComp.AttackDamage = 10 + anomalyLevel*20
	creature.AddComponent(aiComp)

	// Добавляем метаморфный компонент с низкой стабильностью
	metaComp := ecs.NewMetamorphicComponent(0.2) // Очень нестабильны
	metaComp.AbnormalityIndex = 0.7              // Уже довольно аномальны
	creature.AddComponent(metaComp)

	// Добавляем компонент звука
	soundComp := ecs.NewSoundEmitterComponent("creature_"+creatureType+"_ambient", 1.0, 15.0)
	soundComp.IsLooping = true
	soundComp.PlayOnStart = true
	creature.AddComponent(soundComp)

	// Добавляем теги
	creature.AddTag("creature")
	creature.AddTag("hostile")
	creature.AddTag("night")
	creature.AddTag(creatureType)

	// Добавляем компонент света для жутких эффектов
	lightComp := ecs.NewLightComponent(color.RGBA{R: 100, G: 20, B: 20, A: 255}, 0.3, 8.0)
	lightComp.Flickering = true
	creature.AddComponent(lightComp)

	// Добавляем сущность в мир
	world.AddEntity(creature)

	return creature
}

// createAnomaly создает аномалию
func createAnomaly(world *ecs.World, position ecs.Vector3, anomalyType string) *ecs.Entity {
	anomaly := ecs.NewEntity()

	// Добавляем базовые компоненты
	anomaly.AddComponent(ecs.NewTransformComponent(position))

	// Визуальный компонент
	renderComp := ecs.NewRenderComponent("anomaly_"+anomalyType, "anomaly_"+anomalyType+"_texture")
	renderComp.Effects = append(renderComp.Effects, "distortion")
	anomaly.AddComponent(renderComp)

	// Физический компонент (аномалии могут влиять на физику, но не имеют коллизий)
	physicsComp := ecs.NewPhysicsComponent(0, true)
	physicsComp.IsTrigger = true
	anomaly.AddComponent(physicsComp)

	// Добавляем интерактивный компонент
	interactComp := ecs.NewInteractableComponent("examine", "Исследовать аномалию", 3.0)
	anomaly.AddComponent(interactComp)

	// Добавляем метаморфный компонент с очень низкой стабильностью
	metaComp := ecs.NewMetamorphicComponent(0.1) // Крайне нестабильны
	metaComp.AbnormalityIndex = 0.9              // Очень аномальны
	anomaly.AddComponent(metaComp)

	// Добавляем компонент звука
	soundComp := ecs.NewSoundEmitterComponent("anomaly_"+anomalyType+"_ambient", 1.0, 10.0)
	soundComp.IsLooping = true
	soundComp.PlayOnStart = true
	anomaly.AddComponent(soundComp)

	// Добавляем теги
	anomaly.AddTag("anomaly")
	anomaly.AddTag("interactive")
	anomaly.AddTag("special")
	anomaly.AddTag(anomalyType)

	// Эффекты и параметры в зависимости от типа аномалии
	var radius float64
	var intensity float64

	switch anomalyType {
	case "minor":
		radius = 5.0
		intensity = 0.3
		renderComp.Effects = append(renderComp.Effects, "blur")
	case "medium":
		radius = 8.0
		intensity = 0.6
		renderComp.Effects = append(renderComp.Effects, "blur", "color_shift")
	case "major":
		radius = 12.0
		intensity = 0.9
		renderComp.Effects = append(renderComp.Effects, "blur", "color_shift", "warp")
	}

	// Добавляем компонент света
	lightComp := ecs.NewLightComponent(color.RGBA{R: 50, G: 150, B: 200, A: 255}, intensity, radius)
	lightComp.Flickering = true
	anomaly.AddComponent(lightComp)

	// Добавляем сущность в мир
	world.AddEntity(anomaly)

	return anomaly
}

// createDistortedEntity создает искаженную версию сущности
func createDistortedEntity(world *ecs.World, originalEntity *ecs.Entity) *ecs.Entity {
	distorted := ecs.NewEntity()

	// Копируем базовые компоненты с искажениями
	if originalEntity.HasComponent(ecs.TransformComponentID) {
		originalTransform, _ := originalEntity.GetComponent(ecs.TransformComponentID)
		transform := originalTransform.(*ecs.TransformComponent)

		// Создаем новый трансформ с искажениями
		newTransform := ecs.NewTransformComponent(transform.Position)

		// Искажаем масштаб и поворот
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		scaleDistortion := 0.5 + r.Float64()*1.5 // 0.5 - 2.0

		newTransform.Scale = transform.Scale.Multiply(scaleDistortion)
		newTransform.Rotation = transform.Rotation.Add(ecs.Vector3{
			X: r.Float64() * 0.5,
			Y: r.Float64() * 6.28,
			Z: r.Float64() * 0.5,
		})

		distorted.AddComponent(newTransform)
	}

	// Копируем и искажаем визуальный компонент
	if originalEntity.HasComponent(ecs.RenderComponentID) {
		originalRender, _ := originalEntity.GetComponent(ecs.RenderComponentID)
		render := originalRender.(*ecs.RenderComponent)

		// Создаем новый компонент рендера с искажениями
		newRender := ecs.NewRenderComponent(render.ModelID, render.TextureID)

		// Искажаем цвет
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		newRender.Color = color.RGBA{
			R: uint8(r.Intn(255)),
			G: uint8(r.Intn(255)),
			B: uint8(r.Intn(255)),
			A: 255,
		}

		// Добавляем эффекты искажения
		newRender.Effects = append(render.Effects, "distortion", "glow", "warp")
		newRender.Distortion = 0.7 + r.Float64()*0.3 // 0.7 - 1.0

		distorted.AddComponent(newRender)
	}

	// Копируем и модифицируем другие компоненты по необходимости
	// ...

	// Добавляем теги
	for _, tag := range originalEntity.GetTags() {
		distorted.AddTag(tag)
	}

	// Добавляем особые теги
	distorted.AddTag("distorted")
	distorted.AddTag("void")

	// Добавляем сущность в мир
	world.AddEntity(distorted)

	return distorted
}

// createVoidTerrain создает террейн типа "пустота"
func createVoidTerrain(x, y int) *terrain.TerrainData {
	// Создаем базовый террейн
	terrainData := terrain.NewTerrainData(ChunkSize, ChunkSize)

	// Искажаем высоты для создания странного ландшафта
	r := rand.New(rand.NewSource(int64(x*10000 + y)))

	for i := 0; i < ChunkSize; i++ {
		for j := 0; j < ChunkSize; j++ {
			// Создаем странные паттерны высот
			noise1 := math.Sin(float64(i)*0.1 + float64(j)*0.1)
			noise2 := math.Cos(float64(i)*0.05 - float64(j)*0.07)
			noise3 := math.Sin(math.Sqrt(float64(i*i+j*j)) * 0.1)

			height := noise1*5 + noise2*3 + noise3*8 + r.Float64()*2

			// Устанавливаем высоту
			terrainData.SetHeight(i, j, height)
		}
	}

	// Устанавливаем тип местности как "пустота"
	for i := 0; i < ChunkSize; i++ {
		for j := 0; j < ChunkSize; j++ {
			terrainData.SetGroundType(i, j, "void")
		}
	}

	return terrainData
}

// generateSymbolID генерирует ID для символа
func generateSymbolID(randomFactor float64) string {
	symbolTypes := []string{"elemental", "arcane", "primal", "void"}
	symbolType := symbolTypes[int(randomFactor*float64(len(symbolTypes)))]

	return symbolType + "_" + strconv.FormatInt(time.Now().UnixNano(), 16)
}

// containsString проверяет, содержится ли строка в слайсе
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
