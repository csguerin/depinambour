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
	tree := treeprint.New()

	stack := make(map[string]struct{})
	addToTree(tree, &graph.Module, badDeps, refMod, stack, true)

	fmt.Println(tree.String())
}

func addToTree(tree treeprint.Tree, m *module.Module, badDeps map[string]*module.Module, refMod *module.ModuleInfo, stack map[string]struct{}, topLvl bool) {
	tree = tree.AddBranch(nodeStr(m, badDeps, refMod))

	for _, dep := range m.Dependencies {
		if _, in := badDeps[dep.CanonicalName()]; !in && !topLvl {
			continue
		}

		edge := m.CanonicalName() + "->" + dep.CanonicalName()
		if _, ok := stack[edge]; ok {
			tree = tree.AddBranch(nodeStr(dep, badDeps, refMod) + " [...] (cycle)")
			continue
		}
		stack[edge] = struct{}{}

		// cStack := dupStack(stack)

		addToTree(tree, dep, badDeps, refMod, stack, false)
	}
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
