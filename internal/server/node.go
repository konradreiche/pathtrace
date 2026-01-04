package server

import (
	"time"

	"github.com/konradreiche/pathtrace/internal/analyzer"
)

type Node struct {
	ID       int           `json:"id"`
	ParentID int           `json:"parent_id,omitempty"`
	Label    string        `json:"type"`
	Duration time.Duration `json:"duration"`
}

func flatten(root *analyzer.Node) []Node {
	var nodes []Node
	var id int

	var visit func(n *analyzer.Node, parent int)
	visit = func(n *analyzer.Node, parent int) {
		id++
		currentID := id

		nodes = append(nodes, Node{
			ID:       currentID,
			ParentID: parent,
			Label:    n.Type,
			Duration: n.Duration(),
		})

		for _, c := range n.Children {
			visit(c, currentID)
		}
	}

	visit(root, 0)
	return nodes
}
