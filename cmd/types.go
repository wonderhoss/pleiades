package main

// Stoppable is a component that can be instructed to shut down
type Stoppable interface {
	Stop()
}
