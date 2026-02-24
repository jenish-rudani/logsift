package logsift

import (
	"fmt"
	"sync"
	"testing"
)

// filterFactories returns constructors for both Filter implementations
// so every test runs against both with zero duplication.
func filterFactories() map[string]func(bool) Filter {
	return map[string]func(bool) Filter{
		"ConcurrentMapFilter": func(allow bool) Filter { return NewConcurrentMapFilter(allow) },
		"UnsafeMapFilter":     func(allow bool) Filter { return NewUnsafeMapFilter(allow) },
	}
}

func TestFilter_Add_and_Allows(t *testing.T) {
	for name, factory := range filterFactories() {
		t.Run(name, func(t *testing.T) {
			f := factory(false)
			f.Add("auth")

			if !f.Allows("auth") {
				t.Error("expected Allows('auth') to be true after Add")
			}
			if f.Allows("db") {
				t.Error("expected Allows('db') to be false — never added")
			}
		})
	}
}

func TestFilter_Add_IgnoresEmptyString(t *testing.T) {
	for name, factory := range filterFactories() {
		t.Run(name, func(t *testing.T) {
			f := factory(false)
			f.Add("")

			if f.Allows("") {
				t.Error("expected Allows('') to be false — empty string should be ignored")
			}
		})
	}
}

func TestFilter_Remove(t *testing.T) {
	for name, factory := range filterFactories() {
		t.Run(name, func(t *testing.T) {
			f := factory(false)
			f.Add("auth")
			f.Remove("auth")

			if f.Allows("auth") {
				t.Error("expected Allows('auth') to be false after Remove")
			}

			// Removing non-existent filter should not panic
			f.Remove("nonexistent")
		})
	}
}

func TestFilter_Set_ReplacesAll(t *testing.T) {
	for name, factory := range filterFactories() {
		t.Run(name, func(t *testing.T) {
			f := factory(false)
			f.Add("a")
			f.Add("b")
			f.Set("c")

			if f.Allows("a") {
				t.Error("expected 'a' to be gone after Set")
			}
			if f.Allows("b") {
				t.Error("expected 'b' to be gone after Set")
			}
			if !f.Allows("c") {
				t.Error("expected 'c' to be present after Set")
			}
		})
	}
}

func TestFilter_SetMap(t *testing.T) {
	for name, factory := range filterFactories() {
		t.Run(name, func(t *testing.T) {
			f := factory(false)
			f.SetMap(map[string]bool{
				"auth": true,
				"db":   true,
				"":     true, // should be stripped
			})

			if !f.Allows("auth") {
				t.Error("expected 'auth' to be present")
			}
			if !f.Allows("db") {
				t.Error("expected 'db' to be present")
			}
			// Empty key should have been removed
			if f.Allows("") {
				t.Error("expected empty string key to be stripped by SetMap")
			}
		})
	}
}

func TestFilter_Allows_EmptyFilter_AllowTrue(t *testing.T) {
	for name, factory := range filterFactories() {
		t.Run(name, func(t *testing.T) {
			f := factory(true) // allowEmptyFilter = true

			if !f.Allows("anything") {
				t.Error("expected Allows to return true when filter map is empty and allowEmptyFilter=true")
			}
		})
	}
}

func TestFilter_Allows_EmptyFilter_AllowFalse(t *testing.T) {
	for name, factory := range filterFactories() {
		t.Run(name, func(t *testing.T) {
			f := factory(false) // allowEmptyFilter = false

			if f.Allows("anything") {
				t.Error("expected Allows to return false when filter map is empty and allowEmptyFilter=false")
			}
		})
	}
}

func TestFilter_Allows_MultipleValues(t *testing.T) {
	for name, factory := range filterFactories() {
		t.Run(name, func(t *testing.T) {
			f := factory(false)
			f.Add("auth")

			// At least one value matches
			if !f.Allows("db", "auth") {
				t.Error("expected Allows('db', 'auth') to be true — 'auth' matches")
			}

			// No values match
			if f.Allows("db", "cache") {
				t.Error("expected Allows('db', 'cache') to be false — neither matches")
			}
		})
	}
}

func TestFilter_SetAllowEmptyFilter(t *testing.T) {
	for name, factory := range filterFactories() {
		t.Run(name, func(t *testing.T) {
			f := factory(false)

			if f.Allows("x") {
				t.Error("expected false with allowEmptyFilter=false")
			}

			f.SetAllowEmptyFilter(true)

			if !f.Allows("x") {
				t.Error("expected true after SetAllowEmptyFilter(true)")
			}
		})
	}
}

func TestFilter_Allows_NoArgs(t *testing.T) {
	for name, factory := range filterFactories() {
		t.Run(name, func(t *testing.T) {
			f := factory(false)
			f.Add("auth")

			if f.Allows() {
				t.Error("expected Allows() with zero args to return false")
			}
		})
	}
}

func TestConcurrentMapFilter_ThreadSafety(t *testing.T) {
	f := NewConcurrentMapFilter(false)
	var wg sync.WaitGroup
	const goroutines = 50
	const ops = 200

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("filter-%d", id)
			for j := 0; j < ops; j++ {
				f.Add(key)
				f.Allows(key)
				f.Remove(key)
				f.Set(key, fmt.Sprintf("other-%d", j))
				f.SetAllowEmptyFilter(j%2 == 0)
				f.SetMap(map[string]bool{key: true})
			}
		}(i)
	}
	wg.Wait()
}
