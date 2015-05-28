package workpool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const waitTimeout = 5 * time.Second

type WorkPool struct {
	workQueue chan func()
	stopping  chan struct{}
	stopped   int32

	mutex       sync.Mutex
	maxWorkers  int
	numWorkers  int
	idleWorkers int
}

func NewWorkPool(maxWorkers int) (*WorkPool, error) {
	return newWorkPoolWithPending(maxWorkers, 0)
}

func newWorkPoolWithPending(maxWorkers, pending int) (*WorkPool, error) {
	if maxWorkers < 1 || pending < 0 {
		return nil, fmt.Errorf(
			"must provide positive maxWorkers and non-negative pending; provided %d maxWorkers and %d pending",
			maxWorkers,
			pending,
		)
	}

	w := &WorkPool{
		workQueue:  make(chan func(), maxWorkers+pending),
		stopping:   make(chan struct{}),
		maxWorkers: maxWorkers,
	}

	return w, nil
}

func (w *WorkPool) Submit(work func()) {
	if atomic.LoadInt32(&w.stopped) == 1 {
		return
	}

	select {
	case w.workQueue <- work:
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
		if len(w.workQueue) < w.numWorkers {
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
		case <-w.workQueue:
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

		case work := <-w.workQueue:
			timer.Stop()

			w.mutex.Lock()
			w.idleWorkers--
			w.mutex.Unlock()

		NOWORK:
			for {
				work()
				select {
				case work = <-w.workQueue:
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
