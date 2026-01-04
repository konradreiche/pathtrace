package printer

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/konradreiche/pathtrace/internal/analyzer"
	"golang.org/x/exp/trace"
)

type Printer struct {
	tasks       map[trace.TaskID]*analyzer.Task
	nodesByTask map[trace.TaskID][]*analyzer.Node
}

func New(
	tasks map[trace.TaskID]*analyzer.Task,
	nodesByTask map[trace.TaskID][]*analyzer.Node,
) *Printer {
	return &Printer{
		tasks:       tasks,
		nodesByTask: nodesByTask,
	}
}

func (p *Printer) PrintTrees(
	showLabel string,
	printCriticalPath bool,
) {
	tasks := slices.SortedFunc(
		maps.Values(p.tasks),
		func(a, b *analyzer.Task) int {
			return cmp.Compare(a.Duration, b.Duration)
		},
	)

	n := len(tasks)
	p100 := tasks[n-1]
	p90 := tasks[int(0.9*float64(n-1))]
	p50 := tasks[n/2]

	p.printTask(p100, "max", showLabel, printCriticalPath)
	p.printTask(p90, "p90", showLabel, printCriticalPath)
	p.printTask(p50, "p50", showLabel, printCriticalPath)
}

func (p *Printer) printTask(
	task *analyzer.Task,
	label string,
	showLabel string,
	printCriticalPath bool,
) {
	// TODO: move to filter this out after processing
	if task.End == 0 {
		return
	}

	if showLabel == "" {
		fmt.Printf("%s: %s uid=%s (%v)\n", label, task.Type, task.UID, task.Duration)
		return
	}
	if showLabel != label {
		return
	}

	nodes := p.nodesByTask[task.ID]
	if printCriticalPath {
		nodes = analyzer.CriticalPath(nodes[0])
		PrintCriticalPath(nodes)
		return
	}

	p.PrintNodes(nodes)

	fmt.Println()
}

func (p *Printer) PrintNodes(nodes []*analyzer.Node) {
	var roots []*analyzer.Node
	for _, n := range nodes {
		if n.Parent == nil {
			roots = append(roots, n)
		}
	}

	for i, r := range roots {
		last := i == len(roots)-1
		p.printNode(r, "", last)
	}

	fmt.Println()
}

func (p *Printer) printNode(n *analyzer.Node, prefix string, last bool) {
	connector := "├── "
	nextPrefix := prefix + "│   "

	if last {
		connector = "└── "
		nextPrefix = prefix + "    "
	}

	dur := n.End.Sub(n.Start)

	fmt.Printf("%s%s%s (%v)\n",
		prefix,
		connector,
		n.Type,
		dur,
	)

	for i, child := range n.Children {
		p.printNode(
			child,
			nextPrefix,
			i == len(n.Children)-1,
		)
	}
}

func PrintCriticalPath(path []*analyzer.Node) {
	fmt.Println("Critical path:")
	for i, n := range path {
		prefix := ""
		connector := "└── "

		if i > 0 {
			prefix = strings.Repeat("    ", i)
		}

		fmt.Printf("%s%s%s (%s)\n",
			prefix,
			connector,
			n.Type,
			n.Duration(),
		)
	}
}
