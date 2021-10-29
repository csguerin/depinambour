package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/xlab/treeprint"

	"github.com/csguerin/depinambour/module"
)

func main() {
	graph, err := module.NewGraph()
	if err != nil {
		panic(err)
	}

	mInfo, err := module.NewModuleInfo(os.Args[1])
	if err != nil {
		panic(err)
	}

	displayDepTree(graph, graph.ModuleWithDepLowerThan(&mInfo), &mInfo)
}

func displayDepTree(graph *module.Graph, badDeps map[string]*module.Module, refMod *module.ModuleInfo) {
	tree := treeprint.NewWithRoot(nodeStr(&graph.Module, badDeps, refMod))

	stack := make(map[string]struct{})

	for _, dep := range graph.Dependencies {
		addNode(tree, &graph.Module, dep, badDeps, refMod, stack)
	}

	fmt.Println(tree.String())
}

func addToTree(tree treeprint.Tree, m *module.Module, badDeps map[string]*module.Module, refMod *module.ModuleInfo, stack map[string]struct{}) {
	tree = tree.AddBranch(nodeStr(m, badDeps, refMod))

	stack = dupStack(stack)
	for _, dep := range m.Dependencies {
		addNode(tree, m, dep, badDeps, refMod, stack)
	}
}

func addNode(tree treeprint.Tree, m *module.Module, dep *module.Module, badDeps map[string]*module.Module, refMod *module.ModuleInfo, stack map[string]struct{}) {
	if _, in := badDeps[dep.CanonicalName()]; !in {
		return
	}

	if dep.DirectDepOnLower(refMod) {
		tree.AddBranch(nodeStr(dep, badDeps, refMod))
		return
	}

	edge := m.CanonicalName() + "->" + dep.CanonicalName()
	if _, ok := stack[edge]; ok {
		tree.AddBranch(nodeStr(dep, badDeps, refMod)).AddBranch("[...] (cycle)")
		return
	}

	stack[edge] = struct{}{}
	addToTree(tree, dep, badDeps, refMod, stack)
}

func dupStack(s map[string]struct{}) map[string]struct{} {
	stack := make(map[string]struct{}, len(s)+10)
	for k := range s {
		stack[k] = struct{}{}
	}

	return stack
}

func nodeStr(m *module.Module, badDeps map[string]*module.Module, refMod *module.ModuleInfo) string {
	cp := color.New(color.FgGreen).Sprint

	if _, in := badDeps[m.CanonicalName()]; in {
		if m.DirectDepOnLower(refMod) {
			cp = color.New(color.FgRed).Sprint
		} else {
			cp = color.New(color.FgHiYellow).Sprint
		}
	}

	return cp(m.CanonicalName())
}
