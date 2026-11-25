package rag

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestChunkMarkdownPreservesTablesAndHeadings(t *testing.T) {
	md := `# Manual Teste

## 1. Prazos
| Regra | Prazo |
| :--- | :--- |
| Janela MED | 80 dias |
| Analise | 7 dias |

## 2. Desacordo
Desacordo comercial nao e elegivel ao MED (Art. 41-C).
`
	chunks := ChunkMarkdown("manual.md", md)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if !strings.Contains(chunks[0].Content, "| Janela MED | 80 dias |") || !strings.Contains(chunks[0].Content, "| Analise | 7 dias |") {
		t.Fatalf("expected markdown table to remain intact in chunk 0: %q", chunks[0].Content)
	}
	if chunks[0].HeadingPath != "Manual Teste > 1. Prazos" {
		t.Fatalf("unexpected heading path: %q", chunks[0].HeadingPath)
	}
}

func TestHybridRAGRetrievesRegulatoryRules(t *testing.T) {
	corpusDir := filepath.Join("..", "..", "..", "docs", "regulatory")
	chunks, err := LoadCorpusFromDir(corpusDir)
	if err != nil {
		t.Fatalf("LoadCorpusFromDir failed: %v", err)
	}
	if len(chunks) < 6 {
		t.Fatalf("expected at least 6 chunks across 3 regulatory docs, got %d", len(chunks))
	}

	idx, err := NewIndex(context.Background(), chunks, DeterministicEmbedder{Dims: 256})
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	tests := []struct {
		name           string
		query          string
		expectedDoc    string
		expectedPhrase string
	}{
		{
			name:           "Desacordo comercial Art. 41-C",
			query:          "desacordo comercial produto nao entregue Art. 41-C",
			expectedDoc:    "01-manual-pix-med-bacen.md",
			expectedPhrase: "Desacordo Comercial",
		},
		{
			name:           "Janela legal 80 dias MED",
			query:          "prazo maximo janela 80 dias devolucao MED",
			expectedDoc:    "01-manual-pix-med-bacen.md",
			expectedPhrase: "80 dias",
		},
		{
			name:           "Bloqueio Cautelar 72 horas",
			query:          "bloqueio cautelar preventivo 72 horas conta recebedora recem-criada",
			expectedDoc:    "02-bloqueio-cautelar-fraude.md",
			expectedPhrase: "72 horas",
		},
		{
			name:           "Sinalizacao PLD-FT COAF",
			query:          "PLD-FT COAF conta laranja R$ 5.000,00 alçada humana",
			expectedDoc:    "03-politica-interna-antifraude-lgpd.md",
			expectedPhrase: "COAF",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results, err := idx.Search(context.Background(), tc.query, 3)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			if len(results) != 3 {
				t.Fatalf("expected 3 results, got %d", len(results))
			}

			found := false
			for _, r := range results {
				if r.Chunk.DocID == tc.expectedDoc && strings.Contains(r.Chunk.Content, tc.expectedPhrase) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected Top-3 results for %q to contain doc=%s phrase=%q; got %+v",
					tc.query, tc.expectedDoc, tc.expectedPhrase, results)
			}
		})
	}
}
