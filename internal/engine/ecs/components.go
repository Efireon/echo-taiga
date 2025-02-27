package ecs

import (
	"image/color"
	"math"
	"math/rand/v2"
)

// Предопределенные типы компонентов
var (
	TransformComponentID     = RegisterComponentType("transform")
	RenderComponentID        = RegisterComponentType("render")
	PhysicsComponentID       = RegisterComponentType("physics")
	HealthComponentID        = RegisterComponentType("health")
	AIComponentID            = RegisterComponentType("ai")
	PlayerControlComponentID = RegisterComponentType("player_control")
	MetamorphicComponentID   = RegisterComponentType("metamorphic")
	InteractableComponentID  = RegisterComponentType("interactable")
	SymbolComponentID        = RegisterComponentType("symbol")
	LightComponentID         = RegisterComponentType("light")
	SoundEmitterComponentID  = RegisterComponentType("sound_emitter")
	InventoryComponentID     = RegisterComponentType("inventory")
	SurvivalComponentID      = RegisterComponentType("survival")
)

// Vector3 представляет трехмерный вектор
type Vector3 struct {
	X, Y, Z float64
}

// NewVector3 создает новый трехмерный вектор
func NewVector3(x, y, z float64) Vector3 {
	return Vector3{X: x, Y: y, Z: z}
}

