package engine

import (
	"echo-taiga/internal/engine/ecs"
	"math"
	"math/rand/v2"
	"time"
)

// Engine представляет основной игровой движок
type Engine struct {
	world *ecs.World
	// Подсистемы движка
	physicsSystem   *PhysicsSystem
	collisionSystem *CollisionSystem
	aiSystem        *AISystem
}

// PhysicsSystem отвечает за физическую симуляцию
type PhysicsSystem struct {
	world *ecs.World
}

// CollisionSystem отвечает за обработку столкновений
type CollisionSystem struct {
	world *ecs.World
}

// AISystem отвечает за поведение искусственного интеллекта
type AISystem struct {
	world *ecs.World
}

// NewEngine создает новый игровой движок
func NewEngine(world *ecs.World) *Engine {
	engine := &Engine{
		world: world,
	}

	// Инициализация подсистем
	engine.initializeSystems()

	return engine
}

// initializeSystems инициализирует все системы движка
func (e *Engine) initializeSystems() {
	// Создаем и регистрируем физическую систему
	e.physicsSystem = &PhysicsSystem{world: e.world}
	e.world.AddSystem(e.physicsSystem)

	// Создаем и регистрируем систему столкновений
	e.collisionSystem = &CollisionSystem{world: e.world}
	e.world.AddSystem(e.collisionSystem)

	// Создаем и регистрируем систему ИИ
	e.aiSystem = &AISystem{world: e.world}
	e.world.AddSystem(e.aiSystem)
}

// RequiredComponents возвращает компоненты, необходимые для работы системы физики
func (ps *PhysicsSystem) RequiredComponents() []ecs.ComponentID {
	return []ecs.ComponentID{
		ecs.TransformComponentID,
		ecs.PhysicsComponentID,
	}
}

// Update обновляет физическую симуляцию
func (ps *PhysicsSystem) Update(deltaTime float64) {
	// Получаем все сущности с физическими компонентами
	entities := ps.world.GetEntitiesWithComponent(ecs.PhysicsComponentID)

	for _, entity := range entities {
		// Получаем компоненты трансформации и физики
		transformComp, _ := entity.GetComponent(ecs.TransformComponentID)
		physicsComp, _ := entity.GetComponent(ecs.PhysicsComponentID)

		transform := transformComp.(*ecs.TransformComponent)
		physics := physicsComp.(*ecs.PhysicsComponent)

		// Статические объекты не обновляются
		if physics.Static {
			continue
		}

		// Применяем гравитацию
		physics.Velocity.Y -= 9.8 * physics.Gravity * deltaTime

		// Обновляем позицию на основе скорости
		transform.Position = transform.Position.Add(
			physics.Velocity.Multiply(deltaTime),
		)

		// Затухание скорости из-за трения
		physics.Velocity = physics.Velocity.Multiply(1 - physics.Friction*deltaTime)
	}
}

// RequiredComponents возвращает компоненты, необходимые для системы столкновений
func (cs *CollisionSystem) RequiredComponents() []ecs.ComponentID {
	return []ecs.ComponentID{
		ecs.TransformComponentID,
		ecs.PhysicsComponentID,
	}
}

// Update обновляет систему столкновений
func (cs *CollisionSystem) Update(deltaTime float64) {
	// Получаем все сущности с физическими компонентами
	entities := cs.world.GetEntitiesWithComponent(ecs.PhysicsComponentID)

	// Проверяем столкновения между сущностями
	for i, entityA := range entities {
		for j := i + 1; j < len(entities); j++ {
			entityB := entities[j]

			// Проверяем столкновение
			if checkCollision(entityA, entityB) {
				resolveCollision(entityA, entityB)
			}
		}
	}
}

