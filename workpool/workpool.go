package workpool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	waitTimeout = 5 * time.Second
)

type WorkPool struct {
	workQ    chan func()
	stopping chan struct{}
	stopped  int32

	mutex       sync.Mutex
	maxWorkers  int
	numWorkers  int
	idleWorkers int
}

func NewWorkPool(workers int) (*WorkPool, error) {
	return New(workers, 0)
}

func New(workers, pending int) (*WorkPool, error) {
	if workers < 1 || pending < 0 {
		return nil, fmt.Errorf(
			"must provide positive workers and non-negative pending; provided %d workers and %d pending",
			workers,
			pending,
		)
	}

	w := &WorkPool{
		workQ:      make(chan func(), workers+pending),
		stopping:   make(chan struct{}),
		maxWorkers: workers,
	}

	return w, nil
}

func (w *WorkPool) Submit(work func()) {
	if atomic.LoadInt32(&w.stopped) == 1 {
		return
	}

	select {
	case w.workQ <- work:
		if atomic.LoadInt32(&w.stopped) == 1 {
			w.drain()
		} else {
			w.addWorker()
		}
	case <-w.stopping:
	}
}

func (w *WorkPool) Stop() {
	if atomic.CompareAndSwapInt32(&w.stopped, 0, 1) {
		close(w.stopping)
		w.drain()
	}
}

func (w *WorkPool) Stats() (total int, active int) {
	w.mutex.Lock()
	total = w.numWorkers
	active = total - w.idleWorkers
	w.mutex.Unlock()

	return
}

func (w *WorkPool) addWorker() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.idleWorkers > 0 || w.numWorkers == w.maxWorkers {
		return false
	}

	w.numWorkers++
	go worker(w)
	return true
}

func (w *WorkPool) workerStopping(force bool) bool {
	w.mutex.Lock()
	if !force {
		if len(w.workQ) < w.numWorkers {
			w.mutex.Unlock()
			return false
		}
	}

	w.numWorkers--
	w.mutex.Unlock()

	return true
}

func (w *WorkPool) drain() {
	for {
		select {
		case <-w.workQ:
		default:
			return
		}
	}
}

func worker(w *WorkPool) {
	timer := time.NewTimer(waitTimeout)
	defer timer.Stop()

	for {
		if atomic.LoadInt32(&w.stopped) == 1 {
			w.workerStopping(true)
			return
		}

		select {
		case <-timer.C:
			if w.workerStopping(false) {
				return
			}
			timer.Reset(waitTimeout)

		case <-w.stopping:
			w.workerStopping(true)
			return

		case work := <-w.workQ:
			timer.Stop()

			w.mutex.Lock()
			w.idleWorkers--
			w.mutex.Unlock()

		NOWORK:
			for {
				work()
				select {
				case work = <-w.workQ:
				case <-w.stopping:
					break NOWORK
				default:
					break NOWORK
				}
			}

			w.mutex.Lock()
			w.idleWorkers++
			w.mutex.Unlock()

			timer.Reset(waitTimeout)
		}
	}
}
