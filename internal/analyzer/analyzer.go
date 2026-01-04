package analyzer

import (
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/exp/trace"
)

// Analyzer processes a Go execution trace and reconstructs per-task execution
// trees from user regions.
//
// Regions are grouped by task and nested according to their begin and end
// order, producing a tree of steps with measured durations. The resulting
// trees can be used to inspect the tree structure, and compute critical paths.
type Analyzer struct {
	Tasks       map[trace.TaskID]*Task
	NodesByTask map[trace.TaskID][]*Node

	stacks       map[trace.TaskID][]*Node
	taskType     string
	regionPrefix string
}

func New(taskType, regionPrefix string) *Analyzer {
	return &Analyzer{
		Tasks:        make(map[trace.TaskID]*Task),
		stacks:       make(map[trace.TaskID][]*Node),
		NodesByTask:  make(map[trace.TaskID][]*Node),
		taskType:     taskType,
		regionPrefix: regionPrefix,
	}
}

func (a *Analyzer) ProcessTrace(r io.Reader) error {
	reader, err := trace.NewReader(r)
	if err != nil {
		return err
	}

	for {
		ev, err := reader.ReadEvent()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if err := a.processEvent(ev); err != nil {
			return err
		}
	}

	return nil
}

func CriticalPath(root *Node) []*Node {
	var path []*Node
	cur := root

	for cur != nil {
		path = append(path, cur)

		if len(cur.Children) == 0 {
			break
		}

		var next *Node
		var slowest time.Duration

		for _, n := range cur.Children {
			if d := n.Duration(); d > slowest {
				slowest = d
				next = n
			}
		}

		cur = next
	}

	return path
}

func (a *Analyzer) processEvent(ev trace.Event) error {
	switch kind := ev.Kind(); kind {
	case trace.EventTaskBegin:
		task, err := newTask(ev)
		if err != nil {
			return err
		}
		if task.Type != a.taskType {
			return nil
		}
		a.Tasks[task.ID] = task
	case trace.EventTaskEnd:
		task, ok := a.Tasks[ev.Task().ID]
		if !ok {
			return nil
		}
		task.End = ev.Time()
		task.Duration = task.End.Sub(task.Start)
	case trace.EventLog:
		log := ev.Log()
		task, ok := a.Tasks[log.Task]
		if !ok {
			return nil
		}
		if log.Category == "request_id" {
			task.UID = log.Message
		}
	case trace.EventRegionBegin:
		region := ev.Region()
		if region.Task == 0 || region.Type == "" {
			return nil
		}
		if _, ok := a.Tasks[region.Task]; !ok {
			return nil
		}
		taskID := region.Task
		stack := a.stacks[taskID]

		var parent *Node
		if len(stack) > 0 {
			parent = stack[len(stack)-1]
		}

		node := newNode(ev, region, parent, a.regionPrefix)
		a.NodesByTask[region.Task] = append(a.NodesByTask[region.Task], node)
		a.stacks[taskID] = append(stack, node)
	case trace.EventRegionEnd:
		taskID := ev.Region().Task
		stack := a.stacks[taskID]
		if len(stack) == 0 {
			return nil
		}

		node := stack[len(stack)-1]
		a.stacks[taskID] = stack[:len(stack)-1]

		node.End = ev.Time()
	}
	return nil
}

type Task struct {
	ID   trace.TaskID
	Type string
	UID  string

	Start    trace.Time
	End      trace.Time
	Duration time.Duration
}

func newTask(ev trace.Event) (*Task, error) {
	if ev.Kind() != trace.EventTaskBegin {
		return nil, fmt.Errorf("invalid task kind %q; new eask expects: %s", ev.Kind(), trace.EventTaskBegin)
	}
	task := ev.Task()
	return &Task{
		ID:    task.ID,
		Type:  task.Type,
		Start: ev.Time(),
	}, nil
}
