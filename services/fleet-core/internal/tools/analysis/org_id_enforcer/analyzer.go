package org_id_enforcer

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "orgidenforcer",
	Doc:      "checks that SQL queries contain org_id for tenant isolation",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		// Check for common DB methods
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		methodName := sel.Sel.Name
		isSQLMethod := false
		queryArgIdx := -1

		switch methodName {
		case "Exec", "Query", "QueryRow":
			isSQLMethod = true
			queryArgIdx = 0
		case "ExecContext", "QueryContext", "QueryRowContext":
			isSQLMethod = true
			queryArgIdx = 1
		}

		if !isSQLMethod || len(call.Args) <= queryArgIdx {
			return
		}

		// Inspect the query argument
		queryArg := call.Args[queryArgIdx]

		// Support basic string literals
		if lit, ok := queryArg.(*ast.BasicLit); ok {
			query := strings.ToLower(lit.Value)
			
			// Exclude system tables and schema migrations
			if strings.Contains(query, "schema_migrations") || 
			   strings.Contains(query, "information_schema") ||
			   strings.Contains(query, "pg_catalog") {
				return
			}

			// Only check SELECT, UPDATE, DELETE
			isSensitive := strings.Contains(query, "select") || 
			               strings.Contains(query, "update") || 
						   strings.Contains(query, "delete")

			if isSensitive && !strings.Contains(query, "org_id") {
				pass.Reportf(call.Lparen, "SQL query missing org_id filter: %s", lit.Value)
			}
		}

		// Support simple string concatenations or variables could be added later
	})

	return nil, nil
}
