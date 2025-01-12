package pow

import (
	"sync/atomic"
	"time"
	"word-of-wisdom-server/internal/config"
)

type DifficultyManager struct {
	difficulty int32
	config     *config.Config
}

func NewDifficultyManager(config *config.Config) *DifficultyManager {
	return &DifficultyManager{
		difficulty: int32(config.BaseDifficulty),
		config:     config,
	}
}

func (dm *DifficultyManager) GetDifficulty() int {
	return int(atomic.LoadInt32(&dm.difficulty))
}

func (dm *DifficultyManager) SetDifficulty(difficulty int) {
	atomic.StoreInt32(&dm.difficulty, int32(difficulty))
}

func (dm *DifficultyManager) AdjustDifficulty(solveTime time.Duration) {
	for {
		current := atomic.LoadInt32(&dm.difficulty)
		var next int32
		switch {
		case solveTime < dm.config.MinSolveTime && current < int32(dm.config.MaxDifficulty):
			next = current + 1
		case solveTime > dm.config.MaxSolveTime && current > int32(dm.config.MinDifficulty):
			next = current - 1
		default:
			return
		}
		if atomic.CompareAndSwapInt32(&dm.difficulty, current, next) {
			return
		}
	}
}
