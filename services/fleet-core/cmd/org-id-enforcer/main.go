package main

import (
	"github.com/flyfleet/fleet-core/internal/tools/analysis/org_id_enforcer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(org_id_enforcer.Analyzer)
}
