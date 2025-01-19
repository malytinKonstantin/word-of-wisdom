package pow

import (
	"log"
	"sync/atomic"
	"time"
	"word-of-wisdom-server/internal/config"
	"word-of-wisdom-server/internal/interfaces"
)

// DifficultyManager управляет динамическим изменением сложности PoW
type DifficultyManager struct {
	difficulty int32          // Текущее значение сложности
	config     *config.Config // Конфигурация сервера
}

func NewDifficultyManager(config *config.Config) *DifficultyManager {
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
	currentDifficulty := atomic.LoadInt32(&dm.difficulty)
	log.Printf("Корректировка сложности. Текущая: %d, Время решения: %v", currentDifficulty, solveTime)

	for {
		current := atomic.LoadInt32(&dm.difficulty)
		var next int32
		switch {
		case solveTime < dm.config.MinSolveTime && current < int32(dm.config.MaxDifficulty):
			next = current + 1
			log.Printf("Время решения слишком быстрое (%v < %v). Повышение сложности до %d",
				solveTime, dm.config.MinSolveTime, next)
		case solveTime > dm.config.MaxSolveTime && current > int32(dm.config.MinDifficulty):
			next = current - 1
			log.Printf("Время решения слишком долгое (%v > %v). Понижение сложности до %d",
				solveTime, dm.config.MaxSolveTime, next)
		default:
			log.Printf("Сложность остается без изменений: %d", current)
			return
		}
		if atomic.CompareAndSwapInt32(&dm.difficulty, current, next) {
			return
		}
	}
}
