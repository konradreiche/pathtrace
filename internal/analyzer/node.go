package analyzer

import (
	"strings"
	"time"

	"golang.org/x/exp/trace"
)

type Node struct {
	Type   string
	TaskID trace.TaskID
	Start  trace.Time
	End    trace.Time

	Parent   *Node
	Children []*Node
}

func newNode(
	ev trace.Event,
	region trace.Region,
	parent *Node,
	regionPrefix string,
) *Node {
	n := &Node{
		Type:   strings.TrimPrefix(region.Type, regionPrefix),
		TaskID: region.Task,
		Start:  ev.Time(),
		Parent: parent,
	}
	if parent != nil {
		parent.Children = append(parent.Children, n)
	}
	return n
}

func (n *Node) Duration() time.Duration {
	return n.End.Sub(n.Start)
}
