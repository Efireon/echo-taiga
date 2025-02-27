package metamorphosis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"echo-taiga/internal/engine/ecs"
)

// OrderLevel определяет уровень (порядок) метаморфозы
const (
	OrderFirst  OrderLevel = 1 // Визуальные изменения (текстуры, цвета, звуки)
	OrderSecond OrderLevel = 2 // Структурные изменения (форма объектов, локальные аномалии)
	OrderThird  OrderLevel = 3 // Функциональные изменения (поведение объектов, новые свойства)
	OrderFourth OrderLevel = 4 // Системные изменения (физика, пространство, правила)
	OrderFifth  OrderLevel = 5 // Фундаментальные изменения (новые механики, изменение цели)
)

// MetamorphEffect представляет эффект метаморфозы
type MetamorphEffect struct {
	ID               string             // Уникальный идентификатор
	Name             string             // Название эффекта
	Description      string             // Описание эффекта
	Order            OrderLevel         // Порядок метаморфозы
	Category         string             // Категория эффекта (visual, physics, entity, etc.)
	AppliedTime      time.Time          // Время применения
	Duration         time.Duration      // Длительность (0 = постоянно)
	Intensity        float64            // Интенсивность эффекта (0-1)
	AffectedTags     []string           // Теги сущностей, на которые влияет
	AffectedArea     *AffectedArea      // Область воздействия
	ComponentChanges map[string]float64 // Изменения компонентов сущностей
	WorldChanges     map[string]float64 // Изменения правил мира
	VisualEffects    []string           // Визуальные эффекты
	SoundEffects     []string           // Звуковые эффекты
	RelatedSymbols   []string           // Связанные символы

	// Функции, выполняемые при применении/удалении эффекта
	OnApply  func(world *ecs.World, entity *ecs.Entity) error
	OnRemove func(world *ecs.World, entity *ecs.Entity) error
	OnUpdate func(world *ecs.World, entity *ecs.Entity, deltaTime float64) error
}

// AffectedArea определяет область воздействия метаморфозы
type AffectedArea struct {
	Type       string        // sphere, box, cylinder, path
	Center     ecs.Vector3   // Центр области
	Size       ecs.Vector3   // Размер области (для box)
	Radius     float64       // Радиус (для sphere, cylinder)
	Height     float64       // Высота (для cylinder)
	Points     []ecs.Vector3 // Точки пути (для path)
	Falloff    string        // none, linear, quadratic, exponential
	FalloffMin float64       // Минимальное расстояние для начала затухания
	FalloffMax float64       // Максимальное расстояние для полного затухания
}

// MetamorphTrigger определяет условие активации метаморфозы
type MetamorphTrigger struct {
	ID               string                 // Уникальный идентификатор
	Type             string                 // time, location, action, event, ritual, threshold
	Priority         float64                // Приоритет триггера (0-1)
	RequiredTags     []string               // Требуемые теги для активации
	ExcludedTags     []string               // Исключающие теги
	Location         *ecs.Vector3           // Локация для триггера типа location
	Radius           float64                // Радиус для триггера типа location
	ActionType       string                 // Тип действия для триггера типа action
	EventType        string                 // Тип события для триггера типа event
	RitualID         string                 // ID ритуала для триггера типа ritual
	ThresholdType    string                 // Тип порога для триггера типа threshold
	ThresholdValue   float64                // Значение порога
	ThresholdCompare string                 // Сравнение: greater, less, equal
	TimeOfDay        float64                // Время суток (0-1) для триггера типа time
	TimeTolerance    float64                // Допуск по времени
	Conditions       map[string]interface{} // Дополнительные условия

	// Функция проверки условия
	Check func(world *ecs.World, state *WorldState) bool
}

// WorldState содержит текущее состояние мира для проверки триггеров
type WorldState struct {
	TimeOfDay           float64                     // Время суток (0-1)
	PlayerPosition      ecs.Vector3                 // Позиция игрока
	PlayerHealth        float64                     // Здоровье игрока (0-1)
	PlayerSanity        float64                     // Рассудок игрока (0-1)
	TransformationPhase int                         // Текущая фаза трансформации мира
	ActiveEffects       map[string]*MetamorphEffect // Активные эффекты
	AnomalyLevel        float64                     // Общий уровень аномальности (0-1)
	RecentPlayerActions []PlayerAction              // Недавние действия игрока
	DiscoveredSymbols   []string                    // Открытые символы
	CompletedRituals    []string                    // Завершенные ритуалы
	LocalAnomalyLevels  map[string]float64          // Уровни аномальности по областям
	Weather             string                      // Текущая погода
	LastPlayerDeath     time.Time                   // Время последней смерти игрока
	Cycles              int                         // Количество циклов (перерождений)
}

// PlayerAction представляет действие игрока для анализа
type PlayerAction struct {
	Type      string                 // Тип действия
	Timestamp time.Time              // Время действия
	Position  ecs.Vector3            // Позиция
	Target    ecs.EntityID           // Цель действия (если есть)
	Value     float64                // Значение (если применимо)
	Tags      []string               // Теги действия
	Metadata  map[string]interface{} // Дополнительные данные
}

// MetamorphosisManager управляет метаморфозами в игре
type MetamorphosisManager struct {
	world             *ecs.World
	activeEffects     map[string]*MetamorphEffect
	availableTriggers map[string]*MetamorphTrigger
	effectTemplates   map[string]*MetamorphEffect
	triggerTemplates  map[string]*MetamorphTrigger

	// Бюджет аномалий (ограничивает количество одновременных эффектов)
	anomalyBudget    float64
	maxBudget        float64
	regenerationRate float64

	// Состояние мира для проверки триггеров
	worldState *WorldState

	// Фаза трансформации мира
	transformationPhase int

	// История изменений
	changeHistory []HistoryEntry

	// Ограничения по порядкам метаморфоз
	orderThresholds map[OrderLevel]float64

	// Зависимости между эффектами
	effectDependencies map[string][]string

	// Мьютекс для безопасного доступа
	mutex sync.RWMutex

	// Путь для сохранения/загрузки состояния
	savePath string
}

// HistoryEntry представляет запись в истории изменений
type HistoryEntry struct {
	Timestamp   time.Time
	EffectID    string
	Action      string // applied, removed
	EntityID    ecs.EntityID
	Description string
}

// NewMetamorphosisManager создает новый менеджер метаморфоз
func NewMetamorphosisManager(world *ecs.World, savePath string) *MetamorphosisManager {
	manager := &MetamorphosisManager{
		world:               world,
		activeEffects:       make(map[string]*MetamorphEffect),
		availableTriggers:   make(map[string]*MetamorphTrigger),
		effectTemplates:     make(map[string]*MetamorphEffect),
		triggerTemplates:    make(map[string]*MetamorphTrigger),
		anomalyBudget:       100.0,
		maxBudget:           100.0,
		regenerationRate:    0.5, // Единиц в минуту
		transformationPhase: 1,
		changeHistory:       make([]HistoryEntry, 0),
		orderThresholds: map[OrderLevel]float64{
			OrderFirst:  0.0,  // Доступны сразу
			OrderSecond: 0.25, // Требуется 25% прогресса
			OrderThird:  0.5,  // Требуется 50% прогресса
			OrderFourth: 0.75, // Требуется 75% прогресса
			OrderFifth:  0.9,  // Требуется 90% прогресса
		},
		effectDependencies: make(map[string][]string),
		savePath:           savePath,
		worldState: &WorldState{
			TimeOfDay:           0.25, // Начинаем с рассвета
			TransformationPhase: 1,
			ActiveEffects:       make(map[string]*MetamorphEffect),
			RecentPlayerActions: make([]PlayerAction, 0),
			DiscoveredSymbols:   make([]string, 0),
			CompletedRituals:    make([]string, 0),
			LocalAnomalyLevels:  make(map[string]float64),
			Weather:             "clear",
			Cycles:              0,
		},
	}

	// Регистрируем себя как систему в мире ECS
	world.AddSystem(manager)

	return manager
}