// checkCollision проверяет столкновение между двумя сущностями
func checkCollision(entityA, entityB *ecs.Entity) bool {
	// Получаем компоненты трансформации и физики
	transformA, _ := entityA.GetComponent(ecs.TransformComponentID)
	transformB, _ := entityB.GetComponent(ecs.TransformComponentID)

	physicsA, _ := entityA.GetComponent(ecs.PhysicsComponentID)
	physicsB, _ := entityB.GetComponent(ecs.PhysicsComponentID)

	posA := transformA.(*ecs.TransformComponent).Position
	posB := transformB.(*ecs.TransformComponent).Position

	sizeA := physicsA.(*ecs.PhysicsComponent).ColliderSize
	sizeB := physicsB.(*ecs.PhysicsComponent).ColliderSize

	// Простая проверка на пересечение AABB (Axis-Aligned Bounding Box)
	return (posA.X < posB.X+sizeB.X &&
		posA.X+sizeA.X > posB.X &&
		posA.Y < posB.Y+sizeB.Y &&
		posA.Y+sizeA.Y > posB.Y &&
		posA.Z < posB.Z+sizeB.Z &&
		posA.Z+sizeA.Z > posB.Z)
}

// resolveCollision разрешает столкновение между сущностями
func resolveCollision(entityA, entityB *ecs.Entity) {
	transformA, _ := entityA.GetComponent(ecs.TransformComponentID)
	transformB, _ := entityB.GetComponent(ecs.TransformComponentID)

	physicsA, _ := entityA.GetComponent(ecs.PhysicsComponentID)
	physicsB, _ := entityB.GetComponent(ecs.PhysicsComponentID)

	trA := transformA.(*ecs.TransformComponent)
	trB := transformB.(*ecs.TransformComponent)

	phA := physicsA.(*ecs.PhysicsComponent)
	phB := physicsB.(*ecs.PhysicsComponent)

	// Если оба объекта статические, ничего не делаем
	if phA.Static && phB.Static {
		return
	}

	// Простейшее разрешение столкновения - упругое отражение
	// Вычисляем направление отталкивания
	normal := trA.Position.Sub(trB.Position).Normalize()

	// Вычисляем импульс
	totalMass := phA.Mass + phB.Mass
	relativeVelocity := phA.Velocity.Sub(phB.Velocity)

	// Коэффициент восстановления (упругость)
	restitution := math.Min(phA.Restitution, phB.Restitution)

	// Импульс
	impulse := normal.Multiply(
		-(1 + restitution) * relativeVelocity.Dot(normal) / totalMass,
	)

	// Применяем импульс, если объекты не статические
	if !phA.Static {
		phA.Velocity = phA.Velocity.Add(impulse.Multiply(phB.Mass))
	}

	if !phB.Static {
		phB.Velocity = phB.Velocity.Sub(impulse.Multiply(phA.Mass))
	}
}

// RequiredComponents возвращает компоненты, необходимые для системы ИИ
func (as *AISystem) RequiredComponents() []ecs.ComponentID {
	return []ecs.ComponentID{
		ecs.AIComponentID,
		ecs.TransformComponentID,
	}
}

// Update обновляет поведение ИИ
func (as *AISystem) Update(deltaTime float64) {
	// Получаем все сущности с компонентом ИИ
	entities := as.world.GetEntitiesWithComponent(ecs.AIComponentID)

	for _, entity := range entities {
		// Получаем компоненты ИИ и трансформации
		aiComp, _ := entity.GetComponent(ecs.AIComponentID)
		transformComp, _ := entity.GetComponent(ecs.TransformComponentID)

		ai := aiComp.(*ecs.AIComponent)
		transform := transformComp.(*ecs.TransformComponent)

		// Обновляем поведение в зависимости от текущего состояния
		switch ai.CurrentState {
		case "idle":
			as.handleIdleState(ai, transform, deltaTime)
		case "patrol":
			as.handlePatrolState(ai, transform, deltaTime)
		case "chase":
			as.handleChaseState(ai, transform, deltaTime)
		case "attack":
			as.handleAttackState(ai, transform, deltaTime)
		case "flee":
			as.handleFleeState(ai, transform, deltaTime)
		}
	}
}

// Методы обработки состояний ИИ
func (as *AISystem) handleIdleState(ai *ecs.AIComponent, transform *ecs.TransformComponent, deltaTime float64) {
	// Проверяем наличие целей в зоне обнаружения
	target := as.findNearestTarget(ai, transform)

	if target != nil {
		// Переключаем состояние в зависимости от типа ИИ
		switch ai.AIType {
		case "aggressive":
			ai.SetState("chase")
		case "scared":
			ai.SetState("flee")
		case "neutral":
			// Нейтральные существа могут просто наблюдать
			ai.CurrentState = "patrol"
		}
	} else {
		// Если цели нет, периодически меняем направление
		if rand.Float64() < 0.1*deltaTime {
			as.chooseRandomPatrolPoint(ai)
		}
	}
}

