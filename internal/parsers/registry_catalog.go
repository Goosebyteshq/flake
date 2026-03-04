package parsers

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type registerOptions struct {
	Aliases  []string
	Priority int
}

type registration struct {
	Parser   Parser
	Priority int
	Aliases  []string
}

var (
	registryMu sync.RWMutex
	catalog    []registration
)

func registerBuiltin(p Parser, opts registerOptions) {
	if p == nil {
		panic("parsers: nil parser registration")
	}
	if p.Name() == "" || p.Name() == FrameworkAuto {
		panic("parsers: invalid parser name")
	}

	aliases := make([]string, 0, len(opts.Aliases)+1)
	aliases = append(aliases, strings.ToLower(string(p.Name())))
	for _, a := range opts.Aliases {
		if t := strings.ToLower(strings.TrimSpace(a)); t != "" {
			aliases = append(aliases, t)
		}
	}

	registryMu.Lock()
	defer registryMu.Unlock()
	for _, existing := range catalog {
		if existing.Parser.Name() == p.Name() {
			panic(fmt.Sprintf("parsers: duplicate parser name %q", p.Name()))
		}
	}
	catalog = append(catalog, registration{Parser: p, Priority: opts.Priority, Aliases: aliases})
}

func snapshotCatalog() []registration {
	registryMu.RLock()
	defer registryMu.RUnlock()
	cp := make([]registration, len(catalog))
	copy(cp, catalog)
	sort.SliceStable(cp, func(i, j int) bool {
		if cp[i].Priority != cp[j].Priority {
			return cp[i].Priority < cp[j].Priority
		}
		return cp[i].Parser.Name() < cp[j].Parser.Name()
	})
	return cp
}

func registeredAlias(alias string) (Framework, bool) {
	alias = strings.ToLower(strings.TrimSpace(alias))
	if alias == "" {
		return "", false
	}
	for _, reg := range snapshotCatalog() {
		for _, a := range reg.Aliases {
			if alias == a {
				return reg.Parser.Name(), true
			}
		}
	}
	return "", false
}