// Init инициализирует менеджер метаморфоз
func (mm *MetamorphosisManager) Init() error {
	// Загружаем шаблоны эффектов и триггеров
	err := mm.LoadEffectTemplates(filepath.Join(mm.savePath, "effect_templates"))
	if err != nil {
		return fmt.Errorf("failed to load effect templates: %v", err)
	}

	err = mm.LoadTriggerTemplates(filepath.Join(mm.savePath, "trigger_templates"))
	if err != nil {
		return fmt.Errorf("failed to load trigger templates: %v", err)
	}

	// Пытаемся загрузить текущее состояние, если оно есть
	err = mm.LoadState()
	if err != nil {
		// Если не удалось загрузить, создаем начальное состояние
		mm.InitializeDefaultState()
	}

	return nil
}

// LoadEffectTemplates загружает шаблоны эффектов из директории
func (mm *MetamorphosisManager) LoadEffectTemplates(dirPath string) error {
	// Проверяем существование директории
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Директория не существует, создаем ее
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}

		// Создаем базовые шаблоны эффектов
		mm.CreateDefaultEffectTemplates(dirPath)
	}

	// Загружаем файлы шаблонов
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// Загружаем шаблон эффекта
		filePath := filepath.Join(dirPath, file.Name())
		effectTemplate, err := mm.loadEffectTemplateFromFile(filePath)
		if err != nil {
			return err
		}

		// Регистрируем шаблон
		mm.effectTemplates[effectTemplate.ID] = effectTemplate
	}

	return nil
}

// LoadTriggerTemplates загружает шаблоны триггеров из директории
func (mm *MetamorphosisManager) LoadTriggerTemplates(dirPath string) error {
	// Проверяем существование директории
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Директория не существует, создаем ее
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}

		// Создаем базовые шаблоны триггеров
		mm.CreateDefaultTriggerTemplates(dirPath)
	}

	// Загружаем файлы шаблонов
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// Загружаем шаблон триггера
		filePath := filepath.Join(dirPath, file.Name())
		triggerTemplate, err := mm.loadTriggerTemplateFromFile(filePath)
		if err != nil {
			return err
		}

		// Регистрируем шаблон
		mm.triggerTemplates[triggerTemplate.ID] = triggerTemplate
	}

	return nil
}

// LoadState загружает текущее состояние менеджера метаморфоз
func (mm *MetamorphosisManager) LoadState() error {
	statePath := filepath.Join(mm.savePath, "metamorphosis_state.json")

	// Проверяем существование файла
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return fmt.Errorf("state file does not exist")
	}

	// Загружаем файл
	data, err := ioutil.ReadFile(statePath)
	if err != nil {
		return err
	}

	// Временная структура для хранения серилизуемых данных
	type SerializedState struct {
		AnomalyBudget       float64           `json:"anomaly_budget"`
		MaxBudget           float64           `json:"max_budget"`
		RegenerationRate    float64           `json:"regeneration_rate"`
		TransformationPhase int               `json:"transformation_phase"`
		ActiveEffects       map[string]string `json:"active_effects"` // ID -> Template ID
		WorldState          *WorldState       `json:"world_state"`
	}

	var state SerializedState
	err = json.Unmarshal(data, &state)
	if err != nil {
		return err
	}

	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Загружаем основные параметры
	mm.anomalyBudget = state.AnomalyBudget
	mm.maxBudget = state.MaxBudget
	mm.regenerationRate = state.RegenerationRate
	mm.transformationPhase = state.TransformationPhase
	mm.worldState = state.WorldState

	// Восстанавливаем активные эффекты из шаблонов
	mm.activeEffects = make(map[string]*MetamorphEffect)
	for id, templateID := range state.ActiveEffects {
		template, exists := mm.effectTemplates[templateID]
		if !exists {
			continue
		}

		// Создаем копию эффекта из шаблона
		effect := *template
		effect.ID = id

		mm.activeEffects[id] = &effect
	}

	return nil
}

