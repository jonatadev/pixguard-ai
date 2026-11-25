package store

import (
	"sync"
	"testing"

	"github.com/jonatadev/pixguard-ai/internal/domain"
)

func TestMemoryStoreSeedData(t *testing.T) {
	s := NewMemoryStore()
	cases := s.List("", "")

	if len(cases) != 5 {
		t.Fatalf("expected 5 seeded Pix cases, got %d", len(cases))
	}
	for _, c := range cases {
		if c.ID == "" || c.AmountBRL <= 0 || c.RawDescription == "" {
			t.Fatalf("invalid seeded case: %+v", c)
		}
	}
}

func TestMemoryStoreListFilters(t *testing.T) {
	s := NewMemoryStore()

	tests := []struct {
		name          string
		risk          string
		status        string
		expectedCount int
	}{
		{name: "no filter", risk: "", status: "", expectedCount: 5},
		{name: "high risk uppercase", risk: "ALTO", status: "", expectedCount: 3},
		{name: "high risk trimmed lowercase", risk: "  alto  ", status: "", expectedCount: 3},
		{name: "medium risk", risk: "MEDIO", status: "", expectedCount: 1},
		{name: "low risk", risk: "BAIXO", status: "", expectedCount: 1},
		{name: "status em_analise", risk: "", status: "em_analise", expectedCount: 1},
		{name: "high risk and pending", risk: "ALTO", status: "PENDENTE", expectedCount: 2},
		{name: "nonexistent risk", risk: "CRITICO", status: "", expectedCount: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := s.List(tc.risk, tc.status)
			if len(got) != tc.expectedCount {
				t.Fatalf("expected %d cases, got %d", tc.expectedCount, len(got))
			}
		})
	}
}

func TestMemoryStoreGetByID(t *testing.T) {
	s := NewMemoryStore()

	c, ok := s.GetByID("pix-2026-0001")
	if !ok {
		t.Fatalf("expected case-insensitive lookup for pix-2026-0001 to succeed")
	}
	if c.RiskLevel != domain.RiskAlto || !c.MedEligibleExpected {
		t.Fatalf("unexpected PIX-2026-0001 attributes: %+v", c)
	}

	expired, ok := s.GetByID("PIX-2026-0005")
	if !ok {
		t.Fatalf("expected PIX-2026-0005 to exist")
	}
	if expired.DaysSinceTx != 94 || expired.MedEligibleExpected {
		t.Fatalf("unexpected PIX-2026-0005 attributes: %+v", expired)
	}

	if _, ok := s.GetByID("PIX-9999-0000"); ok {
		t.Fatalf("expected unknown ID to return ok=false")
	}
}

func TestMemoryStoreConcurrentReads(t *testing.T) {
	s := NewMemoryStore()

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if len(s.List("ALTO", "")) != 3 {
				t.Errorf("unexpected concurrent list count")
			}
			if _, ok := s.GetByID("PIX-2026-0004"); !ok {
				t.Errorf("expected PIX-2026-0004 during concurrent read")
			}
		}()
	}
	wg.Wait()
}

func TestMemoryStoreUpdateStatusAndFeedback(t *testing.T) {
	s := NewMemoryStore()

	if !s.UpdateStatus("pix-2026-0001", domain.StatusAprovado) {
		t.Fatalf("expected UpdateStatus to succeed for existing case")
	}
	updated, ok := s.GetByID("PIX-2026-0001")
	if !ok || updated.Status != domain.StatusAprovado {
		t.Fatalf("expected updated status APROVADO, got %+v", updated)
	}
	if s.UpdateStatus("PIX-9999-0000", domain.StatusRejeitado) {
		t.Fatalf("expected UpdateStatus to return false for unknown case")
	}

	s.SaveFeedback(domain.AnalystFeedbackRecord{
		CaseID:   "PIX-2026-0001",
		Decision: domain.DecisionAprovado,
		Notes:    "MED confirmado pela mesa",
	})
	s.SaveFeedback(domain.AnalystFeedbackRecord{
		CaseID:   "PIX-2026-0002",
		Decision: domain.DecisionEditado,
	})

	list := s.ListFeedback("pix-2026-0001")
	if len(list) != 1 || list[0].Decision != domain.DecisionAprovado {
		t.Fatalf("expected 1 feedback record for PIX-2026-0001, got %+v", list)
	}
	if empty := s.ListFeedback("PIX-9999-0000"); len(empty) != 0 {
		t.Fatalf("expected empty feedback list for unknown case, got %d", len(empty))
	}
}
