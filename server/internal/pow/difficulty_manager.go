package pow

import (
	"sync/atomic"
	"time"
	"word-of-wisdom-server/internal/config"
	"word-of-wisdom-server/internal/interfaces"
	"word-of-wisdom-server/internal/logger"
)

// DifficultyManager управляет динамическим изменением сложности PoW
type DifficultyManager struct {
	difficulty int32            // Текущее значение сложности
	config     config.PoWConfig // Конфигурация PoW
}

func NewDifficultyManager(config config.PoWConfig) *DifficultyManager {
	return &DifficultyManager{
		difficulty: int32(config.BaseDifficulty),
		config:     config,
	}
}

var _ interfaces.DifficultyManager = (*DifficultyManager)(nil)

func (dm *DifficultyManager) GetDifficulty() int {
	return int(atomic.LoadInt32(&dm.difficulty))
}

func (dm *DifficultyManager) SetDifficulty(difficulty int) {
	atomic.StoreInt32(&dm.difficulty, int32(difficulty))
}

// AdjustDifficulty корректирует сложность на основе времени решения клиентом
func (dm *DifficultyManager) AdjustDifficulty(solveTime time.Duration) {
	currentDifficulty := dm.GetDifficulty()
	logger.Log.Info().
		Int32("current_difficulty", currentDifficulty).
		Dur("solve_time", solveTime).
		Msg("Корректировка сложности")

	for {
		current := atomic.LoadInt32(&dm.difficulty)
		var next int32
		switch {
		case solveTime < dm.config.MinSolveTime && current < int32(dm.config.MaxDifficulty):
			next = current + 1
			logger.Log.Info().
				Int32("current_difficulty", current).
				Int32("next_difficulty", next).
				Dur("solve_time", solveTime).
				Msg("Время решения слишком быстрое. Повышение сложности")
		case solveTime > dm.config.MaxSolveTime && current > int32(dm.config.MinDifficulty):
			next = current - 1
			logger.Log.Info().
				Int32("current_difficulty", current).
				Int32("next_difficulty", next).
				Dur("solve_time", solveTime).
				Msg("Время решения слишком долгое. Понижение сложности")
		default:
			logger.Log.Info().
				Int32("current_difficulty", current).
				Msg("Сложность остается без изменений")
			return
		}
		if atomic.CompareAndSwapInt32(&dm.difficulty, current, next) {
			return
		}
	}
}
