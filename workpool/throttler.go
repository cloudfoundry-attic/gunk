package workpool

type Throttler struct {
	pool  *WorkPool
	works []func()
}

func NewThrottler(maxWorkers int, works []func()) (*Throttler, error) {
	var numWorkers int
	if len(works) < maxWorkers {
		numWorkers = len(works)
	} else {
		numWorkers = maxWorkers
	}

	pool, err := newWorkPoolWithPending(numWorkers, len(works)-numWorkers)
	if err != nil {
		return nil, err
	}

	return &Throttler{
		pool:  pool,
		works: works,
	}, nil
}

func (t *Throttler) Stop() {
	t.pool.Stop()
}

func (t *Throttler) Start() {
	for _, work := range t.works {
		t.pool.Submit(work)
	}
}