// SaveState сохраняет текущее состояние менеджера метаморфоз
func (mm *MetamorphosisManager) SaveState() error {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	// Временная структура для хранения серилизуемых данных
	type SerializedState struct {
		AnomalyBudget       float64           `json:"anomaly_budget"`
		MaxBudget           float64           `json:"max_budget"`
		RegenerationRate    float64           `json:"regeneration_rate"`
		TransformationPhase int               `json:"transformation_phase"`
		ActiveEffects       map[string]string `json:"active_effects"` // ID -> Template ID
		WorldState          *WorldState       `json:"world_state"`
	}

	// Создаем серилизуемое представление
	state := SerializedState{
		AnomalyBudget:       mm.anomalyBudget,
		MaxBudget:           mm.maxBudget,
		RegenerationRate:    mm.regenerationRate,
		TransformationPhase: mm.transformationPhase,
		ActiveEffects:       make(map[string]string),
		WorldState:          mm.worldState,
	}

	// Сохраняем ID шаблонов активных эффектов
	for id, effect := range mm.activeEffects {
		// Находим шаблон по параметрам эффекта
		for templateID, template := range mm.effectTemplates {
			if effect.Name == template.Name && effect.Order == template.Order && effect.Category == template.Category {
				state.ActiveEffects[id] = templateID
				break
			}
		}
	}

	// Сериализуем данные
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	// Создаем директорию, если ее нет
	if _, err := os.Stat(mm.savePath); os.IsNotExist(err) {
		err = os.MkdirAll(mm.savePath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Сохраняем файл
	statePath := filepath.Join(mm.savePath, "metamorphosis_state.json")
	return ioutil.WriteFile(statePath, data, 0644)
}

// RequiredComponents возвращает компоненты, необходимые для работы системы
func (mm *MetamorphosisManager) RequiredComponents() []ecs.ComponentID {
	return []ecs.ComponentID{ecs.MetamorphicComponentID}
}

// Update обновляет состояние системы метаморфоз
func (mm *MetamorphosisManager) Update(deltaTime float64) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Обновляем бюджет аномалий
	mm.updateAnomalyBudget(deltaTime)

	// Обновляем состояние мира
	mm.updateWorldState()

	// Проверяем триггеры для новых метаморфоз
	mm.checkTriggers()

	// Обновляем активные эффекты
	mm.updateActiveEffects(deltaTime)

	// Применяем эффекты к сущностям
	mm.applyEffectsToEntities(deltaTime)
}

// updateAnomalyBudget обновляет бюджет аномалий
func (mm *MetamorphosisManager) updateAnomalyBudget(deltaTime float64) {
	// Преобразуем deltaTime в минуты
	deltaMinutes := deltaTime / 60.0

	// Регенерируем бюджет
	regenerationAmount := mm.regenerationRate * deltaMinutes
	mm.anomalyBudget = math.Min(mm.maxBudget, mm.anomalyBudget+regenerationAmount)
}

// updateWorldState обновляет состояние мира
func (mm *MetamorphosisManager) updateWorldState() {
	// Получаем время суток из мира (предполагаем, что есть система времени)
	// mm.worldState.TimeOfDay = ...

	// Получаем позицию игрока
	playerEntities := mm.world.GetEntitiesWithTag("player")
	if len(playerEntities) > 0 {
		player := playerEntities[0]

		// Получаем позицию
		if transformComp, has := player.GetComponent(ecs.TransformComponentID); has {
			transform := transformComp.(*ecs.TransformComponent)
			mm.worldState.PlayerPosition = transform.Position
		}

		// Получаем здоровье
		if healthComp, has := player.GetComponent(ecs.HealthComponentID); has {
			health := healthComp.(*ecs.HealthComponent)
			mm.worldState.PlayerHealth = health.CurrentHealth / health.MaxHealth
		}

		// Получаем рассудок
		if survivalComp, has := player.GetComponent(ecs.SurvivalComponentID); has {
			survival := survivalComp.(*ecs.SurvivalComponent)
			mm.worldState.PlayerSanity = survival.SanityLevel / 100.0
		}
	}

	// Обновляем активные эффекты в состоянии мира
	mm.worldState.ActiveEffects = mm.activeEffects

	// Обновляем фазу трансформации
	mm.worldState.TransformationPhase = mm.transformationPhase
}

// checkTriggers проверяет триггеры для активации новых метаморфоз
func (mm *MetamorphosisManager) checkTriggers() {
	// Сначала собираем все потенциальные триггеры
	var potentialTriggers []*MetamorphTrigger

	for _, trigger := range mm.availableTriggers {
		// Проверяем условие триггера
		if trigger.Check != nil && trigger.Check(mm.world, mm.worldState) {
			potentialTriggers = append(potentialTriggers, trigger)
		}
	}

	// Сортируем триггеры по приоритету
	sort.Slice(potentialTriggers, func(i, j int) bool {
		return potentialTriggers[i].Priority > potentialTriggers[j].Priority
	})

	// Активируем триггеры, начиная с самого приоритетного
	for _, trigger := range potentialTriggers {
		// Проверяем ограничения фазы трансформации
		effectOrder := getEffectOrderForTrigger(trigger)
		if !mm.isOrderAllowed(effectOrder) {
			continue
		}

		// Пытаемся активировать триггер
		effect := mm.selectEffectForTrigger(trigger)
		if effect != nil {
			// Проверяем бюджет
			if mm.canAffordEffect(effect) {
				// Активируем эффект
				mm.applyMetamorphEffect(effect)

				// Удаляем триггер из доступных
				delete(mm.availableTriggers, trigger.ID)

				// После активации одного эффекта прекращаем (чтобы не было слишком много изменений сразу)
				break
			}
		}
	}
}

// updateActiveEffects обновляет активные эффекты
func (mm *MetamorphosisManager) updateActiveEffects(deltaTime float64) {
	now := time.Now()

	// Проверяем все активные эффекты
	for id, effect := range mm.activeEffects {
		// Если эффект временный и его время истекло
		if effect.Duration > 0 && now.Sub(effect.AppliedTime) >= effect.Duration {
			// Удаляем эффект
			mm.removeMetamorphEffect(id)
			continue
		}

		// Обновляем эффект
		if effect.OnUpdate != nil {
			// Получаем все сущности, подходящие для эффекта
			entities := mm.getEntitiesForEffect(effect)

			for _, entity := range entities {
				err := effect.OnUpdate(mm.world, entity, deltaTime)
				if err != nil {
					// Логгируем ошибку
					fmt.Printf("Error updating effect %s for entity %s: %v\n", effect.ID, entity.ID, err)
				}
			}
		}
	}
}

// applyEffectsToEntities применяет эффекты к сущностям
func (mm *MetamorphosisManager) applyEffectsToEntities(deltaTime float64) {
	// Получаем все сущности с компонентом метаморфичности
	entities := mm.world.GetEntitiesWithComponent(ecs.MetamorphicComponentID)

	for _, entity := range entities {
		// Получаем компонент метаморфичности
		metamorphicComp, _ := entity.GetComponent(ecs.MetamorphicComponentID)
		metamorphic := metamorphicComp.(*ecs.MetamorphicComponent)

		// Проверяем стабильность сущности
		if metamorphic.Stability >= 1.0 {
			continue // Сущность полностью стабильна, пропускаем
		}

		// Получаем все эффекты, которые могут повлиять на эту сущность
		for _, effect := range mm.activeEffects {
			// Проверяем, подходит ли сущность для эффекта
			if !mm.isEntityAffectedByEffect(entity, effect) {
				continue
			}

			// Если сущность еще не подвержена этому эффекту, применяем его
			effectID := effect.ID
			if !containsString(metamorphic.CurrentMetamorphoses, effectID) {
				// Проверяем, может ли сущность мутировать под влиянием этого эффекта
				if metamorphic.CanMutate(effect.Intensity) {
					// Применяем эффект к сущности
					metamorphic.ApplyMetamorphosis(effectID, effect.Intensity)

					// Вызываем колбэк применения эффекта
					if effect.OnApply != nil {
						err := effect.OnApply(mm.world, entity)
						if err != nil {
							// Логгируем ошибку
							fmt.Printf("Error applying effect %s to entity %s: %v\n", effectID, entity.ID, err)
						}
					}

					// Добавляем запись в историю
					mm.recordHistoryEntry(effectID, "applied", entity.ID, fmt.Sprintf("Applied effect %s to entity %s", effect.Name, entity.ID))
				}
			}
		}

		// Проверяем, есть ли у сущности эффекты, которые больше не активны
		for _, effectID := range metamorphic.CurrentMetamorphoses {
			if _, exists := mm.activeEffects[effectID]; !exists {
				// Удаляем эффект из списка активных у сущности
				metamorphic.CurrentMetamorphoses = removeString(metamorphic.CurrentMetamorphoses, effectID)

				// Записываем в историю
				mm.recordHistoryEntry(effectID, "removed", entity.ID, fmt.Sprintf("Removed effect %s from entity %s", effectID, entity.ID))
			}
		}
	}
}

// ApplyMetamorphEffect применяет новый эффект метаморфозы
func (mm *MetamorphosisManager) applyMetamorphEffect(effect *MetamorphEffect) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Устанавливаем время применения
	effect.AppliedTime = time.Now()

	// Добавляем эффект в активные
	mm.activeEffects[effect.ID] = effect

	// Уменьшаем бюджет аномалий
	mm.anomalyBudget -= getEffectCost(effect)

	// Записываем в историю
	mm.recordHistoryEntry(effect.ID, "created", "", fmt.Sprintf("Created new effect: %s", effect.Name))

	// Применяем эффект ко всем подходящим сущностям
	entities := mm.getEntitiesForEffect(effect)
	for _, entity := range entities {
		metamorphicComp, has := entity.GetComponent(ecs.MetamorphicComponentID)
		if !has {
			continue
		}

		metamorphic := metamorphicComp.(*ecs.MetamorphicComponent)

		// Проверяем, может ли сущность мутировать
		if metamorphic.CanMutate(effect.Intensity) {
			// Применяем эффект
			metamorphic.ApplyMetamorphosis(effect.ID, effect.Intensity)

			// Вызываем колбэк
			if effect.OnApply != nil {
				err := effect.OnApply(mm.world, entity)
				if err != nil {
					// Логгируем ошибку
					fmt.Printf("Error applying effect %s to entity %s: %v\n", effect.ID, entity.ID, err)
				}
			}

			// Добавляем запись в историю
			mm.recordHistoryEntry(effect.ID, "applied", entity.ID, fmt.Sprintf("Applied effect %s to entity %s", effect.Name, entity.ID))
		}
	}

	// Проверяем, нужно ли увеличить фазу трансформации
	mm.checkTransformationPhaseProgress()
}

// RemoveMetamorphEffect удаляет эффект метаморфозы
func (mm *MetamorphosisManager) removeMetamorphEffect(effectID string) {
	effect, exists := mm.activeEffects[effectID]
	if !exists {
		return
	}

	// Находим все сущности, подверженные эффекту
	entities := mm.world.GetEntitiesWithComponent(ecs.MetamorphicComponentID)
	for _, entity := range entities {
		metamorphicComp, _ := entity.GetComponent(ecs.MetamorphicComponentID)
		metamorphic := metamorphicComp.(*ecs.MetamorphicComponent)

		// Проверяем, подвержена ли сущность этому эффекту
		if containsString(metamorphic.CurrentMetamorphoses, effectID) {
			// Удаляем эффект из списка активных у сущности
			metamorphic.CurrentMetamorphoses = removeString(metamorphic.CurrentMetamorphoses, effectID)

			// Вызываем колбэк удаления эффекта
			if effect.OnRemove != nil {
				err := effect.OnRemove(mm.world, entity)
				if err != nil {
					// Логгируем ошибку
					fmt.Printf("Error removing effect %s from entity %s: %v\n", effectID, entity.ID, err)
				}
			}

			// Добавляем запись в историю
			mm.recordHistoryEntry(effectID, "removed", entity.ID, fmt.Sprintf("Removed effect %s from entity %s", effect.Name, entity.ID))
		}
	}

	// Удаляем эффект из списка активных
	delete(mm.activeEffects, effectID)

	// Возвращаем часть бюджета аномалий
	mm.anomalyBudget = math.Min(mm.maxBudget, mm.anomalyBudget+getEffectCost(effect)*0.5)

	// Добавляем запись в историю
	mm.recordHistoryEntry(effectID, "removed", "", fmt.Sprintf("Removed effect: %s", effect.Name))
}

