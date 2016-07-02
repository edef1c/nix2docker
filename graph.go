package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

type Graph map[string]GraphNode

type GraphNode struct {
	Deriver    string
	References []string
}

func (g Graph) Paths() []string {
	paths := make([]string, 0, len(g))
	for path := range g {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func (g Graph) Requisites(out Graph, path string) Graph {
	if out == nil {
		out = make(Graph)
	}
	for _, ref := range g[path].References {
		g.Closure(out, ref)
	}
	return out
}

func (g Graph) Closure(out Graph, path string) Graph {
	if out == nil {
		out = make(Graph)
	}
	if _, ok := out[path]; ok {
		return out
	}
	node := g[path]
	out[path] = node
	for _, path := range node.References {
		g.Closure(out, path)
	}
	return out
}

func ParseGraphFile(g Graph, path string) (Graph, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseGraph(g, f)
}

func ParseGraph(g Graph, r io.Reader) (Graph, error) {
	if g == nil {
		g = make(Graph)
	}
	s := bufio.NewScanner(r)
	ok := true
loop:
	for s.Scan() {
		storePath := s.Text()
		if ok = s.Scan(); !ok {
			break loop
		}
		deriver := s.Text()
		if ok = s.Scan(); !ok {
			break loop
		}

		count, err := strconv.Atoi(s.Text())
		if err != nil {
			return nil, fmt.Errorf("couldn't parse reference count: %s", err)
		}

		references := make([]string, count)
		for i := range references {
			if ok = s.Scan(); !ok {
				break loop
			}
			references[i] = s.Text()
		}

		g[storePath] = GraphNode{
			Deriver:    deriver,
			References: references,
		}
	}
	err := s.Err()
	if !ok && err == nil {
		err = io.ErrUnexpectedEOF
	}
	if err != nil {
		return nil, err
	}
	return g, nil
}
