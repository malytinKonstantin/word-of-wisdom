package storage

import (
	"math/rand"
	"sync"
	"sync/atomic"
)

// QuoteStorage предоставляет доступ к хранению и получению цитат
type QuoteStorage struct {
	quotes    []string     // Список цитат
	quoteMux  sync.RWMutex // Мьютекс для безопасного доступа
	randIndex int32        // Индекс для получения случайной цитаты
}

func New() *QuoteStorage {
	qs := &QuoteStorage{
		quotes: []string{
			"Знание — сила. — Фрэнсис Бэкон",
			"Чтобы увидеть радугу, нужно пережить дождь.",
			"Свобода — это свобода мыслить. — Альберт Эйнштейн",
			"Единственный способ делать великую работу — любить то, что вы делаете. — Стив Джобс",
			"Успех — это способность шагать от одной неудачи к другой, не теряя энтузиазма. — Уинстон Черчилль",
			"Самый тёмный час наступает перед рассветом. — Томас Фуллер",
			"Никогда не поздно быть тем, кем ты мог бы быть. — Джордж Элиот",
			"Каждый человек должен превосходить себя самого. — Николай Гоголь",
			"В жизни нет ничего невозможного, если вы не боитесь пробовать. — Сакити Тоёда",
		},
	}
	// Перемешиваем список цитат
	rand.Shuffle(len(qs.quotes), func(i, j int) {
		qs.quotes[i], qs.quotes[j] = qs.quotes[j], qs.quotes[i]
	})
	return qs
}

func (qs *QuoteStorage) GetRandomQuote() string {
	qs.quoteMux.RLock()
	defer qs.quoteMux.RUnlock()
	index := atomic.AddInt32(&qs.randIndex, 1) % int32(len(qs.quotes))
	return qs.quotes[index]
}

func (qs *QuoteStorage) GetAllQuotes() []string {
	qs.quoteMux.RLock()
	defer qs.quoteMux.RUnlock()
	return qs.quotes
}
