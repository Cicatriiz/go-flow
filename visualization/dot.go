package visualization

import (
	"fmt"
	"github.com/forrest/go-flow/core"
	"strings"
)

// ToDOT generates a Graphviz DOT representation of the pipeline.
func ToDOT(p *core.Pipeline) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("digraph \"%s\" {\n", p.Name()))
	b.WriteString("  rankdir=LR;\n")
	b.WriteString("  node [shape=record];\n")

	components := p.GetComponents()
	for name, component := range components {
		label := fmt.Sprintf("{%s|{%s|%s}}", name, getPorts(component.InputPorts()), getPorts(component.OutputPorts()))
		b.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\"];\n", name, label))
	}

	connections := p.GetConnections()
	for _, conn := range connections {
		b.WriteString(fmt.Sprintf("  \"%s\":%s -> \"%s\":%s;\n", conn.FromComponent, conn.FromPort, conn.ToComponent, conn.ToPort))
	}

	b.WriteString("}\n")
	return b.String()
}

func getPorts(ports []core.Port) string {
	var portStrings []string
	for _, p := range ports {
		portStrings = append(portStrings, fmt.Sprintf("<%s> %s (%s)", p.Name(), p.Name(), p.Type()))
	}
	return strings.Join(portStrings, "|")
}
