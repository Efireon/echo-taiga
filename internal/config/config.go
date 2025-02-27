package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Load загружает конфигурацию из файла
func Load() (*Config, error) {
	// Создаем директорию для конфигурации, если она не существует
	configDir := filepath.Join(os.Getenv("HOME"), ".echo-taiga")
	os.MkdirAll(configDir, os.ModePerm)

	// Настраиваем Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Устанавливаем значения по умолчанию
	config := DefaultConfig()
	viper.SetDefault("window_width", config.WindowWidth)
	viper.SetDefault("window_height", config.WindowHeight)
	viper.SetDefault("fullscreen", config.Fullscreen)
	viper.SetDefault("title", config.Title)
	viper.SetDefault("seed", config.Seed)
	viper.SetDefault("difficulty", config.Difficulty)
	viper.SetDefault("world_size", config.WorldSize)
	viper.SetDefault("metamorphosis_rate", config.MetamorphosisRate)
	viper.SetDefault("target_fps", config.TargetFPS)
	viper.SetDefault("enable_vsync", config.EnableVSync)
	viper.SetDefault("chunk_size", config.ChunkSize)
	viper.SetDefault("view_distance", config.ViewDistance)
	viper.SetDefault("enable_shadows", config.EnableShadows)
	viper.SetDefault("texture_quality", config.TextureQuality)

	// Попытка прочитать существующий конфиг
	if err := viper.ReadInConfig(); err != nil {
		// Если файл не найден, создаем его
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			config := DefaultConfig()
			viper.Set("window_width", config.WindowWidth)
			viper.Set("window_height", config.WindowHeight)
			// ... установка других параметров

			configPath := filepath.Join(configDir, "config.yaml")
			if err := viper.WriteConfigAs(configPath); err != nil {
				return nil, fmt.Errorf("ошибка создания файла конфигурации: %v", err)
			}
		} else {
			return nil, fmt.Errorf("ошибка чтения конфигурации: %v", err)
		}
	}

	// Заполняем структуру конфигурации
	config.WindowWidth = viper.GetInt("window_width")
	config.WindowHeight = viper.GetInt("window_height")
	config.Fullscreen = viper.GetBool("fullscreen")
	config.Title = viper.GetString("title")
	config.Seed = viper.GetInt64("seed")
	config.Difficulty = viper.GetString("difficulty")
	config.WorldSize = viper.GetString("world_size")
	config.MetamorphosisRate = viper.GetFloat64("metamorphosis_rate")
	config.TargetFPS = viper.GetInt("target_fps")
	config.EnableVSync = viper.GetBool("enable_vsync")
	config.ChunkSize = viper.GetInt("chunk_size")
	config.ViewDistance = viper.GetInt("view_distance")
	config.EnableShadows = viper.GetBool("enable_shadows")
	config.TextureQuality = viper.GetInt("texture_quality")

	return config, nil
}

// Save сохраняет конфигурацию в файл
func (c *Config) Save() error {
	configDir := filepath.Join(os.Getenv("HOME"), ".echo-taiga")
	os.MkdirAll(configDir, os.ModePerm)

	viper.Set("window_width", c.WindowWidth)
	viper.Set("window_height", c.WindowHeight)
	viper.Set("fullscreen", c.Fullscreen)
	viper.Set("title", c.Title)
	viper.Set("seed", c.Seed)
	viper.Set("difficulty", c.Difficulty)
	viper.Set("world_size", c.WorldSize)
	viper.Set("metamorphosis_rate", c.MetamorphosisRate)
	viper.Set("target_fps", c.TargetFPS)
	viper.Set("enable_vsync", c.EnableVSync)
	viper.Set("chunk_size", c.ChunkSize)
	viper.Set("view_distance", c.ViewDistance)
	viper.Set("enable_shadows", c.EnableShadows)
	viper.Set("texture_quality", c.TextureQuality)

	configPath := filepath.Join(configDir, "config.yaml")
	return viper.WriteConfigAs(configPath)
}
