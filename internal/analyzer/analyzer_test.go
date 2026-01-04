package analyzer

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"runtime/trace"

	exptrace "golang.org/x/exp/trace"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestProcessTrace(t *testing.T) {
	tests := []struct {
		name string
		root *root
		want *Analyzer
	}{
		{
			name: "simple",
			root: newRoot("simple",
				newFork(
					newLeaf("a"),
					newLeaf("b"),
				),
				newLeaf("c"),
			),
			want: &Analyzer{
				Tasks: map[exptrace.TaskID]*Task{
					1: {
						ID:   1,
						Type: "simple",
						UID:  "0",
					},
				},
				stacks: map[exptrace.TaskID][]*Node{1: {}},
				NodesByTask: map[exptrace.TaskID][]*Node{
					1: {
						node("seq",
							node("fork",
								node("b"),
								node("a"),
							),
							node("c"),
						),
						node("fork",
							node("b"),
							node("a"),
						),
						node("b"),
						node("a"),
						node("c"),
					},
				},
				taskType:     "simple",
				regionPrefix: "processer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := trace.Start(&buf); err != nil {
				t.Fatal(err)
			}
			tt.root.process(t.Context())
			trace.Stop()

			analyzer := New("simple", "processer")
			if err := analyzer.ProcessTrace(&buf); err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(analyzer, tt.want,
				cmp.AllowUnexported(Analyzer{}),
				cmpopts.IgnoreFields(
					Task{},
					"Start",
					"End",
					"Duration",
				),
				cmpopts.IgnoreFields(
					Node{},
					"Start",
					"End",
					"Parent",
					"TaskID",
				),
			); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestCriticalPath(t *testing.T) {
	root := newRoot("simple",
		newFork(
			newLeaf("a"),
			newLeaf("b"),
		),
		newLeaf("c", WithSleep(1*time.Millisecond)),
	)

	var buf bytes.Buffer
	if err := trace.Start(&buf); err != nil {
		t.Fatal(err)
	}
	root.process(t.Context())
	trace.Stop()

	analyzer := New("simple", "processer")
	if err := analyzer.ProcessTrace(&buf); err != nil {
		t.Error(err)
	}

	nodes := analyzer.NodesByTask[2]
	criticalPath := CriticalPath(nodes[0])

	want := []*Node{
		node("seq"),
		node("c"),
	}
	if diff := cmp.Diff(criticalPath, want,
		cmpopts.IgnoreFields(
			Node{},
			"Start",
			"End",
			"Parent",
			"Children",
			"TaskID",
		),
	); diff != "" {
		t.Errorf("diff: %s", diff)
	}
}

func PrintCriticalPath(path []*Node) {
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

type processer interface {
	process(ctx context.Context)
	name() string
}

type seq struct {
	children []processer
}

func (p *seq) process(ctx context.Context) {
	region := trace.StartRegion(ctx, ""+p.name())
	defer region.End()
	for _, node := range p.children {
		node.process(ctx)
	}
}

func newSeq(children ...processer) *seq {
	return &seq{
		children: children,
	}
}

func (p *seq) name() string {
	return "seq"
}

type fork struct {
	children []processer
}

func newFork(children ...processer) *fork {
	return &fork{
		children: children,
	}
}

func (p *fork) process(ctx context.Context) {
	region := trace.StartRegion(ctx, ""+p.name())
	defer region.End()
	var wg sync.WaitGroup
	for _, node := range p.children {
		wg.Go(func() {
			node.process(ctx)
		})
	}
	wg.Wait()
}

func (p *fork) name() string {
	return "fork"
}

type leaf struct {
	label string
	sleep time.Duration
}

func newLeaf(label string, opts ...LeafOption) *leaf {
	cfg := leafOptions{}
	WithLeafOptions(opts...)(&cfg)
	return &leaf{
		label: label,
		sleep: cfg.sleep,
	}
}

func (p *leaf) process(ctx context.Context) {
	region := trace.StartRegion(ctx, ""+p.name())
	defer region.End()
	time.Sleep(p.sleep)
}

func (p *leaf) name() string {
	return p.label
}

type leafOptions struct {
	sleep time.Duration
}

type LeafOption func(*leafOptions)

func WithSleep(sleep time.Duration) LeafOption {
	return func(o *leafOptions) {
		o.sleep = sleep
	}
}

func WithLeafOptions(opts ...LeafOption) LeafOption {
	return func(o *leafOptions) {
		for _, opt := range opts {
			opt(o)
		}
	}
}

type root struct {
	label string
	child processer
}

func newRoot(label string, children ...processer) *root {
	return &root{
		label: label,
		child: newSeq(children...),
	}
}

func (p *root) process(ctx context.Context) {
	ctx, task := trace.NewTask(ctx, p.label)
	defer task.End()
	trace.Log(ctx, "request_id", "0")

	p.child.process(ctx)
}

func (p *root) name() string {
	return "root"
}

func node(label string, children ...*Node) *Node {
	return &Node{
		Type:     label,
		Children: children,
	}
}
