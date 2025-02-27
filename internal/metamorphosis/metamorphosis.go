package metamorphosis

// OrderLevel определяет уровень (порядок) метаморфозы
type OrderLevel int

// Effect представляет конкретный эффект метаморфозы
type Effect interface {
	ID() string
	Order() OrderLevel
	Apply(target interface{}) error
	Revert(target interface{}) error
	Intensity() float64
}

// Trigger определяет условие, при котором активируется метаморфоза
type Trigger interface {
	ID() string
	Check(state interface{}) bool
	Priority() float64
}

// Manager управляет всеми метаморфозами в игре
type Manager struct {
	activeEffects     []Effect
	availableTriggers []Trigger
	budget            float64
	maxBudget         float64
	regenerationRate  float64
}

// NewManager создает новый менеджер метаморфоз
func NewManager() *Manager {
	return &Manager{
		activeEffects:     make([]Effect, 0),
		availableTriggers: make([]Trigger, 0),
		budget:            100.0,
		maxBudget:         100.0,
		regenerationRate:  0.5,
	}
}

// Update обновляет состояние всех метаморфоз
func (m *Manager) Update(deltaTime float64) {
	// TODO: реализовать логику обновления
}

// ApplyEffect применяет эффект метаморфозы к цели
func (m *Manager) ApplyEffect(effect Effect, target interface{}) error {
	// TODO: реализовать логику применения эффекта
	return nil
}

// GetEffectsByOrder возвращает все активные эффекты указанного порядка
func (m *Manager) GetEffectsByOrder(order OrderLevel) []Effect {
	// TODO: реализовать фильтрацию эффектов
	return nil
}
