package app

import "sync"

var (
	finalizers      []func()
	finalizersMutex sync.Mutex
)

func OnFinalize(f func()) {
	finalizersMutex.Lock()
	defer finalizersMutex.Unlock()
	finalizers = append(finalizers, f)
}

func InvokeFinalizers() {
	finalizersMutex.Lock()
	defer finalizersMutex.Unlock()
	for _, f := range finalizers {
		f()
	}
}