// GetActiveEffects возвращает список активных эффектов
func (mm *MetamorphosisManager) GetActiveEffects() []*MetamorphEffect {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	effects := make([]*MetamorphEffect, 0, len(mm.activeEffects))
	for _, effect := range mm.activeEffects {
		effects = append(effects, effect)
	}

	return effects
}

// GetEffectsByOrder возвращает список активных эффектов указанного порядка
func (mm *MetamorphosisManager) GetEffectsByOrder(order OrderLevel) []*MetamorphEffect {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	effects := make([]*MetamorphEffect, 0)
	for _, effect := range mm.activeEffects {
		if effect.Order == order {
			effects = append(effects, effect)
		}
	}

	return effects
}

// GetEffect возвращает эффект по ID
func (mm *MetamorphosisManager) GetEffect(effectID string) (*MetamorphEffect, bool) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	effect, exists := mm.activeEffects[effectID]
	return effect, exists
}

// AddPlayerAction добавляет действие игрока для анализа
func (mm *MetamorphosisManager) AddPlayerAction(action PlayerAction) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Добавляем действие в историю
	mm.worldState.RecentPlayerActions = append(mm.worldState.RecentPlayerActions, action)

	// Ограничиваем размер истории
	if len(mm.worldState.RecentPlayerActions) > 100 {
		mm.worldState.RecentPlayerActions = mm.worldState.RecentPlayerActions[1:]
	}
}

// AddDiscoveredSymbol добавляет открытый символ
func (mm *MetamorphosisManager) AddDiscoveredSymbol(symbolID string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Проверяем, не открыт ли уже символ
	if containsString(mm.worldState.DiscoveredSymbols, symbolID) {
		return
	}

	// Добавляем символ
	mm.worldState.DiscoveredSymbols = append(mm.worldState.DiscoveredSymbols, symbolID)

	// Проверяем прогресс фазы трансформации
	mm.checkTransformationPhaseProgress()
}

// AddCompletedRitual добавляет завершенный ритуал
func (mm *MetamorphosisManager) AddCompletedRitual(ritualID string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Проверяем, не завершен ли уже ритуал
	if containsString(mm.worldState.CompletedRituals, ritualID) {
		return
	}

	// Добавляем ритуал
	mm.worldState.CompletedRituals = append(mm.worldState.CompletedRituals, ritualID)

	// Проверяем прогресс фазы трансформации
	mm.checkTransformationPhaseProgress()
}

// SetLocalAnomalyLevel устанавливает уровень аномальности для области
func (mm *MetamorphosisManager) SetLocalAnomalyLevel(areaID string, level float64) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.worldState.LocalAnomalyLevels[areaID] = level

	// Обновляем общий уровень аномальности
	mm.updateGlobalAnomalyLevel()
}

// SetWeather устанавливает текущую погоду
func (mm *MetamorphosisManager) SetWeather(weather string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.worldState.Weather = weather
}

// RecordPlayerDeath записывает смерть игрока
func (mm *MetamorphosisManager) RecordPlayerDeath() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.worldState.LastPlayerDeath = time.Now()
	mm.worldState.Cycles++

	// Увеличиваем бюджет аномалий при перерождении
	mm.maxBudget += 25.0
	mm.anomalyBudget = mm.maxBudget

	// Проверяем прогресс фазы трансформации
	mm.checkTransformationPhaseProgress()
}

// GetTransformationPhase возвращает текущую фазу трансформации
func (mm *MetamorphosisManager) GetTransformationPhase() int {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	return mm.transformationPhase
}

// GetAnomalyBudget возвращает текущий бюджет аномалий
func (mm *MetamorphosisManager) GetAnomalyBudget() float64 {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	return mm.anomalyBudget
}

// GetMaxBudget возвращает максимальный бюджет аномалий
func (mm *MetamorphosisManager) GetMaxBudget() float64 {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	return mm.maxBudget
}

// GetRegenerationRate возвращает скорость регенерации бюджета аномалий
func (mm *MetamorphosisManager) GetRegenerationRate() float64 {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	return mm.regenerationRate
}

// SetTransformationPhase устанавливает фазу трансформации
func (mm *MetamorphosisManager) SetTransformationPhase(phase int) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.transformationPhase = phase
	mm.worldState.TransformationPhase = phase

	// Обновляем ограничения метаморфоз в зависимости от фазы
	if phase >= 2 {
		mm.orderThresholds[OrderSecond] = 0.0 // Сразу доступны
	}
	if phase >= 3 {
		mm.orderThresholds[OrderThird] = 0.0 // Сразу доступны
	}
	if phase >= 4 {
		mm.orderThresholds[OrderFourth] = 0.0 // Сразу доступны
	}
	if phase >= 5 {
		mm.orderThresholds[OrderFifth] = 0.0 // Сразу доступны
	}
}

// CreateEffectFromTemplate создает новый эффект на основе шаблона
func (mm *MetamorphosisManager) CreateEffectFromTemplate(templateID string) (*MetamorphEffect, error) {
	template, exists := mm.effectTemplates[templateID]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	// Создаем копию эффекта
	effect := *template

	// Генерируем уникальный ID
	effect.ID = fmt.Sprintf("%s_%s", templateID, generateUUID())

	// Сбрасываем время применения
	effect.AppliedTime = time.Time{}

	return &effect, nil
}

// Внутренние методы

// loadEffectTemplateFromFile загружает шаблон эффекта из файла
func (mm *MetamorphosisManager) loadEffectTemplateFromFile(filePath string) (*MetamorphEffect, error) {
	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Читаем содержимое файла
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Парсим JSON
	var effect MetamorphEffect
	err = json.Unmarshal(data, &effect)
	if err != nil {
		return nil, err
	}

	// Настраиваем колбэки в зависимости от типа эффекта
	mm.setupEffectCallbacks(&effect)

	return &effect, nil
}

// loadTriggerTemplateFromFile загружает шаблон триггера из файла
func (mm *MetamorphosisManager) loadTriggerTemplateFromFile(filePath string) (*MetamorphTrigger, error) {
	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Читаем содержимое файла
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Парсим JSON
	var trigger MetamorphTrigger
	err = json.Unmarshal(data, &trigger)
	if err != nil {
		return nil, err
	}

	// Настраиваем функцию проверки в зависимости от типа триггера
	mm.setupTriggerCheck(&trigger)

	return &trigger, nil
}

