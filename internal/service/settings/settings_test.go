package settings

import "testing"

func TestSettingsMax(t *testing.T) {
	s := New()
	s.Add(10)
	s.Add(20)
	maxPeriod, ok := s.Max()
	if !ok || maxPeriod != 20 {
		t.Fatalf("max: got %d %v", maxPeriod, ok)
	}
	s.Remove(10)
	maxPeriod, ok = s.Max()
	if !ok || maxPeriod != 20 {
		t.Fatalf("after remove: got %d", maxPeriod)
	}
}
