package execution

import (
	"fmt"
	"github.com/forrest/go-flow/core"
)

//- `nodes`: A map where keys are component names and values are the component objects.
//- `edges`: A map representing the directed edges of the graph, where each key is a component name and the value is a list of names of components it connects to.
//- `inDegree`: A map where keys are component names and values are their in-degrees (the number of incoming edges).
type Graph struct {
	nodes    map[string]core.Component
	edges    map[string][]string
	inDegree map[string]int
}

// NewGraph creates a new graph from a pipeline definition.
func NewGraph(p *core.Pipeline) *Graph {
	g := &Graph{
		nodes:    make(map[string]core.Component),
		edges:    make(map[string][]string),
		inDegree: make(map[string]int),
	}

	components := p.GetComponents()
	for name, comp := range components {
		g.nodes[name] = comp
		g.inDegree[name] = 0
	}

	for _, conn := range p.GetConnections() {
		g.edges[conn.FromComponent] = append(g.edges[conn.FromComponent], conn.ToComponent)
		g.inDegree[conn.ToComponent]++
	}

	return g
}

// TopologicalSort performs a topological sort of the graph and returns a list of
// component names in execution order.
func (g *Graph) TopologicalSort() ([]string, error) {
	var sorted []string
	queue := []string{}

	for name, degree := range g.inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		sorted = append(sorted, node)

		for _, neighbor := range g.edges[node] {
			g.inDegree[neighbor]--
			if g.inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(sorted) != len(g.nodes) {
		return nil, fmt.Errorf("graph has a cycle")
	}

	return sorted, nil
}
