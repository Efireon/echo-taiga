package audio

import (
	"log"
)

// Manager отвечает за аудио в игре
type Manager struct {
	// TODO: Добавить поля для работы с аудио
	isMuted bool
	volume  float64
}

// NewManager создает новый аудио менеджер
func NewManager() *Manager {
	return &Manager{
		isMuted: false,
		volume:  0.5, // Средняя громкость по умолчанию
	}
}

// PlayMusic начинает проигрывание музыкального трека
func (m *Manager) PlayMusic(trackName string) error {
	if m.isMuted {
		return nil
	}
	log.Printf("Проигрывание музыкального трека: %s", trackName)
	// TODO: Реальная логика проигрывания музыки
	return nil
}

// PlaySound воспроизводит звуковой эффект
func (m *Manager) PlaySound(soundName string) error {
	if m.isMuted {
		return nil
	}
	log.Printf("Проигрывание звукового эффекта: %s", soundName)
	// TODO: Реальная логика проигрывания звука
	return nil
}

// SetVolume устанавливает общую громкость
func (m *Manager) SetVolume(volume float64) {
	m.volume = volume
	// TODO: Применение настроек громкости
}

// Mute выключает звук
func (m *Manager) Mute() {
	m.isMuted = true
}

// Unmute включает звук
func (m *Manager) Unmute() {
	m.isMuted = false
}
