package player

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"echo-taiga/internal/engine/ecs"
	"echo-taiga/internal/world"
)

// Player представляет игрока в игре
type Player struct {
	entity    *ecs.Entity
	transform *ecs.TransformComponent
	physics   *ecs.PhysicsComponent
	control   *ecs.PlayerControlComponent
	survival  *ecs.SurvivalComponent
	inventory *ecs.InventoryComponent
}

// CreatePlayerEntity создает сущность игрока в мире
func CreatePlayerEntity(world *ecs.World, gameWorld *world.World) (*Player, error) {
	// Создаем новую сущность игрока
	playerEntity := ecs.NewEntity()

	// Позиция спавна
	startPosition := ecs.Vector3{
		X: 0,
		Y: 1,
		Z: 0,
	}

	// Компонент трансформации
	transformComp := ecs.NewTransformComponent(startPosition)
	playerEntity.AddComponent(transformComp)

	// Физический компонент
	physicsComp := ecs.NewPhysicsComponent(70, false)
	physicsComp.Collider = "capsule"
	physicsComp.ColliderSize = ecs.Vector3{X: 0.5, Y: 1.8, Z: 0.5}
	playerEntity.AddComponent(physicsComp)

	// Компонент управления
	controlComp := ecs.NewPlayerControlComponent()
	playerEntity.AddComponent(controlComp)

	// Компонент выживания
	survivalComp := ecs.NewSurvivalComponent()
	playerEntity.AddComponent(survivalComp)

	// Компонент инвентаря
	inventoryComp := ecs.NewInventoryComponent(24, 50.0)
	playerEntity.AddComponent(inventoryComp)

	// Добавляем теги
	playerEntity.AddTag("player")
	playerEntity.AddTag("living")

	// Добавляем сущность в мир
	world.AddEntity(playerEntity)

	// Устанавливаем позицию игрока в мире
	gameWorld.SetPlayerPosition(startPosition)

	// Создаем структуру Player
	player := &Player{
		entity:    playerEntity,
		transform: transformComp,
		physics:   physicsComp,
		control:   controlComp,
		survival:  survivalComp,
		inventory: inventoryComp,
	}

	return player, nil
}

// Update обновляет состояние игрока
func (p *Player) Update(deltaTime float64, world *world.World) {
	// Обработка ввода
	p.handleInput(deltaTime)

	// Обновление состояния выживания
	p.survival.Update(deltaTime)

	// Обновление физики
	p.updatePhysics(deltaTime)

	// Обновление позиции мира
	world.SetPlayerPosition(p.transform.Position)
}

// handleInput обрабатывает ввод пользователя
func (p *Player) handleInput(deltaTime float64) {
	// Получаем текущую скорость движения
	speed := p.control.GetCurrentMovementSpeed()

	// Проверяем нажатие клавиш
	moveX, moveZ := 0.0, 0.0

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		moveZ -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		moveZ += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		moveX -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		moveX += 1
	}

	// Бег при зажатом Shift
	if ebiten.IsKeyPressed(ebiten.KeyShift) && p.control.CanSprint {
		speed = p.control.GetCurrentSprintSpeed()
	}

	// Приседание при зажатом Ctrl
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		p.control.IsCrouching = true
		speed *= p.control.CrouchMultiplier
	} else {
		p.control.IsCrouching = false
	}

	// Нормализация движения по диагонали
	if moveX != 0 && moveZ != 0 {
		moveX *= 0.7071 // 1 / sqrt(2)
		moveZ *= 0.7071
	}

	// Применяем движение с учетом направления камеры
	moveVector := ecs.Vector3{
		X: moveX * speed * deltaTime,
		Y: 0,
		Z: moveZ * speed * deltaTime,
	}

	// Обновляем позицию
	p.transform.Position = p.transform.Position.Add(moveVector)

	// Обработка прыжка
	if ebiten.IsKeyPressed(ebiten.KeySpace) && p.control.CanJump && p.control.IsGrounded {
		p.physics.Velocity.Y = p.control.JumpForce
		p.control.IsGrounded = false
	}

	// Обработка взаимодействия
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		p.control.IsInteracting = true
		p.control.LastInteractTime = float64(time.Now().UnixNano()) / 1e9
	} else {
		p.control.IsInteracting = false
	}
}

// updatePhysics обновляет физику игрока
func (p *Player) updatePhysics(deltaTime float64) {
	// Применяем гравитацию
	p.physics.Velocity.Y -= 9.8 * p.physics.Gravity * deltaTime

	// Обновляем позицию
	p.transform.Position = p.transform.Position.Add(
		p.physics.Velocity.Multiply(deltaTime),
	)

	// Проверка столкновений и земли будет реализована в системе физики
	// Временно считаем, что игрок всегда на земле
	p.control.IsGrounded = p.transform.Position.Y <= 1.0

	// Если игрок ниже уровня земли, возвращаем на поверхность
	if p.transform.Position.Y < 1.0 {
		p.transform.Position.Y = 1.0
		p.physics.Velocity.Y = 0
	}
}

// GetPosition возвращает текущую позицию игрока
func (p *Player) GetPosition() ecs.Vector3 {
	return p.transform.Position
}

// GetEntity возвращает сущность игрока
func (p *Player) GetEntity() *ecs.Entity {
	return p.entity
}
