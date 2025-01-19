package config

// PoWConfig содержит настройки Proof of Work
type PoWConfig struct {
	BaseDifficulty int `mapstructure:"base_difficulty"` // Начальная сложность PoW
	MaxDifficulty  int `mapstructure:"max_difficulty"`  // Максимальная сложность PoW
	MinDifficulty  int `mapstructure:"min_difficulty"`  // Минимальная сложность PoW
	MinSolveTime   int `mapstructure:"min_solve_time"`  // Минимальное время решения для регулировки сложности (в секундах)
	MaxSolveTime   int `mapstructure:"max_solve_time"`  // Максимальное время решения для регулировки сложности (в секундах)
}
