package main

import (
	"flag"
	"log"
	"os"

	"github.com/konradreiche/pathtrace/internal/analyzer"
	"github.com/konradreiche/pathtrace/internal/printer"
	"github.com/konradreiche/pathtrace/internal/server"
	"github.com/konradreiche/pathtrace/ui"
)

func main() {
	path := flag.String("trace", "trace.out", "path to Go execution trace")
	task := flag.String("task", "", "task type to perform analysis on")
	printForLabel := flag.String("show", "", "print nodes for label (p50, p90, max)")
	printCriticalPath := flag.Bool("critical", false, "print nodes for the critical path")
	regionPrefix := flag.String("prefix", "", "common region prefix to remove")
	addr := flag.String("ui", "", "address for the UI server")

	flag.Parse()

	f, err := os.Open(*path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	analyzer := analyzer.New(*task, *regionPrefix)
	if err := analyzer.ProcessTrace(f); err != nil {
		log.Fatal(err)
	}

	if *addr == "" {
		printer := printer.New(analyzer.Tasks, analyzer.NodesByTask)
		printer.PrintTrees(*printForLabel, *printCriticalPath)
		return
	}

	srv := server.New(ui.Assets, analyzer)
	if err := srv.Start(*addr); err != nil {
		log.Fatal(err)
	}
}
