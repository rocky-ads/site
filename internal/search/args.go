package search

import "fmt"

type pgArgs struct {
	args []any
}

func (p *pgArgs) add(v any) string {
	p.args = append(p.args, v)
	return fmt.Sprintf("$%d", len(p.args))
}
