package module

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"
)

type Graph struct {
	Modules map[string]*Module `json:"-"`
	Module
}

func NewGraph() (*Graph, error) {
	pr, pw := io.Pipe()
	defer pw.Close()
	defer pr.Close()

	var cmdError error
	go func() {
		defer pw.Close()

		cmd := exec.Command("go", "mod", "graph")
		cmd.Stdout = pw
		cmdError = cmd.Run()
	}()

	scanner := bufio.NewScanner(pr)

	var graph Graph

	for scanner.Scan() {
		line := scanner.Text()

		entry := strings.Split(line, " ")
		if len(entry) != 2 {
			continue
		}

		if err := graph.AddDependency(entry[0], entry[1]); err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if cmdError != nil {
		return nil, cmdError
	}

	return &graph, nil
}

func (g *Graph) AddDependency(parent, child string) error {
	pInfo, err := NewModuleInfo(parent)
	if err != nil {
		return err
	}

	cInfo, err := NewModuleInfo(child)
	if err != nil {
		return err
	}

	if g.Modules == nil {
		g.Modules = map[string]*Module{}
	}

	cMod, in := g.Modules[cInfo.CanonicalName()]
	if !in {
		cMod = &Module{
			ModuleInfo: cInfo,
		}

		g.Modules[cInfo.CanonicalName()] = cMod
	}

	mod, ok := g.Modules[pInfo.CanonicalName()]
	if ok {
		mod.addDependency(cMod)
		return nil
	}

	g.Module = Module{
		ModuleInfo: pInfo,
	}
	g.Module.addDependency(cMod)
	g.Modules[g.Module.CanonicalName()] = &g.Module

	return nil
}

func (g *Graph) ModuleWithDepLowerThan(dep *ModuleInfo) map[string]*Module {
	mods := make(map[string]*Module, len(g.Modules))

	for _, m := range g.Modules {
		if m.DirectDepOnLower(dep) {
			mods[m.CanonicalName()] = m
		}
	}

	for {
		modsLen := len(mods)

		for _, dep := range mods {
			for _, m := range g.Modules {
				if _, in := m.Dependencies[dep.CanonicalName()]; in {
					mods[m.CanonicalName()] = m
				}
			}
		}

		if modsLen == len(mods) {
			break
		}
	}

	return mods
}

type Module struct {
	ModuleInfo
	Dependencies map[string]*Module
}

func (m *Module) DirectDepOnLower(mod *ModuleInfo) bool {
	for _, dep := range m.Dependencies {
		if dep.ModuleInfo.Path == mod.Path {
			if dep.Version.LessThan(mod.Version) {
				return true
			}
		}
	}

	return false
}

func (m *Module) addDependency(mod *Module) {
	if m.Dependencies == nil {
		m.Dependencies = map[string]*Module{}
	}

	m.Dependencies[mod.CanonicalName()] = mod
}

type ModuleInfo struct {
	Version *semver.Version
	Path    string
}

func (m *ModuleInfo) CanonicalName() string {
	if m.Version == nil {
		return m.Path
	}

	return m.Path + "@" + m.Version.Original()
}

func (m ModuleInfo) String() string {
	return m.CanonicalName()
}

func NewModuleInfo(mStr string) (ModuleInfo, error) {
	split := strings.Split(mStr, "@")

	mInfo := ModuleInfo{
		Path: split[0],
	}

	if len(split) == 1 {
		return mInfo, nil
	}

	version, err := semver.NewVersion(split[1])
	if err != nil {
		return mInfo, fmt.Errorf("cannot parse mod version: %w", err)
	}

	mInfo.Version = version
	return mInfo, nil
}