// CreateDefaultEffectTemplates создает базовые шаблоны эффектов
func (mm *MetamorphosisManager) CreateDefaultEffectTemplates(dirPath string) {
	// Создаем ряд базовых шаблонов эффектов разных порядков
	effects := []*MetamorphEffect{
		// Эффекты первого порядка (визуальные изменения)
		{
			ID:           "visual_distortion",
			Name:         "Visual Distortion",
			Description:  "Causes visual distortions and color shifts",
			Order:        OrderFirst,
			Category:     "visual",
			Duration:     10 * time.Minute,
			Intensity:    0.5,
			AffectedTags: []string{"visible"},
			ComponentChanges: map[string]float64{
				"render.distortion": 0.3,
			},
			VisualEffects: []string{"color_shift", "blur"},
		},
		{
			ID:           "eerie_sounds",
			Name:         "Eerie Sounds",
			Description:  "Causes strange and unsettling sounds",
			Order:        OrderFirst,
			Category:     "audio",
			Duration:     15 * time.Minute,
			Intensity:    0.4,
			AffectedTags: []string{"audio"},
			SoundEffects: []string{"whispers", "echoes"},
		},

		// Эффекты второго порядка (структурные изменения)
		{
			ID:           "twisted_vegetation",
			Name:         "Twisted Vegetation",
			Description:  "Causes plants to grow in twisted, unnatural ways",
			Order:        OrderSecond,
			Category:     "environment",
			Duration:     0, // Постоянный эффект
			Intensity:    0.6,
			AffectedTags: []string{"plant", "tree"},
			ComponentChanges: map[string]float64{
				"transform.scale.y": 1.5,
				"render.distortion": 0.4,
			},
			VisualEffects: []string{"twisted", "overgrown"},
		},
		{
			ID:          "gravity_anomaly",
			Name:        "Gravity Anomaly",
			Description: "Creates localized gravity disturbances",
			Order:       OrderSecond,
			Category:    "physics",
			Duration:    30 * time.Minute,
			Intensity:   0.7,
			AffectedArea: &AffectedArea{
				Type:       "sphere",
				Radius:     15.0,
				Falloff:    "quadratic",
				FalloffMin: 5.0,
				FalloffMax: 15.0,
			},
			ComponentChanges: map[string]float64{
				"physics.gravity": 0.5, // Уменьшенная гравитация
			},
		},

		// Эффекты третьего порядка (функциональные изменения)
		{
			ID:           "creature_mutation",
			Name:         "Creature Mutation",
			Description:  "Causes creatures to mutate and develop new abilities",
			Order:        OrderThird,
			Category:     "entity",
			Duration:     0, // Постоянный эффект
			Intensity:    0.8,
			AffectedTags: []string{"animal", "creature"},
			ComponentChanges: map[string]float64{
				"health.maxHealth":  1.5,  // Увеличенное здоровье
				"ai.detectionRange": 1.2,  // Улучшенное обнаружение
				"ai.aggressiveness": 1.3,  // Повышенная агрессивность
				"render.scale":      1.25, // Увеличенный размер
			},
			VisualEffects: []string{"mutation", "glow"},
		},
		{
			ID:          "time_distortion",
			Name:        "Time Distortion",
			Description: "Causes time to flow differently in an area",
			Order:       OrderThird,
			Category:    "reality",
			Duration:    45 * time.Minute,
			Intensity:   0.75,
			AffectedArea: &AffectedArea{
				Type:       "sphere",
				Radius:     20.0,
				Falloff:    "linear",
				FalloffMin: 10.0,
				FalloffMax: 20.0,
			},
			WorldChanges: map[string]float64{
				"time.flow": 0.5, // Замедление времени
			},
		},

		// Эффекты четвертого порядка (системные изменения)
		{
			ID:          "reality_tear",
			Name:        "Reality Tear",
			Description: "Creates a tear in reality, causing multiple anomalies",
			Order:       OrderFourth,
			Category:    "reality",
			Duration:    60 * time.Minute,
			Intensity:   0.9,
			AffectedArea: &AffectedArea{
				Type:       "sphere",
				Radius:     30.0,
				Falloff:    "exponential",
				FalloffMin: 10.0,
				FalloffMax: 30.0,
			},
			WorldChanges: map[string]float64{
				"physics.gravity":      0.3, // Сильно уменьшенная гравитация
				"physics.friction":     0.5, // Уменьшенное трение
				"rendering.distortion": 0.8, // Сильные визуальные искажения
				"audio.distortion":     0.7, // Искажение звука
			},
			VisualEffects: []string{"reality_tear", "void_particles", "color_inversion"},
			SoundEffects:  []string{"void_whispers", "reality_cracking"},
		},
		{
			ID:           "nightmare_manifestation",
			Name:         "Nightmare Manifestation",
			Description:  "Manifests player's fears as physical entities",
			Order:        OrderFourth,
			Category:     "entity",
			Duration:     40 * time.Minute,
			Intensity:    0.85,
			AffectedTags: []string{"player"},
			WorldChanges: map[string]float64{
				"ai.spawn.nightmare": 1.0, // Активация спавна кошмаров
			},
		},

		// Эффекты пятого порядка (фундаментальные изменения)
		{
			ID:          "reality_rewrite",
			Name:        "Reality Rewrite",
			Description: "Completely rewrites the rules of reality in an area",
			Order:       OrderFifth,
			Category:    "reality",
			Duration:    120 * time.Minute,
			Intensity:   1.0,
			AffectedArea: &AffectedArea{
				Type:    "sphere",
				Radius:  50.0,
				Falloff: "none", // Резкая граница
			},
			WorldChanges: map[string]float64{
				"physics.rules":       -1.0, // Инверсия физических правил
				"reality.consistency": -0.5, // Снижение "стабильности" реальности
				"mechanics.core":      0.5,  // Активация новых механик
			},
			VisualEffects: []string{"reality_inversion", "void_world"},
			SoundEffects:  []string{"reality_collapse", "void_chorus"},
		},
		{
			ID:          "metamorphic_awakening",
			Name:        "Metamorphic Awakening",
			Description: "Awakens metamorphic potential in all entities",
			Order:       OrderFifth,
			Category:    "entity",
			Duration:    0, // Постоянный эффект
			Intensity:   1.0,
			ComponentChanges: map[string]float64{
				"metamorphic.stability": -0.5, // Снижение стабильности всех сущностей
			},
			WorldChanges: map[string]float64{
				"metamorphosis.rate": 2.0, // Удвоенная скорость метаморфоз
				"anomaly.generation": 1.5, // Повышенная генерация аномалий
			},
		},
	}

	// Сохраняем шаблоны в файлы
	for _, effect := range effects {
		filename := filepath.Join(dirPath, effect.ID+".json")

		// Сериализуем эффект
		data, err := json.MarshalIndent(effect, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling effect %s: %v\n", effect.ID, err)
			continue
		}

		// Записываем в файл
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			fmt.Printf("Error writing effect template %s: %v\n", filename, err)
		}
	}
}

// CreateDefaultTriggerTemplates создает базовые шаблоны триггеров
func (mm *MetamorphosisManager) CreateDefaultTriggerTemplates(dirPath string) {
	// Создаем ряд базовых шаблонов триггеров
	triggers := []*MetamorphTrigger{
		// Триггеры по времени
		{
			ID:            "midnight_trigger",
			Type:          "time",
			Priority:      0.8,
			TimeOfDay:     0.0,  // Полночь
			TimeTolerance: 0.05, // ±5% от суток
			Conditions: map[string]interface{}{
				"min_phase": 1, // Минимальная фаза трансформации
			},
		},
		{
			ID:            "dawn_trigger",
			Type:          "time",
			Priority:      0.7,
			TimeOfDay:     0.25, // Рассвет
			TimeTolerance: 0.05,
			Conditions: map[string]interface{}{
				"min_phase": 1,
			},
		},

		// Триггеры по локации
		{
			ID:       "anomaly_zone_trigger",
			Type:     "location",
			Priority: 0.85,
			Radius:   10.0,
			Conditions: map[string]interface{}{
				"min_phase":    2,
				"requires_tag": "anomaly_zone",
			},
		},
		{
			ID:       "ritual_site_trigger",
			Type:     "location",
			Priority: 0.9,
			Radius:   5.0,
			Conditions: map[string]interface{}{
				"min_phase":    1,
				"requires_tag": "ritual_site",
			},
		},

		// Триггеры по действиям
		{
			ID:         "symbol_discovery_trigger",
			Type:       "action",
			Priority:   0.95,
			ActionType: "discover_symbol",
			Conditions: map[string]interface{}{
				"min_phase": 1,
			},
		},
		{
			ID:         "ritual_completion_trigger",
			Type:       "action",
			Priority:   1.0,
			ActionType: "complete_ritual",
			Conditions: map[string]interface{}{
				"min_phase": 2,
			},
		},

		// Триггеры по событиям
		{
			ID:        "player_death_trigger",
			Type:      "event",
			Priority:  0.9,
			EventType: "player_death",
			Conditions: map[string]interface{}{
				"min_phase": 1,
			},
		},
		{
			ID:        "blood_spill_trigger",
			Type:      "event",
			Priority:  0.75,
			EventType: "blood_spill",
			Conditions: map[string]interface{}{
				"min_phase":  2,
				"min_amount": 50.0,
			},
		},

		// Триггеры по порогам
		{
			ID:               "sanity_low_trigger",
			Type:             "threshold",
			Priority:         0.8,
			ThresholdType:    "player_sanity",
			ThresholdValue:   0.3,
			ThresholdCompare: "less",
			Conditions: map[string]interface{}{
				"min_phase": 2,
			},
		},
		{
			ID:               "anomaly_high_trigger",
			Type:             "threshold",
			Priority:         0.85,
			ThresholdType:    "anomaly_level",
			ThresholdValue:   0.7,
			ThresholdCompare: "greater",
			Conditions: map[string]interface{}{
				"min_phase": 3,
			},
		},
	}

	// Сохраняем шаблоны в файлы
	for _, trigger := range triggers {
		filename := filepath.Join(dirPath, trigger.ID+".json")

		// Сериализуем триггер
		data, err := json.MarshalIndent(trigger, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling trigger %s: %v\n", trigger.ID, err)
			continue
		}

		// Записываем в файл
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			fmt.Printf("Error writing trigger template %s: %v\n", filename, err)
		}
	}
}

