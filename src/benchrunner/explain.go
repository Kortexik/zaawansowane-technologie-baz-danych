package benchrunner

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ExplainPlan captures the raw query plan returned by EXPLAIN for a single
// (database, scenario, with_index) combination. Stored at the largest data
// scale because that's where the planner has the most realistic statistics.
type ExplainPlan struct {
	Database  string `json:"database"`
	Entity    string `json:"entity"`
	WithIndex bool   `json:"with_index"`
	DataScale int    `json:"data_scale"`
	Query     string `json:"query"`
	Plan      string `json:"plan"`
}

// ExplainCollector accumulates plans across all benchmark phases.
type ExplainCollector struct {
	plans []ExplainPlan
}

// NewExplainCollector returns an empty collector.
func NewExplainCollector() *ExplainCollector {
	return &ExplainCollector{}
}

// Add appends one plan.
func (c *ExplainCollector) Add(p ExplainPlan) {
	c.plans = append(c.plans, p)
}

// Plans returns a defensive copy of the accumulated plans.
func (c *ExplainCollector) Plans() []ExplainPlan {
	out := make([]ExplainPlan, len(c.plans))
	copy(out, c.plans)
	return out
}

// SaveJSON writes the raw plan data to the given path. Used by the notebook.
func (c *ExplainCollector) SaveJSON(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(c.plans)
}

// SaveMarkdown writes a human-readable comparison: pairs each (db, entity)
// without-index and with-index plan side by side. The output is structured
// so it can be pasted into the report verbatim.
func (c *ExplainCollector) SaveMarkdown(path string) error {
	// Group by (db, entity), each carrying up to one with_index=false and one
	// with_index=true plan.
	type key struct{ db, entity string }
	grouped := map[key][2]*ExplainPlan{}
	for i := range c.plans {
		p := c.plans[i]
		k := key{p.Database, p.Entity}
		entry := grouped[k]
		idx := 0
		if p.WithIndex {
			idx = 1
		}
		entry[idx] = &p
		grouped[k] = entry
	}

	// Stable ordering: by database then entity.
	keys := make([]key, 0, len(grouped))
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].db != keys[j].db {
			return keys[i].db < keys[j].db
		}
		return keys[i].entity < keys[j].entity
	})

	var b strings.Builder
	b.WriteString("# Analiza planów zapytań (EXPLAIN)\n\n")
	b.WriteString("Plany zebrane dla pierwszego zapytania z każdego pliku SELECT,\n")
	b.WriteString("przy największej dostępnej skali danych. Pokazują wpływ utworzenia\n")
	b.WriteString("indeksu na wybór ścieżki dostępu przez optymalizator.\n\n")
	b.WriteString("---\n\n")

	for _, k := range keys {
		entries := grouped[k]
		b.WriteString(fmt.Sprintf("## %s · `%s`\n\n", strings.ToUpper(k.db), k.entity))

		if entries[0] != nil {
			b.WriteString(fmt.Sprintf("**Zapytanie**: `%s`  \n", entries[0].Query))
			b.WriteString(fmt.Sprintf("**Skala**: %d rows\n\n", entries[0].DataScale))
		} else if entries[1] != nil {
			b.WriteString(fmt.Sprintf("**Zapytanie**: `%s`  \n", entries[1].Query))
			b.WriteString(fmt.Sprintf("**Skala**: %d rows\n\n", entries[1].DataScale))
		}

		b.WriteString("### Bez indeksu\n\n```\n")
		if entries[0] != nil {
			b.WriteString(entries[0].Plan)
		} else {
			b.WriteString("(brak — nie zebrano)\n")
		}
		b.WriteString("```\n\n")

		b.WriteString("### Z indeksem\n\n```\n")
		if entries[1] != nil {
			b.WriteString(entries[1].Plan)
		} else {
			b.WriteString("(brak — nie zebrano)\n")
		}
		b.WriteString("```\n\n")

		b.WriteString("---\n\n")
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}
