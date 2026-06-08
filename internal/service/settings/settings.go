// Package settings учитывает окна calc_period всех активных gRPC-клиентов.
package settings

import "sync"

// Service отслеживает максимальное окно calc_period среди подключённых клиентов.
type Service struct {
	mu      sync.Mutex
	periods map[uint32]int
}

// New создаёт пустой реестр периодов.
func New() *Service {
	return &Service{periods: make(map[uint32]int)}
}

// Add регистрирует calc_period клиента.
func (s *Service) Add(period uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.periods[period]++
}

// Remove снимает регистрацию; возвращает true, если больше нет клиентов с этим периодом.
func (s *Service) Remove(period uint32) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.periods[period]--
	if s.periods[period] <= 0 {
		delete(s.periods, period)
	}
	return len(s.periods) == 0
}

// Max возвращает максимальный calc_period среди активных клиентов.
func (s *Service) Max() (uint32, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.periods) == 0 {
		return 0, false
	}
	var maxPeriod uint32
	for p := range s.periods {
		if p > maxPeriod {
			maxPeriod = p
		}
	}
	return maxPeriod, true
}