// InitializeDefaultState инициализирует состояние по умолчанию
func (mm *MetamorphosisManager) InitializeDefaultState() {
	// Инициализируем мир с базовыми предустановками
	mm.transformationPhase = 1
	mm.anomalyBudget = 100.0
	mm.maxBudget = 100.0
	mm.regenerationRate = 0.5

	// Инициализируем состояние мира
	mm.worldState.TimeOfDay = 0.25 // Рассвет
	mm.worldState.TransformationPhase = 1
	mm.worldState.ActiveEffects = make(map[string]*MetamorphEffect)
	mm.worldState.RecentPlayerActions = make([]PlayerAction, 0)
	mm.worldState.DiscoveredSymbols = make([]string, 0)
	mm.worldState.CompletedRituals = make([]string, 0)
	mm.worldState.LocalAnomalyLevels = make(map[string]float64)
	mm.worldState.Weather = "clear"
	mm.worldState.Cycles = 0

	// Генерируем начальные триггеры
	mm.generateInitialTriggers()
}

// generateInitialTriggers генерирует начальные триггеры
func (mm *MetamorphosisManager) generateInitialTriggers() {
	// Загружаем несколько базовых триггеров из шаблонов
	for id, template := range mm.triggerTemplates {
		// Создаем копию триггера
		trigger := *template
		trigger.ID = fmt.Sprintf("%s_%s", id, generateUUID())

		// Добавляем триггер в доступные
		mm.availableTriggers[trigger.ID] = &trigger

		// Ограничиваем количество начальных триггеров
		if len(mm.availableTriggers) >= 5 {
			break
		}
	}
}

// setupEffectCallbacks настраивает колбэки для эффекта в зависимости от его типа
func (mm *MetamorphosisManager) setupEffectCallbacks(effect *MetamorphEffect) {
	// Настраиваем колбэки в зависимости от категории эффекта
	switch effect.Category {
	case "visual":
		// Настраиваем колбэки для визуальных эффектов
		effect.OnApply = func(world *ecs.World, entity *ecs.Entity) error {
			// Получаем компонент рендеринга
			renderComp, has := entity.GetComponent(ecs.RenderComponentID)
			if !has {
				return nil
			}

			render := renderComp.(*ecs.RenderComponent)

			// Применяем визуальные эффекты
			for _, visualEffect := range effect.VisualEffects {
				if !containsString(render.Effects, visualEffect) {
					render.Effects = append(render.Effects, visualEffect)
				}
			}

			// Применяем искажение
			if distortion, exists := effect.ComponentChanges["render.distortion"]; exists {
				render.Distortion = math.Min(1.0, render.Distortion+distortion)
			}

			return nil
		}

		effect.OnRemove = func(world *ecs.World, entity *ecs.Entity) error {
			// Получаем компонент рендеринга
			renderComp, has := entity.GetComponent(ecs.RenderComponentID)
			if !has {
				return nil
			}

			render := renderComp.(*ecs.RenderComponent)

			// Удаляем визуальные эффекты
			for _, visualEffect := range effect.VisualEffects {
				render.Effects = removeString(render.Effects, visualEffect)
			}

			// Уменьшаем искажение
			if distortion, exists := effect.ComponentChanges["render.distortion"]; exists {
				render.Distortion = math.Max(0.0, render.Distortion-distortion)
			}

			return nil
		}

	case "physics":
		// Настраиваем колбэки для физических эффектов
		effect.OnApply = func(world *ecs.World, entity *ecs.Entity) error {
			// Получаем компонент физики
			physicsComp, has := entity.GetComponent(ecs.PhysicsComponentID)
			if !has {
				return nil
			}

			physics := physicsComp.(*ecs.PhysicsComponent)

			// Применяем изменения компонентов
			if gravity, exists := effect.ComponentChanges["physics.gravity"]; exists {
				physics.Gravity = gravity
			}

			if friction, exists := effect.ComponentChanges["physics.friction"]; exists {
				physics.Friction = friction
			}

			return nil
		}

		effect.OnRemove = func(world *ecs.World, entity *ecs.Entity) error {
			// Получаем компонент физики
			physicsComp, has := entity.GetComponent(ecs.PhysicsComponentID)
			if !has {
				return nil
			}

			physics := physicsComp.(*ecs.PhysicsComponent)

			// Возвращаем нормальные значения
			physics.Gravity = 1.0
			physics.Friction = 0.5

			return nil
		}

	case "entity":
		// Настраиваем колбэки для эффектов сущностей
		effect.OnApply = func(world *ecs.World, entity *ecs.Entity) error {
			// Применяем изменения в зависимости от типа компонента

			// Изменения здоровья
			if healthComp, has := entity.GetComponent(ecs.HealthComponentID); has {
				health := healthComp.(*ecs.HealthComponent)

				if maxHealth, exists := effect.ComponentChanges["health.maxHealth"]; exists {
					// Запоминаем текущий процент здоровья
					healthPercent := health.CurrentHealth / health.MaxHealth

					// Изменяем максимальное здоровье
					health.MaxHealth *= maxHealth

					// Обновляем текущее здоровье, сохраняя процентное соотношение
					health.CurrentHealth = health.MaxHealth * healthPercent
				}
			}

			// Изменения ИИ
			if aiComp, has := entity.GetComponent(ecs.AIComponentID); has {
				ai := aiComp.(*ecs.AIComponent)

				if detectionRange, exists := effect.ComponentChanges["ai.detectionRange"]; exists {
					ai.DetectionRange *= detectionRange
				}

				if attackDamage, exists := effect.ComponentChanges["ai.attackDamage"]; exists {
					ai.AttackDamage *= attackDamage
				}

				if aggressiveness, exists := effect.ComponentChanges["ai.aggressiveness"]; exists {
					// Устанавливаем более агрессивное поведение
					if aggressiveness > 1.0 && ai.AIType == "neutral" {
						ai.AIType = "aggressive"
					}
				}
			}

			// Изменения трансформации
			if transformComp, has := entity.GetComponent(ecs.TransformComponentID); has {
				transform := transformComp.(*ecs.TransformComponent)

				if scale, exists := effect.ComponentChanges["render.scale"]; exists {
					transform.Scale = transform.Scale.Multiply(scale)
				}
			}

			return nil
		}

	case "audio":
		// Настраиваем колбэки для звуковых эффектов
		effect.OnApply = func(world *ecs.World, entity *ecs.Entity) error {
			// Получаем компонент звука
			soundComp, has := entity.GetComponent(ecs.SoundEmitterComponentID)
			if !has {
				return nil
			}

			sound := soundComp.(*ecs.SoundEmitterComponent)

			// Добавляем звуковые эффекты
			for _, soundEffect := range effect.SoundEffects {
				if soundID, exists := sound.Sounds[soundEffect]; exists {
					sound.SoundID = soundID
					sound.Play()
				}
			}

			return nil
		}

	case "reality":
		// Для эффектов реальности не задаем колбэки на уровне сущностей,
		// они будут обрабатываться глобально
		break
	}
}

