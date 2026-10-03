package jvmmon

import (
	"sync"

	ui "github.com/gizak/termui/v3"
)

// uiMu serializes all widget mutations and rendering. termui is not
// goroutine-safe and EventBus handlers run concurrently.
var uiMu sync.Mutex

// WithUI runs fn while holding the UI lock.
func WithUI(fn func()) {
	uiMu.Lock()
	defer uiMu.Unlock()
	fn()
}

// Render renders drawables while holding the UI lock.
func Render(items ...ui.Drawable) {
	WithUI(func() { ui.Render(items...) })
}
