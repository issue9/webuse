// SPDX-License-Identifier: MIT

package health

import "sort"

type memory struct {
	states []*State
}

// NewMemory 声明 Store 的内存实现方式
func NewMemory(capacity int64) Store {
	return &memory{
		states: make([]*State, 0, capacity),
	}
}

func (mem *memory) Get(method, path string) *State {
	for _, state := range mem.states {
		if state.Method == method && state.Path == path {
			return state
		}
	}

	state := &State{
		Method: method,
		Path:   path,
	}
	mem.states = append(mem.states, state)
	return state
}

func (mem *memory) Save(state *State) {
	s := mem.Get(state.Method, state.Path)
	*s = *state
}

func (mem *memory) All() []*State {
	sort.SliceStable(mem.states, func(i, j int) bool {
		ii := mem.states[i]
		jj := mem.states[j]
		if ii.Path != jj.Path {
			return ii.Path > jj.Path
		}
		return ii.Method > jj.Method
	})
	return mem.states
}