func (as *AISystem) handlePatrolState(ai *ecs.AIComponent, transform *ecs.TransformComponent, deltaTime float64) {
	// Движение к следующей точке патрулирования
	nextPoint := ai.GetNextPatrolPoint()

	// Направление к точке
	direction := nextPoint.Sub(transform.Position).Normalize()

	// Движение в направлении точки
	transform.Position = transform.Position.Add(
		direction.Multiply(ai.DetectionRange * deltaTime),
	)

	// Проверяем, достигнута ли точка
	if transform.Position.Distance(nextPoint) < 0.1 {
		ai.CurrentPatrolIdx = (ai.CurrentPatrolIdx + 1) % len(ai.PatrolPoints)
	}
}

func (as *AISystem) handleChaseState(ai *ecs.AIComponent, transform *ecs.TransformComponent, deltaTime float64) {
	// Ищем текущую цель
	target := as.findNearestTarget(ai, transform)

	if target == nil {
		// Цель потеряна, возвращаемся в состояние патрулирования
		ai.SetState("patrol")
		return
	}

	// Направление к цели
	direction := target.Sub(transform.Position).Normalize()

	// Движение к цели
	transform.Position = transform.Position.Add(
		direction.Multiply(ai.DetectionRange * deltaTime),
	)

	// Если цель в зоне атаки, атакуем
	if transform.Position.Distance(target) <= ai.AttackRange {
		ai.SetState("attack")
	}
}

// Исправьте типы и преобразования
func (as *AISystem) handleAttackState(ai *ecs.AIComponent, transform *ecs.TransformComponent, deltaTime float64) {
	// Проверяем, можем ли атаковать
	currentTime := float64(time.Now().Unix())
	if ai.CanAttack(currentTime) {
		damage := ai.Attack(currentTime)

		// TODO: Применение урона к цели
		// здесь должна быть логика нанесения урона целевой сущности
		_ = damage // Временно игнорируем неиспользуемую переменную
	}
}

func (as *AISystem) findNearestTarget(ai *ecs.AIComponent, transform *ecs.TransformComponent) *ecs.Vector3 {
	// TODO: Реализовать поиск ближайшей цели в радиусе обнаружения
	return nil
}

func (as *AISystem) findNearestThreat(ai *ecs.AIComponent, transform *ecs.TransformComponent) *ecs.Vector3 {
	// TODO: Реализовать поиск ближайшей угрозы в радиусе обнаружения
	return nil
}

func (as *AISystem) handleFleeState(ai *ecs.AIComponent, transform *ecs.TransformComponent, deltaTime float64) {
	// Ищем ближайшую угрозу
	threat := as.findNearestThreat(ai, transform)

	if threat == nil {
		// Угроза исчезла, возвращаемся в состояние патрулирования
		ai.SetState("patrol")
		return
	}

	// Направление от угрозы
	direction := transform.Position.Sub(threat).Normalize()

	// Быстрое движение от угрозы
	transform.Position = transform.Position.Add(
		direction.Multiply(ai.DetectionRange * 1.5 * deltaTime),
	)
}

// Вспомогательные методы для поиска целей и угроз
func (as *AISystem) findNearestTarget(ai *ecs.AIComponent, transform *ecs.TransformComponent) *ecs.Vector3 {
	// TODO: Реализовать поиск ближайшей цели в радиусе обнаружения
	return nil
}

func (as *AISystem) findNearestThreat(ai *ecs.AIComponent, transform *ecs.TransformComponent) *ecs.Vector3 {
	// TODO: Реализовать поиск ближайшей угрозы в радиусе обнаружения
	return nil
}

func (as *AISystem) chooseRandomPatrolPoint(ai *ecs.AIComponent) {
	// Если нет точек патрулирования, создаем случайные
	if len(ai.PatrolPoints) == 0 {
		for i := 0; i < 3; i++ {
			ai.AddPatrolPoint(ecs.Vector3{
				X: rand.Float64()*100 - 50,
				Y: 0,
				Z: rand.Float64()*100 - 50,
			})
		}
	}
}