// Add складывает два вектора
func (v Vector3) Add(other Vector3) Vector3 {
	return Vector3{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

// Sub вычитает вектор из другого вектора
func (v Vector3) Sub(other Vector3) Vector3 {
	return Vector3{
		X: v.X - other.X,
		Y: v.Y - other.Y,
		Z: v.Z - other.Z,
	}
}

// Multiply умножает вектор на скаляр
func (v Vector3) Multiply(scalar float64) Vector3 {
	return Vector3{
		X: v.X * scalar,
		Y: v.Y * scalar,
		Z: v.Z * scalar,
	}
}

// Magnitude возвращает длину вектора
func (v Vector3) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Normalize возвращает нормализованный вектор
func (v Vector3) Normalize() Vector3 {
	mag := v.Magnitude()
	if mag == 0 {
		return v
	}
	return Vector3{
		X: v.X / mag,
		Y: v.Y / mag,
		Z: v.Z / mag,
	}
}

// Distance возвращает расстояние между двумя векторами
func (v Vector3) Distance(other Vector3) float64 {
	return v.Sub(other).Magnitude()
}

// Dot возвращает скалярное произведение двух векторов
func (v Vector3) Dot(other Vector3) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Cross возвращает векторное произведение двух векторов
func (v Vector3) Cross(other Vector3) Vector3 {
	return Vector3{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

// TransformComponent содержит информацию о позиции, вращении и масштабе сущности
type TransformComponent struct {
	BaseComponent
	Position Vector3
	Rotation Vector3 // Углы Эйлера в радианах (yaw, pitch, roll)
	Scale    Vector3
	Parent   EntityID
	Children []EntityID
}

// NewTransformComponent создает новый компонент трансформации
func NewTransformComponent(position Vector3) *TransformComponent {
	return &TransformComponent{
		BaseComponent: NewBaseComponent(TransformComponentID),
		Position:      position,
		Rotation:      Vector3{0, 0, 0},
		Scale:         Vector3{1, 1, 1},
		Children:      make([]EntityID, 0),
	}
}

// Forward возвращает вектор направления "вперед" в локальной системе координат
func (t *TransformComponent) Forward() Vector3 {
	// Преобразуем углы Эйлера в вектор направления
	yaw := t.Rotation.Y
	pitch := t.Rotation.X

	// Вычисляем вектор направления
	x := math.Sin(yaw) * math.Cos(pitch)
	y := math.Sin(pitch)
	z := math.Cos(yaw) * math.Cos(pitch)

	return Vector3{X: x, Y: y, Z: z}.Normalize()
}

// Right возвращает вектор направления "вправо" в локальной системе координат
func (t *TransformComponent) Right() Vector3 {
	forward := t.Forward()
	up := Vector3{0, 1, 0}

	return forward.Cross(up).Normalize()
}

// Up возвращает вектор направления "вверх" в локальной системе координат
func (t *TransformComponent) Up() Vector3 {
	forward := t.Forward()
	right := t.Right()

	return right.Cross(forward).Normalize()
}

// RenderComponent содержит информацию для отрисовки сущности
type RenderComponent struct {
	BaseComponent
	ModelID      string
	TextureID    string
	Color        color.RGBA
	Visible      bool
	CastShadow   bool
	Layer        int
	Distortion   float64  // Для метаморфоз: 0 - нет искажений, 1 - максимальные искажения
	Effects      []string // Список применяемых эффектов (свечение, размытие и т.д.)
	Pixel        bool     // Использовать пиксельный рендеринг
	CustomShader string   // Идентификатор пользовательского шейдера
}

// NewRenderComponent создает новый компонент рендеринга
func NewRenderComponent(modelID, textureID string) *RenderComponent {
	return &RenderComponent{
		BaseComponent: NewBaseComponent(RenderComponentID),
		ModelID:       modelID,
		TextureID:     textureID,
		Color:         color.RGBA{255, 255, 255, 255},
		Visible:       true,
		CastShadow:    true,
		Layer:         0,
		Distortion:    0,
		Effects:       make([]string, 0),
		Pixel:         true, // По умолчанию используем пиксельный рендеринг
	}
}

// PhysicsComponent содержит информацию для физического моделирования
type PhysicsComponent struct {
	BaseComponent
	Velocity     Vector3
	Acceleration Vector3
	Mass         float64
	Friction     float64
	Restitution  float64 // Упругость (коэффициент восстановления)
	Gravity      float64 // Модификатор гравитации (1.0 - нормальная)
	Static       bool    // Статический объект (не двигается)
	Collider     string  // Тип коллайдера: "box", "sphere", "capsule", "mesh"
	ColliderSize Vector3 // Размеры коллайдера
	IsTrigger    bool    // Является ли триггером (не создает физических столкновений)
}

// NewPhysicsComponent создает новый компонент физики
func NewPhysicsComponent(mass float64, static bool) *PhysicsComponent {
	return &PhysicsComponent{
		BaseComponent: NewBaseComponent(PhysicsComponentID),
		Mass:          mass,
		Friction:      0.5,
		Restitution:   0.3,
		Gravity:       1.0,
		Static:        static,
		Collider:      "box",
		ColliderSize:  Vector3{1, 1, 1},
	}
}

// HealthComponent содержит информацию о здоровье и состоянии сущности
type HealthComponent struct {
	BaseComponent
	MaxHealth         float64
	CurrentHealth     float64
	IsInvulnerable    bool
	RegenRate         float64            // Скорость восстановления здоровья в единицах в секунду
	DamageMultipliers map[string]float64 // Множители урона по типам
	StatusEffects     map[string]float64 // Текущие эффекты состояния и их длительность
}

// NewHealthComponent создает новый компонент здоровья
func NewHealthComponent(maxHealth float64) *HealthComponent {
	return &HealthComponent{
		BaseComponent:     NewBaseComponent(HealthComponentID),
		MaxHealth:         maxHealth,
		CurrentHealth:     maxHealth,
		IsInvulnerable:    false,
		RegenRate:         0,
		DamageMultipliers: make(map[string]float64),
		StatusEffects:     make(map[string]float64),
	}
}

// IsDead проверяет, мертва ли сущность
func (h *HealthComponent) IsDead() bool {
	return h.CurrentHealth <= 0
}

// TakeDamage наносит урон сущности
func (h *HealthComponent) TakeDamage(amount float64, damageType string) float64 {
	if h.IsInvulnerable {
		return 0
	}

	// Применяем модификаторы урона
	if multiplier, exists := h.DamageMultipliers[damageType]; exists {
		amount *= multiplier
	}

	h.CurrentHealth -= amount
	if h.CurrentHealth < 0 {
		h.CurrentHealth = 0
	}

	return amount
}

// Heal восстанавливает здоровье сущности
func (h *HealthComponent) Heal(amount float64) float64 {
	oldHealth := h.CurrentHealth
	h.CurrentHealth += amount
	if h.CurrentHealth > h.MaxHealth {
		h.CurrentHealth = h.MaxHealth
	}

	return h.CurrentHealth - oldHealth
}

// MetamorphicComponent определяет, как сущность может изменяться под воздействием метаморфоз
type MetamorphicComponent struct {
	BaseComponent
	Stability            float64            // 0-1: насколько устойчив к изменениям (1 = невосприимчив)
	AbnormalityIndex     float64            // 0-1: насколько сильно уже изменен
	CurrentMetamorphoses []string           // Идентификаторы активных метаморфоз
	PossibleMutations    []string           // Идентификаторы возможных мутаций
	PropertyModifiers    map[string]float64 // Модификаторы свойств от метаморфоз
}

// NewMetamorphicComponent создает новый компонент метаморфичности
func NewMetamorphicComponent(stability float64) *MetamorphicComponent {
	return &MetamorphicComponent{
		BaseComponent:        NewBaseComponent(MetamorphicComponentID),
		Stability:            stability,
		AbnormalityIndex:     0,
		CurrentMetamorphoses: make([]string, 0),
		PossibleMutations:    make([]string, 0),
		PropertyModifiers:    make(map[string]float64),
	}
}

// CanMutate проверяет, может ли сущность мутировать под воздействием метаморфозы
func (m *MetamorphicComponent) CanMutate(metamorphStrength float64) bool {
	// Если стабильность 1, то не может мутировать
	if m.Stability >= 1.0 {
		return false
	}

	// Вероятность мутации зависит от силы метаморфозы и стабильности сущности
	mutationChance := metamorphStrength * (1.0 - m.Stability)
	return rand.Float64() < mutationChance
}

// ApplyMetamorphosis применяет метаморфозу к сущности
func (m *MetamorphicComponent) ApplyMetamorphosis(metamorphID string, intensity float64) {
	// Добавляем метаморфозу в список активных
	m.CurrentMetamorphoses = append(m.CurrentMetamorphoses, metamorphID)

	// Увеличиваем индекс аномальности
	m.AbnormalityIndex += intensity * (1.0 - m.Stability)
	if m.AbnormalityIndex > 1.0 {
		m.AbnormalityIndex = 1.0
	}
}

// SurvivalComponent содержит информацию о выживаемости (для игрока и NPC)
type SurvivalComponent struct {
	BaseComponent
	Hunger          float64 // 0-100
	Thirst          float64 // 0-100
	Temperature     float64 // Температура тела в градусах Цельсия
	Fatigue         float64 // 0-100: усталость
	SanityLevel     float64 // 0-100: уровень рассудка
	HungerRate      float64 // Скорость увеличения голода
	ThirstRate      float64 // Скорость увеличения жажды
	FatigueRate     float64 // Скорость увеличения усталости
	SanityDecayRate float64 // Скорость снижения рассудка
}

// NewSurvivalComponent создает новый компонент выживания
func NewSurvivalComponent() *SurvivalComponent {
	return &SurvivalComponent{
		BaseComponent:   NewBaseComponent(SurvivalComponentID),
		Hunger:          0,
		Thirst:          0,
		Temperature:     37.0, // Нормальная температура тела
		Fatigue:         0,
		SanityLevel:     100,
		HungerRate:      0.5, // 0.5 единиц в минуту
		ThirstRate:      0.8, // 0.8 единиц в минуту
		FatigueRate:     0.3, // 0.3 единиц в минуту
		SanityDecayRate: 0.1, // 0.1 единиц в минуту
	}
}

// Update обновляет состояние компонента выживания
func (s *SurvivalComponent) Update(deltaTime float64) {
	// Преобразуем deltaTime из секунд в минуты для расчетов
	deltaMinutes := deltaTime / 60.0

	// Обновляем показатели
	s.Hunger += s.HungerRate * deltaMinutes
	s.Thirst += s.ThirstRate * deltaMinutes
	s.Fatigue += s.FatigueRate * deltaMinutes
	s.SanityLevel -= s.SanityDecayRate * deltaMinutes

	// Ограничиваем значения
	s.Hunger = math.Max(0, math.Min(100, s.Hunger))
	s.Thirst = math.Max(0, math.Min(100, s.Thirst))
	s.Fatigue = math.Max(0, math.Min(100, s.Fatigue))
	s.SanityLevel = math.Max(0, math.Min(100, s.SanityLevel))
}

// GetSurvivalStatus возвращает общий статус выживания
func (s *SurvivalComponent) GetSurvivalStatus() string {
	if s.Hunger >= 90 || s.Thirst >= 90 {
		return "critical"
	} else if s.Hunger >= 70 || s.Thirst >= 70 || s.Fatigue >= 90 || s.SanityLevel <= 20 {
		return "bad"
	} else if s.Hunger >= 50 || s.Thirst >= 50 || s.Fatigue >= 70 || s.SanityLevel <= 50 {
		return "moderate"
	} else {
		return "good"
	}
}

// SymbolComponent содержит информацию о символе (для системы ритуалов)
type SymbolComponent struct {
	BaseComponent
	SymbolID        string
	SymbolType      string
	Complexity      float64 // 0-1: сложность символа
	Power           float64 // 0-1: сила символа
	Discovered      bool    // Обнаружен ли символ игроком
	DiscoveryRadius float64
	KnowledgeLevel  float64  // 0-1: насколько хорошо игрок понимает символ
	RelatedSymbols  []string // Связанные символы
	Meaning         []string // Набор значений символа
	VisualData      string   // Ссылка на визуальное представление
}

// NewSymbolComponent создает новый компонент символа
func NewSymbolComponent(symbolID, symbolType string, complexity, power float64) *SymbolComponent {
	return &SymbolComponent{
		BaseComponent:  NewBaseComponent(SymbolComponentID),
		SymbolID:       symbolID,
		SymbolType:     symbolType,
		Complexity:     complexity,
		Power:          power,
		Discovered:     false,
		KnowledgeLevel: 0,
		RelatedSymbols: make([]string, 0),
		Meaning:        make([]string, 0),
	}
}

// LightComponent содержит информацию об источнике света
type LightComponent struct {
	BaseComponent
	Color       color.RGBA
	Intensity   float64 // Сила света
	Range       float64 // Радиус действия
	Flickering  bool    // Мерцание
	CastShadows bool    // Отбрасывание теней
	SpotLight   bool    // Прожектор или точечный свет
	SpotAngle   float64 // Угол конуса для прожектора (в радианах)
}

// NewLightComponent создает новый компонент света
func NewLightComponent(color color.RGBA, intensity, range_ float64) *LightComponent {
	return &LightComponent{
		BaseComponent: NewBaseComponent(LightComponentID),
		Color:         color,
		Intensity:     intensity,
		Range:         range_,
		Flickering:    false,
		CastShadows:   true,
		SpotLight:     false,
		SpotAngle:     math.Pi / 4, // 45 градусов по умолчанию
	}
}

// InventoryComponent содержит информацию об инвентаре сущности
type InventoryComponent struct {
	BaseComponent
	Items         []EntityID          // Идентификаторы предметов в инвентаре
	Capacity      int                 // Максимальное количество предметов
	MaxWeight     float64             // Максимальный вес
	CurrentWeight float64             // Текущий вес
	Equipped      map[string]EntityID // Экипированные предметы по слотам
}

// NewInventoryComponent создает новый компонент инвентаря
func NewInventoryComponent(capacity int, maxWeight float64) *InventoryComponent {
	return &InventoryComponent{
		BaseComponent: NewBaseComponent(InventoryComponentID),
		Items:         make([]EntityID, 0),
		Capacity:      capacity,
		MaxWeight:     maxWeight,
		CurrentWeight: 0,
		Equipped:      make(map[string]EntityID),
	}
}

// AddItem добавляет предмет в инвентарь
func (i *InventoryComponent) AddItem(itemID EntityID, weight float64) bool {
	if len(i.Items) >= i.Capacity || i.CurrentWeight+weight > i.MaxWeight {
		return false
	}

	i.Items = append(i.Items, itemID)
	i.CurrentWeight += weight
	return true
}

// RemoveItem удаляет предмет из инвентаря
func (i *InventoryComponent) RemoveItem(itemID EntityID, weight float64) bool {
	for idx, id := range i.Items {
		if id == itemID {
			// Удаляем предмет, сохраняя порядок
			i.Items = append(i.Items[:idx], i.Items[idx+1:]...)
			i.CurrentWeight -= weight

			// Проверяем, был ли предмет экипирован
			for slot, equippedID := range i.Equipped {
				if equippedID == itemID {
					delete(i.Equipped, slot)
					break
				}
			}

			return true
		}
	}
	return false
}

// EquipItem экипирует предмет в указанный слот
func (i *InventoryComponent) EquipItem(itemID EntityID, slot string) bool {
	// Проверяем, есть ли предмет в инвентаре
	hasItem := false
	for _, id := range i.Items {
		if id == itemID {
			hasItem = true
			break
		}
	}

	if !hasItem {
		return false
	}

	// Если в слоте уже есть предмет, снимаем его

	delete(i.Equipped, slot)

	// Экипируем новый предмет
	i.Equipped[slot] = itemID
	return true
}

// GetEquippedItem возвращает экипированный предмет в указанном слоте
func (i *InventoryComponent) GetEquippedItem(slot string) (EntityID, bool) {
	itemID, exists := i.Equipped[slot]
	return itemID, exists
}

// PlayerControlComponent содержит информацию об управлении игроком
type PlayerControlComponent struct {
	BaseComponent
	MovementSpeed    float64 // Скорость передвижения
	RotationSpeed    float64 // Скорость поворота
	JumpForce        float64 // Сила прыжка
	CanJump          bool    // Может ли игрок прыгать
	CanSprint        bool    // Может ли игрок бежать
	SprintMultiplier float64 // Множитель скорости при беге
	IsGrounded       bool    // Находится ли игрок на земле
	IsCrouching      bool    // Присел ли игрок
	CrouchMultiplier float64 // Множитель скорости при приседании
	IsInteracting    bool    // Взаимодействует ли игрок с чем-то
	LastInteractTime float64 // Время последнего взаимодействия
}

// NewPlayerControlComponent создает новый компонент управления игроком
func NewPlayerControlComponent() *PlayerControlComponent {
	return &PlayerControlComponent{
		BaseComponent:    NewBaseComponent(PlayerControlComponentID),
		MovementSpeed:    5.0, // Единиц в секунду
		RotationSpeed:    3.0, // Радиан в секунду
		JumpForce:        8.0,
		CanJump:          true,
		CanSprint:        true,
		SprintMultiplier: 1.5,
		IsGrounded:       true,
		IsCrouching:      false,
		CrouchMultiplier: 0.5,
		IsInteracting:    false,
	}
}

// GetCurrentMovementSpeed возвращает текущую скорость передвижения с учетом модификаторов
func (p *PlayerControlComponent) GetCurrentMovementSpeed() float64 {
	speed := p.MovementSpeed

	if p.IsCrouching {
		speed *= p.CrouchMultiplier
	}

	return speed
}

// GetCurrentSprintSpeed возвращает текущую скорость бега с учетом модификаторов
func (p *PlayerControlComponent) GetCurrentSprintSpeed() float64 {
	return p.GetCurrentMovementSpeed() * p.SprintMultiplier
}

// InteractableComponent содержит информацию о взаимодействии с сущностью
type InteractableComponent struct {
	BaseComponent
	InteractionType   string                      // Тип взаимодействия: "pickup", "examine", "use", "talk", "ritual"
	InteractionPrompt string                      // Текст подсказки при взаимодействии
	InteractionRange  float64                     // Расстояние, с которого можно взаимодействовать
	RequiredItems     []string                    // Предметы, необходимые для взаимодействия
	CooldownTime      float64                     // Время перезарядки взаимодействия
	LastInteractTime  float64                     // Время последнего взаимодействия
	InteractCallback  func(*Entity, *Entity) bool // Функция обратного вызова при взаимодействии
}

// NewInteractableComponent создает новый компонент взаимодействия
func NewInteractableComponent(interactionType, prompt string, range_ float64) *InteractableComponent {
	return &InteractableComponent{
		BaseComponent:     NewBaseComponent(InteractableComponentID),
		InteractionType:   interactionType,
		InteractionPrompt: prompt,
		InteractionRange:  range_,
		RequiredItems:     make([]string, 0),
		CooldownTime:      0,
		LastInteractTime:  0,
	}
}

// CanInteract проверяет, можно ли взаимодействовать с сущностью
func (i *InteractableComponent) CanInteract(time float64) bool {
	return time-i.LastInteractTime >= i.CooldownTime
}

// Interact выполняет взаимодействие с сущностью
func (i *InteractableComponent) Interact(actor, target *Entity, time float64) bool {
	if !i.CanInteract(time) {
		return false
	}

	if i.InteractCallback != nil {
		result := i.InteractCallback(actor, target)
		if result {
			i.LastInteractTime = time
		}
		return result
	}

	// Если нет колбэка, считаем взаимодействие успешным
	i.LastInteractTime = time
	return true
}

// SoundEmitterComponent содержит информацию о звуках, издаваемых сущностью
type SoundEmitterComponent struct {
	BaseComponent
	SoundID          string            // Идентификатор звука
	Volume           float64           // Громкость (0-1)
	Pitch            float64           // Высота тона (0.5-2.0)
	Range            float64           // Радиус слышимости
	IsLooping        bool              // Зацикленный звук
	IsPlaying        bool              // Проигрывается ли звук
	PlayOnStart      bool              // Проигрывать при создании сущности
	RandomPitchRange float64           // Диапазон случайного изменения высоты тона
	Sounds           map[string]string // Словарь доступных звуков по ключам
}

// NewSoundEmitterComponent создает новый компонент звука
func NewSoundEmitterComponent(soundID string, volume, range_ float64) *SoundEmitterComponent {
	return &SoundEmitterComponent{
		BaseComponent:    NewBaseComponent(SoundEmitterComponentID),
		SoundID:          soundID,
		Volume:           volume,
		Pitch:            1.0,
		Range:            range_,
		IsLooping:        false,
		IsPlaying:        false,
		PlayOnStart:      false,
		RandomPitchRange: 0,
		Sounds:           make(map[string]string),
	}
}

// Play начинает проигрывание звука
func (s *SoundEmitterComponent) Play() {
	s.IsPlaying = true
}

// Stop останавливает проигрывание звука
func (s *SoundEmitterComponent) Stop() {
	s.IsPlaying = false
}

// PlaySound проигрывает указанный звук
func (s *SoundEmitterComponent) PlaySound(key string) bool {
	if soundID, exists := s.Sounds[key]; exists {
		s.SoundID = soundID
		s.Play()
		return true
	}
	return false
}

// AIComponent содержит информацию об искусственном интеллекте сущности
type AIComponent struct {
	BaseComponent
	AIType             string    // Тип ИИ: "passive", "neutral", "aggressive", "scared", "smart"
	DetectionRange     float64   // Радиус обнаружения
	CurrentState       string    // Текущее состояние: "idle", "patrol", "chase", "attack", "flee"
	TargetID           EntityID  // Идентификатор цели
	LastKnownTargetPos Vector3   // Последняя известная позиция цели
	PatrolPoints       []Vector3 // Точки патрулирования
	CurrentPatrolIdx   int       // Индекс текущей точки патрулирования
	AttackDamage       float64   // Урон от атаки
	AttackRange        float64   // Радиус атаки
	AttackCooldown     float64   // Время перезарядки атаки
	LastAttackTime     float64   // Время последней атаки
	FearLevel          float64   // Уровень страха (0-1)
	AwarenessLevel     float64   // Уровень осведомленности (0-1)
	Behaviors          []string  // Список поведений
}

// NewAIComponent создает новый компонент ИИ
func NewAIComponent(aiType string, detectionRange float64) *AIComponent {
	return &AIComponent{
		BaseComponent:    NewBaseComponent(AIComponentID),
		AIType:           aiType,
		DetectionRange:   detectionRange,
		CurrentState:     "idle",
		PatrolPoints:     make([]Vector3, 0),
		CurrentPatrolIdx: 0,
		AttackDamage:     10,
		AttackRange:      1.5,
		AttackCooldown:   1.0,
		FearLevel:        0,
		AwarenessLevel:   0,
		Behaviors:        make([]string, 0),
	}
}

// CanAttack проверяет, может ли сущность атаковать
func (a *AIComponent) CanAttack(time float64) bool {
	return time-a.LastAttackTime >= a.AttackCooldown
}

// Attack выполняет атаку
func (a *AIComponent) Attack(time float64) float64 {
	if !a.CanAttack(time) {
		return 0
	}

	a.LastAttackTime = time
	return a.AttackDamage
}

// SetState устанавливает новое состояние ИИ
func (a *AIComponent) SetState(state string) {
	a.CurrentState = state
}

// AddPatrolPoint добавляет точку патрулирования
func (a *AIComponent) AddPatrolPoint(point Vector3) {
	a.PatrolPoints = append(a.PatrolPoints, point)
}

// GetNextPatrolPoint возвращает следующую точку патрулирования
func (a *AIComponent) GetNextPatrolPoint() Vector3 {
	if len(a.PatrolPoints) == 0 {
		return Vector3{}
	}

	point := a.PatrolPoints[a.CurrentPatrolIdx]
	a.CurrentPatrolIdx = (a.CurrentPatrolIdx + 1) % len(a.PatrolPoints)
	return point
}
