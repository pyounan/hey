package socket

var registry map[string]chan Event

func init() {
	registry = make(map[string]chan Event)
}

// Register hooks a module channel to the socket registery, so when a new
func Register(module string, ch chan Event) {
	registry[module] = ch
}
