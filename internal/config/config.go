package config

// Config содержит основные настройки игры
type Config struct {
	// Общие настройки
	WindowWidth  int    `json:"window_width"`
	WindowHeight int    `json:"window_height"`
	Fullscreen   bool   `json:"fullscreen"`
	Title        string `json:"title"`
	
	// Настройки игры
	Seed              int64   `json:"seed"`
	Difficulty        string  `json:"difficulty"`
	WorldSize         string  `json:"world_size"`
	MetamorphosisRate float64 `json:"metamorphosis_rate"`
	
	// Настройки производительности
	TargetFPS      int  `json:"target_fps"`
	EnableVSync    bool `json:"enable_vsync"`
	ChunkSize      int  `json:"chunk_size"`
	ViewDistance   int  `json:"view_distance"`
	EnableShadows  bool `json:"enable_shadows"`
	TextureQuality int  `json:"texture_quality"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		WindowWidth:      800,
		WindowHeight:     600,
		Fullscreen:       false,
		Title:            "Эхо Тайги",
		Seed:             0, // 0 означает случайный сид
		Difficulty:       "normal",
		WorldSize:        "medium",
		MetamorphosisRate: 1.0,
		TargetFPS:        60,
		EnableVSync:      true,
		ChunkSize:        64,
		ViewDistance:     5,
		EnableShadows:    true,
		TextureQuality:   1,
	}
}

// Load загружает конфигурацию из файла
func Load() (*Config, error) {
	// TODO: реализовать загрузку из файла
	return DefaultConfig(), nil
}

// Save сохраняет конфигурацию в файл
func (c *Config) Save() error {
	// TODO: реализовать сохранение в файл
	return nil
}
