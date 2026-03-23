package control

import "sort"

type CombinedAPIReport struct {
	DeclaredTotal     int              `json:"declared_total"`
	UniqueMethodPath  int              `json:"unique_method_path"`
	DuplicateEntries  int              `json:"duplicate_entries"`
	DuplicateKeys     int              `json:"duplicate_keys"`
	TopDuplicatePaths []DuplicateEntry `json:"top_duplicate_paths"`
}

type DuplicateEntry struct {
	Key         string `json:"key"`
	Occurrences int    `json:"occurrences"`
}

// AnalyzeOpenAPIPaths computes uniqueness and duplicates from OpenAPI paths map.
func AnalyzeOpenAPIPaths(paths map[string]map[string]interface{}) CombinedAPIReport {
	occ := map[string]int{}
	total := 0

	for p, methods := range paths {
		for m := range methods {
			switch m {
			case "get", "post", "put", "patch", "delete", "options", "head":
				total++
				key := m + " " + p
				occ[key]++
			}
		}
	}

	dups := make([]DuplicateEntry, 0, 32)
	dupEntries := 0
	for k, c := range occ {
		if c > 1 {
			dups = append(dups, DuplicateEntry{Key: k, Occurrences: c})
			dupEntries += c - 1
		}
	}

	sort.Slice(dups, func(i, j int) bool {
		if dups[i].Occurrences == dups[j].Occurrences {
			return dups[i].Key < dups[j].Key
		}
		return dups[i].Occurrences > dups[j].Occurrences
	})

	if len(dups) > 30 {
		dups = dups[:30]
	}

	return CombinedAPIReport{
		DeclaredTotal:     total,
		UniqueMethodPath:  len(occ),
		DuplicateEntries:  dupEntries,
		DuplicateKeys:     len(dups),
		TopDuplicatePaths: dups,
	}
}
