package rituals

import (
	"echo-taiga/internal/metamorphosis"
	"echo-taiga/internal/symbols"
)

// Компонент ритуала
type Component struct {
	Type  string      // item, action, location, etc.
	Value interface{} // Конкретное значение компонента
}

// Эффект ритуала
type Effect struct {
	Type       string  // player_ability, metamorphosis, etc.
	Target     string  // На что влияет
	Value      float64 // Величина эффекта
	Duration   int     // Длительность в секундах (0 = постоянно)
	OrderLevel int     // Для метаморфоз - порядок изменений
}

// Ritual представляет ритуал, который может выполнить игрок
type Ritual struct {
	ID              string
	Name            string
	RequiredSymbols []symbols.Symbol
	Components      []Component
	Effects         []Effect
	MasteryLevel    float64 // От 0 до 1, влияет на эффективность
	SuccessChance   float64 // Базовый шанс успеха
}

// Perform выполняет ритуал с указанными компонентами
func (r *Ritual) Perform(providedComponents []Component) (bool, []Effect, error) {
	// TODO: реализовать логику выполнения ритуала
	return false, nil, nil
}

// RitualManager управляет всеми ритуалами в игре
type RitualManager struct {
	knownRituals   map[string]*Ritual
	metamorphMgr   *metamorphosis.Manager
	symbolRegistry *symbols.Registry
}

// NewRitualManager создает новый менеджер ритуалов
func NewRitualManager(metamorphMgr *metamorphosis.Manager, symbolRegistry *symbols.Registry) *RitualManager {
	return &RitualManager{
		knownRituals:   make(map[string]*Ritual),
		metamorphMgr:   metamorphMgr,
		symbolRegistry: symbolRegistry,
	}
}

// RegisterRitual регистрирует новый ритуал
func (rm *RitualManager) RegisterRitual(ritual *Ritual) {
	rm.knownRituals[ritual.ID] = ritual
}

// GetRitual возвращает ритуал по ID
func (rm *RitualManager) GetRitual(id string) (*Ritual, bool) {
	ritual, exists := rm.knownRituals[id]
	return ritual, exists
}
