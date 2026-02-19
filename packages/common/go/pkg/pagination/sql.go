package pagination

import "fmt"

// ApplyToQuery appends LIMIT and OFFSET clauses to a SQL query.
// It returns the modified query string and the two args (limit, offset)
// that should be appended to the existing argument slice.
//
// argIndex is the next positional placeholder index (e.g., if $3 is the
// next available, pass 3). The returned query uses $argIndex for LIMIT
// and $(argIndex+1) for OFFSET.
//
// Example:
//
//	q, limitArgs := pagination.ApplyToQuery(baseQuery, params, 3)
//	args = append(args, limitArgs...)
func ApplyToQuery(query string, p Params, argIndex int) (string, []interface{}) {
	p.Normalize()
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	return query, []interface{}{p.Limit, p.Offset()}
}
