package logsift

import "sync"

type Filter interface {
	Add(filters ...string)
	Remove(filters ...string)
	Set(filters ...string)
	SetMap(filters map[string]bool)
	SetAllowEmptyFilter(allowEmpty bool)
	Allows(values ...string) bool
}

type concurrentMapFilter struct {
	sync.RWMutex
	allowEmptyFilter bool
	filters          map[string]bool
}

func NewConcurrentMapFilter(allowEmptyFilter bool) Filter {
	return &concurrentMapFilter{
		allowEmptyFilter: allowEmptyFilter,
		filters:          make(map[string]bool),
	}
}

func (f *concurrentMapFilter) Add(filters ...string) {
	f.Lock()
	defer f.Unlock()
	for _, filter := range filters {
		if filter != "" {
			f.filters[filter] = true
		}
	}
}

func (f *concurrentMapFilter) Remove(filters ...string) {
	f.Lock()
	defer f.Unlock()
	for _, filter := range filters {
		if filter != "" {
			delete(f.filters, filter)
		}
	}
}

func (f *concurrentMapFilter) Set(filters ...string) {
	f.Lock()
	defer f.Unlock()
	f.filters = make(map[string]bool)
	for _, filter := range filters {
		if filter != "" {
			f.filters[filter] = true
		}
	}
}

func (f *concurrentMapFilter) SetMap(filters map[string]bool) {
	f.Lock()
	defer f.Unlock()
	f.filters = filters
	for filter := range f.filters {
		if filter == "" {
			delete(f.filters, filter)
		}
	}
}

func (f *concurrentMapFilter) SetAllowEmptyFilter(allowEmpty bool) {
	f.Lock()
	defer f.Unlock()
	f.allowEmptyFilter = allowEmpty
}

func (f *concurrentMapFilter) Allows(values ...string) bool {
	f.RLock()
	defer f.RUnlock()
	if f.filters == nil || len(f.filters) == 0 {
		return f.allowEmptyFilter
	}
	for _, value := range values {
		if _, ok := f.filters[value]; ok {
			return true
		}
	}
	return false
}

type unsafeMapFilter struct {
	allowEmptyFilter bool
	filters          map[string]bool
}

func NewUnsafeMapFilter(allowEmptyFilter bool) Filter {
	return &unsafeMapFilter{
		allowEmptyFilter: allowEmptyFilter,
		filters:          make(map[string]bool),
	}
}

func (f *unsafeMapFilter) Add(filters ...string) {
	for _, filter := range filters {
		if filter != "" {
			f.filters[filter] = true
		}
	}
}

func (f *unsafeMapFilter) Remove(filters ...string) {
	for _, filter := range filters {
		if filter != "" {
			delete(f.filters, filter)
		}
	}
}

func (f *unsafeMapFilter) Set(filters ...string) {
	f.filters = make(map[string]bool)
	for _, filter := range filters {
		if filter != "" {
			f.filters[filter] = true
		}
	}
}

func (f *unsafeMapFilter) SetMap(filters map[string]bool) {
	f.filters = filters
	for filter := range f.filters {
		if filter == "" {
			delete(f.filters, filter)
		}
	}
}

func (f *unsafeMapFilter) SetAllowEmptyFilter(allowEmpty bool) {
	f.allowEmptyFilter = allowEmpty
}

func (f *unsafeMapFilter) Allows(values ...string) bool {
	if f.filters == nil || len(f.filters) == 0 {
		return f.allowEmptyFilter
	}
	for _, value := range values {
		if _, ok := f.filters[value]; ok {
			return true
		}
	}
	return false
}