// setupTriggerCheck настраивает функцию проверки для триггера
func (mm *MetamorphosisManager) setupTriggerCheck(trigger *MetamorphTrigger) {
	// Настраиваем функцию проверки в зависимости от типа триггера
	switch trigger.Type {
	case "time":
		trigger.Check = func(world *ecs.World, state *WorldState) bool {
			// Проверяем, соответствует ли текущее время суток
			timeDiff := math.Abs(state.TimeOfDay - trigger.TimeOfDay)
			if timeDiff > 0.5 {
				timeDiff = 1.0 - timeDiff // Обрабатываем случай, когда разница переходит через полночь
			}

			if timeDiff <= trigger.TimeTolerance {
				// Проверяем дополнительные условия
				if minPhase, ok := trigger.Conditions["min_phase"].(float64); ok {
					return state.TransformationPhase >= int(minPhase)
				}
				return true
			}

			return false
		}

	case "location":
		trigger.Check = func(world *ecs.World, state *WorldState) bool {
			// Если задана конкретная локация
			if trigger.Location != nil {
				// Проверяем, находится ли игрок в указанном радиусе
				distance := state.PlayerPosition.Distance(*trigger.Location)
				if distance <= trigger.Radius {
					// Проверяем дополнительные условия
					if minPhase, ok := trigger.Conditions["min_phase"].(float64); ok {
						return state.TransformationPhase >= int(minPhase)
					}
					return true
				}
			} else if requiresTag, ok := trigger.Conditions["requires_tag"].(string); ok {
				// Проверяем, находится ли игрок рядом с объектом нужного типа
				entities := world.GetEntitiesWithTag(requiresTag)

				for _, entity := range entities {
					transformComp, has := entity.GetComponent(ecs.TransformComponentID)
					if !has {
						continue
					}

					transform := transformComp.(*ecs.TransformComponent)
					distance := state.PlayerPosition.Distance(transform.Position)

					if distance <= trigger.Radius {
						if minPhase, ok := trigger.Conditions["min_phase"].(float64); ok {
							return state.TransformationPhase >= int(minPhase)
						}
						return true
					}
				}
			}

			return false
		}

	case "action":
		trigger.Check = func(world *ecs.World, state *WorldState) bool {
			// Проверяем, выполнил ли игрок нужное действие недавно
			for _, action := range state.RecentPlayerActions {
				if action.Type == trigger.ActionType {
					// Проверяем, было ли действие выполнено недавно (в течение последних 5 минут)
					if time.Since(action.Timestamp) <= 5*time.Minute {
						// Проверяем дополнительные условия
						if minPhase, ok := trigger.Conditions["min_phase"].(float64); ok {
							return state.TransformationPhase >= int(minPhase)
						}
						return true
					}
				}
			}

			return false
		}

	case "event":
		trigger.Check = func(world *ecs.World, state *WorldState) bool {
			// Проверяем события в зависимости от их типа
			switch trigger.EventType {
			case "player_death":
				// Проверяем, умирал ли игрок недавно
				if !state.LastPlayerDeath.IsZero() && time.Since(state.LastPlayerDeath) <= 10*time.Minute {
					if minPhase, ok := trigger.Conditions["min_phase"].(float64); ok {
						return state.TransformationPhase >= int(minPhase)
					}
					return true
				}

			case "blood_spill":
				// Ищем действие кровопролития среди недавних действий
				for _, action := range state.RecentPlayerActions {
					if action.Type == "blood_spill" {
						// Проверяем, было ли действие выполнено недавно
						if time.Since(action.Timestamp) <= 5*time.Minute {
							// Проверяем количество крови
							if minAmount, ok := trigger.Conditions["min_amount"].(float64); ok {
								return action.Value >= minAmount
							}
							return true
						}
					}
				}
			}

			return false
		}

	case "threshold":
		trigger.Check = func(world *ecs.World, state *WorldState) bool {
			// Проверяем различные пороговые значения
			var value float64

			switch trigger.ThresholdType {
			case "player_sanity":
				value = state.PlayerSanity
			case "player_health":
				value = state.PlayerHealth
			case "anomaly_level":
				value = state.AnomalyLevel
			default:
				return false
			}

			// Сравниваем значение с порогом
			switch trigger.ThresholdCompare {
			case "greater":
				if value > trigger.ThresholdValue {
					if minPhase, ok := trigger.Conditions["min_phase"].(float64); ok {
						return state.TransformationPhase >= int(minPhase)
					}
					return true
				}
			case "less":
				if value < trigger.ThresholdValue {
					if minPhase, ok := trigger.Conditions["min_phase"].(float64); ok {
						return state.TransformationPhase >= int(minPhase)
					}
					return true
				}
			case "equal":
				if math.Abs(value-trigger.ThresholdValue) < 0.01 {
					if minPhase, ok := trigger.Conditions["min_phase"].(float64); ok {
						return state.TransformationPhase >= int(minPhase)
					}
					return true
				}
			}

			return false
		}

	case "ritual":
		trigger.Check = func(world *ecs.World, state *WorldState) bool {
			// Проверяем, был ли выполнен указанный ритуал
			for _, ritualID := range state.CompletedRituals {
				if ritualID == trigger.RitualID {
					// Проверяем, был ли ритуал выполнен недавно
					for _, action := range state.RecentPlayerActions {
						if action.Type == "complete_ritual" && action.Target == ecs.EntityID(ritualID) {
							if time.Since(action.Timestamp) <= 10*time.Minute {
								if minPhase, ok := trigger.Conditions["min_phase"].(float64); ok {
									return state.TransformationPhase >= int(minPhase)
								}
								return true
							}
						}
					}
				}
			}

			return false
		}
	}
}

// isEntityAffectedByEffect проверяет, подходит ли сущность для эффекта
func (mm *MetamorphosisManager) isEntityAffectedByEffect(entity *ecs.Entity, effect *MetamorphEffect) bool {
	// Проверяем теги
	if len(effect.AffectedTags) > 0 {
		hasTag := false
		for _, tag := range effect.AffectedTags {
			if entity.HasTag(tag) {
				hasTag = true
				break
			}
		}

		if !hasTag {
			return false
		}
	}

	// Если эффект имеет область действия, проверяем, находится ли сущность в ней
	if effect.AffectedArea != nil {
		transformComp, has := entity.GetComponent(ecs.TransformComponentID)
		if !has {
			return false
		}

		transform := transformComp.(*ecs.TransformComponent)

		// Проверяем тип области
		switch effect.AffectedArea.Type {
		case "sphere":
			// Вычисляем расстояние от центра области до сущности
			distance := transform.Position.Distance(effect.AffectedArea.Center)

			// Проверяем, находится ли сущность в радиусе действия
			if distance > effect.AffectedArea.Radius {
				return false
			}

			// Учитываем затухание эффекта с расстоянием
			if effect.AffectedArea.Falloff != "none" {
				if distance > effect.AffectedArea.FalloffMin {
					// Нормализуем расстояние от 0 до 1
					normalizedDistance := (distance - effect.AffectedArea.FalloffMin) / (effect.AffectedArea.FalloffMax - effect.AffectedArea.FalloffMin)
					normalizedDistance = math.Min(1.0, math.Max(0.0, normalizedDistance))

					// Вычисляем затухание в зависимости от типа
					var falloff float64
					switch effect.AffectedArea.Falloff {
					case "linear":
						falloff = 1.0 - normalizedDistance
					case "quadratic":
						falloff = 1.0 - normalizedDistance*normalizedDistance
					case "exponential":
						falloff = math.Exp(-3.0 * normalizedDistance)
					default:
						falloff = 1.0 - normalizedDistance
					}

					// Если затухание слишком сильное, считаем, что сущность не подвержена эффекту
					if falloff < 0.05 {
						return false
					}
				}
			}

		case "box":
			// Проверяем, находится ли сущность внутри бокса
			position := transform.Position
			center := effect.AffectedArea.Center
			size := effect.AffectedArea.Size

			if math.Abs(position.X-center.X) > size.X/2 ||
				math.Abs(position.Y-center.Y) > size.Y/2 ||
				math.Abs(position.Z-center.Z) > size.Z/2 {
				return false
			}

		case "cylinder":
			// Вычисляем расстояние в горизонтальной плоскости
			position := transform.Position
			center := effect.AffectedArea.Center

			horizontalDistance := math.Sqrt(
				math.Pow(position.X-center.X, 2) +
					math.Pow(position.Z-center.Z, 2))

			// Проверяем радиус и высоту цилиндра
			if horizontalDistance > effect.AffectedArea.Radius ||
				math.Abs(position.Y-center.Y) > effect.AffectedArea.Height/2 {
				return false
			}
		}
	}

	return true
}

