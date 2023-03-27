package consul

// "time"

type Watcher struct {
	client *Client
	// update    chan *naming.Update
	side      chan int
	close     chan bool
	address   []string
	lastIndex uint64
}

type Resolver struct {
	w      *Watcher
	nodes  []string
	scheme string
	tag    string
	proto  string
}

func NewConsulResolver(nodes []string, tag string) *Resolver {
	r := Resolver{
		nodes:  nodes,
		scheme: "http",
		tag:    tag,
		proto:  "http",
	}
	return &r
}

func newConsulWatcher(nodes []string, scheme string) (*Watcher, error) {
	c, err := New(nodes, scheme)
	if err != nil {
		return nil, err
	}
	w := Watcher{
		client: c,
		// update: make(chan *naming.Update, 1),
		side:  make(chan int, 1),
		close: make(chan bool),
	}
	return &w, nil
}

func (w *Watcher) watch(target, tag, proto string) ([]string, error) {
	return nil, nil
}

func (w *Watcher) Close() {
	close(w.side)
	close(w.close)
}
