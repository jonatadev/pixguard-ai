package rag

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"
)

const rrfK = 60.0

// Embedder generates L2-normalized dense vectors for chunks and queries.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// DeterministicEmbedder produces fast offline L2-normalized character n-gram embeddings.
type DeterministicEmbedder struct {
	Dims int
}

// Embed hashes normalized tokens and character trigrams into a unit vector.
func (e DeterministicEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	dims := e.Dims
	if dims <= 0 {
		dims = 256
	}
	vec := make([]float32, dims)
	tokens := tokenizePTBR(text)
	for _, tok := range tokens {
		addFeature(vec, tok, 2.0)
		runes := []rune(tok)
		for i := 0; i+3 <= len(runes); i++ {
			addFeature(vec, string(runes[i:i+3]), 1.0)
		}
	}
	normalizeL2(vec)
	return vec, nil
}

func addFeature(vec []float32, feat string, weight float32) {
	var h uint32 = 2166136261
	for i := 0; i < len(feat); i++ {
		h ^= uint32(feat[i])
		h *= 16777619
	}
	idx := int(h % uint32(len(vec)))
	vec[idx] += weight
}

func normalizeL2(vec []float32) {
	var sum float64
	for _, v := range vec {
		sum += float64(v) * float64(v)
	}
	if sum == 0 {
		return
	}
	inv := float32(1.0 / math.Sqrt(sum))
	for i := range vec {
		vec[i] *= inv
	}
}

type indexedChunk struct {
	chunk     Chunk
	embedding []float32
	termFreq  map[string]int
	length    int
}

// Index holds regulatory chunks with both lexical (BM25-like) and vector representations.
type Index struct {
	items    []indexedChunk
	docFreq  map[string]int
	avgLen   float64
	embedder Embedder
}

// NewIndex builds a hybrid search index over regulatory chunks.
func NewIndex(ctx context.Context, chunks []Chunk, embedder Embedder) (*Index, error) {
	if len(chunks) == 0 {
		return nil, fmt.Errorf("cannot build rag index with zero chunks")
	}
	if embedder == nil {
		embedder = DeterministicEmbedder{Dims: 256}
	}

	idx := &Index{
		items:    make([]indexedChunk, 0, len(chunks)),
		docFreq:  make(map[string]int),
		embedder: embedder,
	}

	totalLen := 0
	for _, ch := range chunks {
		emb, err := embedder.Embed(ctx, ch.HeadingPath+"\n"+ch.Content)
		if err != nil {
			return nil, fmt.Errorf("embed chunk %s: %w", ch.ID, err)
		}

		tokens := tokenizePTBR(ch.HeadingPath + " " + ch.Content)
		tf := make(map[string]int, len(tokens))
		for _, t := range tokens {
			tf[t]++
		}
		for term := range tf {
			idx.docFreq[term]++
		}

		totalLen += len(tokens)
		idx.items = append(idx.items, indexedChunk{
			chunk:     ch,
			embedding: emb,
			termFreq:  tf,
			length:    len(tokens),
		})
	}
	idx.avgLen = float64(totalLen) / float64(len(chunks))
	if idx.avgLen == 0 {
		idx.avgLen = 1
	}
	return idx, nil
}

// Search retrieves Top-K chunks combining BM25 lexical scores and cosine similarity via RRF (k=60).
func (idx *Index) Search(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	if idx == nil || len(idx.items) == 0 {
		return nil, fmt.Errorf("rag index is empty")
	}
	if topK <= 0 {
		topK = 4
	}

	qVec, err := idx.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	qTokens := tokenizePTBR(query)

	type scored struct {
		pos   int
		score float64
	}

	n := len(idx.items)
	lexScores := make([]scored, n)
	vecScores := make([]scored, n)

	for i, item := range idx.items {
		lexScores[i] = scored{pos: i, score: idx.bm25(item, qTokens)}
		vecScores[i] = scored{pos: i, score: cosine(qVec, item.embedding)}
	}

	sort.Slice(lexScores, func(i, j int) bool {
		if lexScores[i].score == lexScores[j].score {
			return lexScores[i].pos < lexScores[j].pos
		}
		return lexScores[i].score > lexScores[j].score
	})
	sort.Slice(vecScores, func(i, j int) bool {
		if vecScores[i].score == vecScores[j].score {
			return vecScores[i].pos < vecScores[j].pos
		}
		return vecScores[i].score > vecScores[j].score
	})

	lexRank := make([]int, n)
	vecRank := make([]int, n)
	for rank := 0; rank < n; rank++ {
		lexRank[lexScores[rank].pos] = rank + 1
		vecRank[vecScores[rank].pos] = rank + 1
	}

	results := make([]SearchResult, n)
	for i, item := range idx.items {
		lr := lexRank[i]
		vr := vecRank[i]
		rrf := 1.0/(rrfK+float64(lr)) + 1.0/(rrfK+float64(vr))
		results[i] = SearchResult{
			Chunk:       item.chunk,
			LexicalRank: lr,
			VectorRank:  vr,
			RRFScore:    rrf,
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].RRFScore == results[j].RRFScore {
			return results[i].Chunk.ID < results[j].Chunk.ID
		}
		return results[i].RRFScore > results[j].RRFScore
	})

	if topK > len(results) {
		topK = len(results)
	}
	return results[:topK], nil
}

func (idx *Index) bm25(item indexedChunk, qTokens []string) float64 {
	const k1 = 1.5
	const b = 0.75
	nDocs := float64(len(idx.items))
	var score float64

	for _, qt := range qTokens {
		tf := float64(item.termFreq[qt])
		if tf == 0 {
			continue
		}
		df := float64(idx.docFreq[qt])
		idf := math.Log(1.0 + (nDocs-df+0.5)/(df+0.5))
		denom := tf + k1*(1.0-b+b*(float64(item.length)/idx.avgLen))
		score += idf * ((tf * (k1 + 1.0)) / denom)
	}
	return score
}

func cosine(a, b []float32) float64 {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	var dot float64
	for i := 0; i < limit; i++ {
		dot += float64(a[i]) * float64(b[i])
	}
	return dot
}

func tokenizePTBR(input string) []string {
	cleaned := foldAccents(strings.ToLower(input))
	var out []string
	var cur strings.Builder

	flush := func() {
		if cur.Len() > 1 {
			out = append(out, cur.String())
		}
		cur.Reset()
	}

	for _, r := range cleaned {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}

func foldAccents(s string) string {
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a",
		"é", "e", "ê", "e",
		"í", "i",
		"ó", "o", "ô", "o", "õ", "o",
		"ú", "u", "ü", "u",
		"ç", "c",
	)
	return replacer.Replace(s)
}