// getEntitiesForEffect возвращает все сущности, подходящие для эффекта
func (mm *MetamorphosisManager) getEntitiesForEffect(effect *MetamorphEffect) []*ecs.Entity {
	// Если нет тегов и области, возвращаем все сущности с компонентом метаморфичности
	if len(effect.AffectedTags) == 0 && effect.AffectedArea == nil {
		return mm.world.GetEntitiesWithComponent(ecs.MetamorphicComponentID)
	}

	// Если есть теги, но нет области, получаем сущности по тегам
	if len(effect.AffectedTags) > 0 && effect.AffectedArea == nil {
		var entities []*ecs.Entity

		for _, tag := range effect.AffectedTags {
			tagEntities := mm.world.GetEntitiesWithTag(tag)
			for _, entity := range tagEntities {
				if entity.HasComponent(ecs.MetamorphicComponentID) {
					entities = append(entities, entity)
				}
			}
		}

		return entities
	}

	// Если есть область, получаем все сущности и фильтруем по положению
	entities := mm.world.GetEntitiesWithComponent(ecs.MetamorphicComponentID)
	var result []*ecs.Entity

	for _, entity := range entities {
		if mm.isEntityAffectedByEffect(entity, effect) {
			result = append(result, entity)
		}
	}

	return result
}

// selectEffectForTrigger выбирает подходящий эффект для триггера
func (mm *MetamorphosisManager) selectEffectForTrigger(trigger *MetamorphTrigger) *MetamorphEffect {
	// Определяем порядок эффекта на основе фазы трансформации
	order := getEffectOrderForTrigger(trigger)

	// Выбираем подходящие шаблоны эффектов
	var suitableTemplates []*MetamorphEffect

	for _, template := range mm.effectTemplates {
		if template.Order == order {
			suitableTemplates = append(suitableTemplates, template)
		}
	}

	if len(suitableTemplates) == 0 {
		return nil
	}

	// Выбираем случайный шаблон
	template := suitableTemplates[rand.Intn(len(suitableTemplates))]

	// Создаем копию эффекта
	effect := *template
	effect.ID = fmt.Sprintf("%s_%s", template.ID, generateUUID())

	return &effect
}

// getEffectOrderForTrigger определяет порядок эффекта для триггера
func getEffectOrderForTrigger(trigger *MetamorphTrigger) OrderLevel {
	// Получаем минимальную фазу трансформации из условий
	minPhase := 1
	if phase, ok := trigger.Conditions["min_phase"].(float64); ok {
		minPhase = int(phase)
	}

	// Определяем порядок эффекта в зависимости от фазы и приоритета
	switch {
	case minPhase >= 5 || trigger.Priority >= 0.95:
		return OrderFifth
	case minPhase >= 4 || trigger.Priority >= 0.85:
		return OrderFourth
	case minPhase >= 3 || trigger.Priority >= 0.75:
		return OrderThird
	case minPhase >= 2 || trigger.Priority >= 0.6:
		return OrderSecond
	default:
		return OrderFirst
	}
}

// getEffectCost возвращает стоимость эффекта в бюджете аномалий
func getEffectCost(effect *MetamorphEffect) float64 {
	// Базовая стоимость зависит от порядка эффекта
	baseCost := float64(effect.Order) * 10.0

	// Умножаем на интенсивность
	cost := baseCost * effect.Intensity

	// Учитываем длительность (постоянные эффекты дороже)
	if effect.Duration == 0 {
		cost *= 1.5
	} else {
		// Чем дольше эффект, тем он дороже
		durationHours := effect.Duration.Hours()
		cost *= 1.0 + durationHours/10.0
	}

	return cost
}

// canAffordEffect проверяет, хватает ли бюджета аномалий для эффекта
func (mm *MetamorphosisManager) canAffordEffect(effect *MetamorphEffect) bool {
	cost := getEffectCost(effect)
	return mm.anomalyBudget >= cost
}

// isOrderAllowed проверяет, разрешен ли указанный порядок метаморфоз
func (mm *MetamorphosisManager) isOrderAllowed(order OrderLevel) bool {
	// Получаем порог для указанного порядка
	threshold, exists := mm.orderThresholds[order]
	if !exists {
		return false
	}

	// Вычисляем прогресс трансформации
	progress := mm.getTransformationProgress()

	// Проверяем, достигнут ли порог
	return progress >= threshold
}

// getTransformationProgress возвращает прогресс трансформации (0-1)
func (mm *MetamorphosisManager) getTransformationProgress() float64 {
	// Вычисляем прогресс на основе нескольких факторов

	// Прогресс на основе открытых символов (вес: 0.3)
	symbolProgress := 0.0
	if len(mm.worldState.DiscoveredSymbols) > 0 {
		// Предполагаем, что всего 20 символов для полного прогресса
		symbolProgress = math.Min(1.0, float64(len(mm.worldState.DiscoveredSymbols))/20.0)
	}

	// Прогресс на основе выполненных ритуалов (вес: 0.3)
	ritualProgress := 0.0
	if len(mm.worldState.CompletedRituals) > 0 {
		// Предполагаем, что всего 10 ритуалов для полного прогресса
		ritualProgress = math.Min(1.0, float64(len(mm.worldState.CompletedRituals))/10.0)
	}

	// Прогресс на основе циклов перерождения (вес: 0.2)
	cycleProgress := 0.0
	if mm.worldState.Cycles > 0 {
		// Предполагаем, что 5 циклов для полного прогресса
		cycleProgress = math.Min(1.0, float64(mm.worldState.Cycles)/5.0)
	}

	// Прогресс на основе общего уровня аномальности (вес: 0.2)
	anomalyProgress := mm.worldState.AnomalyLevel

	// Вычисляем взвешенный прогресс
	progress := symbolProgress*0.3 + ritualProgress*0.3 + cycleProgress*0.2 + anomalyProgress*0.2

	return progress
}

// checkTransformationPhaseProgress проверяет и обновляет фазу трансформации
func (mm *MetamorphosisManager) checkTransformationPhaseProgress() {
	// Получаем текущий прогресс
	progress := mm.getTransformationProgress()

	// Определяем, нужно ли увеличить фазу
	newPhase := mm.transformationPhase

	switch {
	case progress >= 0.9 && mm.transformationPhase < 5:
		newPhase = 5
	case progress >= 0.7 && mm.transformationPhase < 4:
		newPhase = 4
	case progress >= 0.5 && mm.transformationPhase < 3:
		newPhase = 3
	case progress >= 0.25 && mm.transformationPhase < 2:
		newPhase = 2
	}

	// Если нужно увеличить фазу, делаем это
	if newPhase > mm.transformationPhase {
		mm.SetTransformationPhase(newPhase)

		// Записываем в историю
		mm.recordHistoryEntry("", "phase_change", "", fmt.Sprintf("Advanced to transformation phase %d", newPhase))
	}
}

// updateGlobalAnomalyLevel обновляет общий уровень аномальности
func (mm *MetamorphosisManager) updateGlobalAnomalyLevel() {
	// Вычисляем средний уровень аномальности по всем областям
	totalLevel := 0.0
	count := 0

	for _, level := range mm.worldState.LocalAnomalyLevels {
		totalLevel += level
		count++
	}

	// Добавляем вклад от активных эффектов
	for _, effect := range mm.activeEffects {
		totalLevel += effect.Intensity * 0.2
		count++
	}

	// Вычисляем средний уровень
	if count > 0 {
		mm.worldState.AnomalyLevel = math.Min(1.0, totalLevel/float64(count))
	} else {
		mm.worldState.AnomalyLevel = 0.0
	}
}

// recordHistoryEntry записывает событие в историю
func (mm *MetamorphosisManager) recordHistoryEntry(effectID string, action string, entityID ecs.EntityID, description string) {
	entry := HistoryEntry{
		Timestamp:   time.Now(),
		EffectID:    effectID,
		Action:      action,
		EntityID:    entityID,
		Description: description,
	}

	mm.changeHistory = append(mm.changeHistory, entry)

	// Ограничиваем размер истории
	if len(mm.changeHistory) > 1000 {
		mm.changeHistory = mm.changeHistory[len(mm.changeHistory)-1000:]
	}
}

// generateUUID генерирует уникальный идентификатор
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Вспомогательные функции для работы со строками

// containsString проверяет, содержится ли строка в слайсе
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// removeString удаляет строку из слайса
func removeString(slice []string, str string) []string {
	var result []string

	for _, s := range slice {
		if s != str {
			result = append(result, s)
		}
	}

	return result
}
