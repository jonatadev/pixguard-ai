package rag

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Chunk represents a single structural section of a regulatory Markdown document.
type Chunk struct {
	ID          string `json:"id"`
	DocID       string `json:"docId"`
	HeadingPath string `json:"headingPath"`
	Content     string `json:"content"`
}

// SearchResult pairs a retrieved chunk with its lexical, vector, and RRF scores.
type SearchResult struct {
	Chunk       Chunk   `json:"chunk"`
	LexicalRank int     `json:"lexicalRank"`
	VectorRank  int     `json:"vectorRank"`
	RRFScore    float64 `json:"rrfScore"`
}

// ChunkMarkdown splits Markdown by ## / ### headings while keeping tables intact.
func ChunkMarkdown(docID, content string) []Chunk {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	var docTitle string
	var h2 string
	var h3 string
	var buf []string
	var chunks []Chunk
	seq := 0

	flush := func() {
		body := strings.TrimSpace(strings.Join(buf, "\n"))
		buf = nil
		if body == "" {
			return
		}
		seq++
		path := buildHeadingPath(docTitle, h2, h3)
		chunks = append(chunks, Chunk{
			ID:          fmt.Sprintf("%s#%02d", docID, seq),
			DocID:       docID,
			HeadingPath: path,
			Content:     body,
		})
	}

	inFence := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
		}

		if !inFence {
			switch {
			case strings.HasPrefix(trimmed, "# ") && docTitle == "":
				docTitle = strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
				continue
			case strings.HasPrefix(trimmed, "## "):
				flush()
				h2 = strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
				h3 = ""
				buf = append(buf, line)
				continue
			case strings.HasPrefix(trimmed, "### "):
				flush()
				h3 = strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
				buf = append(buf, line)
				continue
			}
		}

		// Ignore blockquote/hr metadata before first section if no content yet.
		if h2 == "" && h3 == "" && (trimmed == "" || trimmed == "---") {
			continue
		}
		buf = append(buf, line)
	}
	flush()
	return chunks
}

func buildHeadingPath(docTitle, h2, h3 string) string {
	parts := make([]string, 0, 3)
	if docTitle != "" {
		parts = append(parts, docTitle)
	}
	if h2 != "" {
		parts = append(parts, h2)
	}
	if h3 != "" {
		parts = append(parts, h3)
	}
	if len(parts) == 0 {
		return "root"
	}
	return strings.Join(parts, " > ")
}

// LoadCorpusFromDir reads and chunks all .md regulatory files in dir sorted by filename.
func LoadCorpusFromDir(dir string) ([]Chunk, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read regulatory corpus dir %q: %w", dir, err)
	}

	var mdFiles []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}
		mdFiles = append(mdFiles, entry.Name())
	}
	sort.Strings(mdFiles)

	if len(mdFiles) == 0 {
		return nil, fmt.Errorf("no markdown regulatory files found in %q", dir)
	}

	all := make([]Chunk, 0, len(mdFiles)*4)
	for _, name := range mdFiles {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read regulatory file %q: %w", name, err)
		}
		all = append(all, ChunkMarkdown(name, string(raw))...)
	}
	return all, nil
}
