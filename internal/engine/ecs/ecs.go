package ecs

import (
	"reflect"
	"sync"

	"github.com/google/uuid"
)

// ComponentID уникально идентифицирует тип компонента
type ComponentID string

// EntityID уникально идентифицирует сущность
type EntityID string

// Component представляет базовый интерфейс для всех компонентов
type Component interface {
	// Type возвращает уникальный ID типа компонента
	Type() ComponentID
}

// Entity представляет игровую сущность
type Entity struct {
	ID         EntityID
	components map[ComponentID]Component
	tags       map[string]bool
	world      *World
}

// NewEntity создает новую сущность
func NewEntity() *Entity {
	return &Entity{
		ID:         EntityID(uuid.New().String()),
		components: make(map[ComponentID]Component),
		tags:       make(map[string]bool),
	}
}

// AddComponent добавляет компонент к сущности
func (e *Entity) AddComponent(c Component) {
	e.components[c.Type()] = c
	if e.world != nil {
		e.world.entityComponentChanged(e, c, true)
	}
}

// RemoveComponent удаляет компонент из сущности
func (e *Entity) RemoveComponent(id ComponentID) {
	if comp, exists := e.components[id]; exists {
		delete(e.components, id)
		if e.world != nil {
			e.world.entityComponentChanged(e, comp, false)
		}
	}
}

// GetComponent возвращает компонент указанного типа
func (e *Entity) GetComponent(id ComponentID) (Component, bool) {
	comp, exists := e.components[id]
	return comp, exists
}

// HasComponent проверяет, есть ли у сущности компонент указанного типа
func (e *Entity) HasComponent(id ComponentID) bool {
	_, exists := e.components[id]
	return exists
}

// HasAllComponents проверяет, есть ли у сущности все указанные компоненты
func (e *Entity) HasAllComponents(ids ...ComponentID) bool {
	for _, id := range ids {
		if !e.HasComponent(id) {
			return false
		}
	}
	return true
}

// AddTag добавляет тег к сущности
func (e *Entity) AddTag(tag string) {
	e.tags[tag] = true
	if e.world != nil {
		e.world.entityTagChanged(e, tag, true)
	}
}

// RemoveTag удаляет тег из сущности
func (e *Entity) RemoveTag(tag string) {
	if _, exists := e.tags[tag]; exists {
		delete(e.tags, tag)
		if e.world != nil {
			e.world.entityTagChanged(e, tag, false)
		}
	}
}

// HasTag проверяет, есть ли у сущности указанный тег
func (e *Entity) HasTag(tag string) bool {
	_, exists := e.tags[tag]
	return exists
}

// GetTags возвращает все теги сущности
func (e *Entity) GetTags() []string {
	tags := make([]string, 0, len(e.tags))
	for tag := range e.tags {
		tags = append(tags, tag)
	}
	return tags
}

// System представляет систему, которая обрабатывает сущности с определенными компонентами
type System interface {
	// Update обновляет состояние системы и связанных сущностей
	Update(deltaTime float64)

	// RequiredComponents возвращает список компонентов, необходимых для работы системы
	RequiredComponents() []ComponentID
}

// World представляет игровой мир, содержащий все сущности и системы
type World struct {
	entities map[EntityID]*Entity
	systems  []System

	// Индексы для быстрого доступа
	entitiesByComponent map[ComponentID]map[EntityID]*Entity
	entitiesByTag       map[string]map[EntityID]*Entity

	// Блокировки для безопасного доступа из разных горутин
	entitiesMutex sync.RWMutex
	systemsMutex  sync.RWMutex
}

// NewWorld создает новый игровой мир
func NewWorld() *World {
	return &World{
		entities:            make(map[EntityID]*Entity),
		systems:             make([]System, 0),
		entitiesByComponent: make(map[ComponentID]map[EntityID]*Entity),
		entitiesByTag:       make(map[string]map[EntityID]*Entity),
	}
}

// AddEntity добавляет сущность в мир
func (w *World) AddEntity(e *Entity) {
	w.entitiesMutex.Lock()
	defer w.entitiesMutex.Unlock()

	w.entities[e.ID] = e
	e.world = w

	// Индексируем компоненты
	for _, comp := range e.components {
		w.indexEntityComponent(e, comp.Type(), true)
	}

	// Индексируем теги
	for tag := range e.tags {
		w.indexEntityTag(e, tag, true)
	}
}

// RemoveEntity удаляет сущность из мира
func (w *World) RemoveEntity(id EntityID) {
	w.entitiesMutex.Lock()
	defer w.entitiesMutex.Unlock()

	if e, exists := w.entities[id]; exists {
		// Удаляем из индексов компонентов
		for compID := range e.components {
			w.indexEntityComponent(e, compID, false)
		}

		// Удаляем из индексов тегов
		for tag := range e.tags {
			w.indexEntityTag(e, tag, false)
		}

		delete(w.entities, id)
		e.world = nil
	}
}

// GetEntity возвращает сущность по ID
func (w *World) GetEntity(id EntityID) (*Entity, bool) {
	w.entitiesMutex.RLock()
	defer w.entitiesMutex.RUnlock()

	e, exists := w.entities[id]
	return e, exists
}

// GetEntities возвращает все сущности в мире
func (w *World) GetEntities() []*Entity {
	w.entitiesMutex.RLock()
	defer w.entitiesMutex.RUnlock()

	entities := make([]*Entity, 0, len(w.entities))
	for _, e := range w.entities {
		entities = append(entities, e)
	}
	return entities
}

