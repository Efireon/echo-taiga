package main

import (
	"fmt"
	"log"
	"os"

	"echo-taiga/internal/config"
	"echo-taiga/internal/core"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Не удалось загрузить конфигурацию: %v. Используется конфигурация по умолчанию.", err)
		cfg = config.DefaultConfig()
	}

	// Создаем игру
	game, err := core.NewGame(cfg)
	if err != nil {
		fmt.Printf("Ошибка создания игры: %v\n", err)
		os.Exit(1)
	}

	// Запускаем игру
	if err := game.Run(); err != nil {
		fmt.Printf("Ошибка во время игры: %v\n", err)
		os.Exit(1)
	}
}