// GetEntitiesWithComponent возвращает все сущности с указанным компонентом
func (w *World) GetEntitiesWithComponent(compID ComponentID) []*Entity {
	w.entitiesMutex.RLock()
	defer w.entitiesMutex.RUnlock()

	entities := make([]*Entity, 0)
	if compMap, exists := w.entitiesByComponent[compID]; exists {
		for _, e := range compMap {
			entities = append(entities, e)
		}
	}
	return entities
}

// GetEntitiesWithAllComponents возвращает все сущности со всеми указанными компонентами
func (w *World) GetEntitiesWithAllComponents(compIDs ...ComponentID) []*Entity {
	if len(compIDs) == 0 {
		return nil
	}

	w.entitiesMutex.RLock()
	defer w.entitiesMutex.RUnlock()

	// Находим компонент с наименьшим количеством сущностей
	var minCompID ComponentID
	minCount := -1

	for _, compID := range compIDs {
		if compMap, exists := w.entitiesByComponent[compID]; exists {
			count := len(compMap)
			if minCount == -1 || count < minCount {
				minCount = count
				minCompID = compID
			}
		} else {
			// Если нет сущностей с одним из компонентов, сразу возвращаем пустой слайс
			return []*Entity{}
		}
	}

	// Начинаем с наименьшего набора и фильтруем
	var entities []*Entity
	if compMap, exists := w.entitiesByComponent[minCompID]; exists {
		for _, e := range compMap {
			allComponents := true
			for _, compID := range compIDs {
				if compID != minCompID && !e.HasComponent(compID) {
					allComponents = false
					break
				}
			}
			if allComponents {
				entities = append(entities, e)
			}
		}
	}

	return entities
}

// GetEntitiesWithTag возвращает все сущности с указанным тегом
func (w *World) GetEntitiesWithTag(tag string) []*Entity {
	w.entitiesMutex.RLock()
	defer w.entitiesMutex.RUnlock()

	entities := make([]*Entity, 0)
	if tagMap, exists := w.entitiesByTag[tag]; exists {
		for _, e := range tagMap {
			entities = append(entities, e)
		}
	}
	return entities
}

// AddSystem добавляет систему в мир
func (w *World) AddSystem(s System) {
	w.systemsMutex.Lock()
	defer w.systemsMutex.Unlock()

	w.systems = append(w.systems, s)
}

// RemoveSystem удаляет систему из мира
func (w *World) RemoveSystem(system System) {
	w.systemsMutex.Lock()
	defer w.systemsMutex.Unlock()

	systemType := reflect.TypeOf(system)
	for i, s := range w.systems {
		if reflect.TypeOf(s) == systemType {
			// Удаляем систему, сохраняя порядок остальных
			w.systems = append(w.systems[:i], w.systems[i+1:]...)
			return
		}
	}
}

// Update обновляет все системы в мире
func (w *World) Update(deltaTime float64) {
	w.systemsMutex.RLock()
	systems := w.systems // Создаем копию для безопасного итерирования
	w.systemsMutex.RUnlock()

	for _, system := range systems {
		system.Update(deltaTime)
	}
}

// Внутренние методы для индексирования

// entityComponentChanged вызывается, когда компонент сущности изменяется
func (w *World) entityComponentChanged(e *Entity, c Component, added bool) {
	w.entitiesMutex.Lock()
	defer w.entitiesMutex.Unlock()

	w.indexEntityComponent(e, c.Type(), added)
}

// entityTagChanged вызывается, когда тег сущности изменяется
func (w *World) entityTagChanged(e *Entity, tag string, added bool) {
	w.entitiesMutex.Lock()
	defer w.entitiesMutex.Unlock()

	w.indexEntityTag(e, tag, added)
}

// indexEntityComponent индексирует сущность по компоненту
func (w *World) indexEntityComponent(e *Entity, compID ComponentID, add bool) {
	if add {
		if _, exists := w.entitiesByComponent[compID]; !exists {
			w.entitiesByComponent[compID] = make(map[EntityID]*Entity)
		}
		w.entitiesByComponent[compID][e.ID] = e
	} else {
		if compMap, exists := w.entitiesByComponent[compID]; exists {
			delete(compMap, e.ID)
		}
	}
}

// indexEntityTag индексирует сущность по тегу
func (w *World) indexEntityTag(e *Entity, tag string, add bool) {
	if add {
		if _, exists := w.entitiesByTag[tag]; !exists {
			w.entitiesByTag[tag] = make(map[EntityID]*Entity)
		}
		w.entitiesByTag[tag][e.ID] = e
	} else {
		if tagMap, exists := w.entitiesByTag[tag]; exists {
			delete(tagMap, e.ID)
		}
	}
}

// RegisterComponentType регистрирует новый тип компонента и возвращает его ID
func RegisterComponentType(name string) ComponentID {
	return ComponentID(name)
}

// BaseComponent предоставляет базовую реализацию интерфейса Component
type BaseComponent struct {
	TypeID ComponentID
}

// Type возвращает ID типа компонента
func (bc *BaseComponent) Type() ComponentID {
	return bc.TypeID
}

// NewBaseComponent создает новый базовый компонент с указанным ID типа
func NewBaseComponent(typeID ComponentID) BaseComponent {
	return BaseComponent{TypeID: typeID}
}
