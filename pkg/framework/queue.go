package framework

import (
	"container/heap"
	"sync"
	"time"
)

type t interface{}

// Queue abstract queue
type Queue interface {
	Add(item interface{})
	Get() (item interface{}, shutdown bool)
	ShutDown()
	ShuttingDown() bool
}

type queue struct {
	q            []t
	cond         *sync.Cond
	shuttingDown bool
}

func newQueue() *queue {
	return &queue{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (w *queue) Add(item interface{}) {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	if w.shuttingDown {
		return
	}

	w.q = append(w.q, item)
	w.cond.Signal()
}

func (w *queue) Get() (item interface{}, shutdown bool) {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	for len(w.q) == 0 && !w.shuttingDown {
		w.cond.Wait()
	}
	if len(w.q) == 0 {
		return nil, true
	}

	item, w.q = w.q[0], w.q[1:]

	return item, false
}

func (w *queue) ShutDown() {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	w.shuttingDown = true
	w.cond.Broadcast()
}

func (w *queue) ShuttingDown() bool {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	return w.shuttingDown
}

///////////////    workQueue   /////////////////

type priorityQueue []*waitItem

func (pq priorityQueue) Len() int {
	return len(pq)
}
func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].readyAt.Before(pq[j].readyAt)
}
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push adds an item to the queue
func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*waitItem)
	item.index = n
	*pq = append(*pq, item)
}

// Pop return the item and than remove it
func (pq *priorityQueue) Pop() interface{} {
	n := len(*pq)
	item := (*pq)[n-1]
	item.index = -1
	*pq = (*pq)[0:(n - 1)]
	return item
}

// Peek just return the first element of the queue
func (pq priorityQueue) Peek() interface{} {
	return pq[0]
}

// waitItem wrap the data and readyAt indicates this elements should be added at this time
type waitItem struct {
	data    t
	readyAt time.Time
	// index in the priority queue
	index int
}

type workQueue struct {
	Queue

	// waitingForAddCh
	waitingForAddCh chan *waitItem

	stopOnce sync.Once
	stopCh   chan struct{}
}

// newWorkQueue
func newWorkQueue(buffer uint) *workQueue {
	w := &workQueue{
		Queue:           newQueue(),
		stopCh:          make(chan struct{}),
		waitingForAddCh: make(chan *waitItem, buffer),
	}

	go w.waitLoop()

	return w
}

// AddAfter
func (w *workQueue) AddAfter(item interface{}, duration time.Duration) {
	if w.ShuttingDown() {
		return
	}

	// add item immediately
	if duration <= 0 {
		w.Add(item)
		return
	}

	select {
	case <-w.stopCh:
		// it will be added to queue at readyAt time moment
	case w.waitingForAddCh <- &waitItem{data: item, readyAt: time.Now().Add(duration)}:
	}
}

// ShutDown
func (w *workQueue) ShutDown() {
	w.stopOnce.Do(func() {
		w.Queue.ShutDown()
		close(w.stopCh)
	})
}

// waitLoop handle data waiting to be enqueued
func (w *workQueue) waitLoop() {

	never := make(<-chan time.Time)

	var nextReadyTimer *time.Timer

	waitingForQueue := &priorityQueue{}
	heap.Init(waitingForQueue)

	// prevent repeat data but different readyAt,we should update same data's time
	waitingEntryByData := map[t]*waitItem{}

	for {
		if w.Queue.ShuttingDown() {
			return
		}

		now := time.Now()

		// add ready data
		for waitingForQueue.Len() > 0 {
			entry := waitingForQueue.Peek().(*waitItem)
			if entry.readyAt.After(now) {
				break
			}

			entry = heap.Pop(waitingForQueue).(*waitItem)
			w.Add(entry.data)
			delete(waitingEntryByData, entry.data)
		}

		// set timer because we should notice the first waiting item
		nextReadyAt := never
		if waitingForQueue.Len() > 0 {
			if nextReadyTimer != nil {
				nextReadyTimer.Stop()
			}
			entry := waitingForQueue.Peek().(*waitItem)
			nextReadyTimer = time.NewTimer(entry.readyAt.Sub(now))
			nextReadyAt = nextReadyTimer.C
		}

		select {
		case <-w.stopCh:
			return

		case <-nextReadyAt:

		case waitEntry := <-w.waitingForAddCh:
			// 1.put data into the waitingForQueue when time is not up yet
			if waitEntry.readyAt.After(time.Now()) {
				insert(waitingForQueue, waitingEntryByData, waitEntry)
			} else {
				// 2.add data to the queue when time is up
				w.Add(waitEntry.data)
			}

			// drain entry design
			drained := false
			for !drained {
				select {
				case waitEntry := <-w.waitingForAddCh:
					if waitEntry.readyAt.After(time.Now()) {
						insert(waitingForQueue, waitingEntryByData, waitEntry)
					} else {
						w.Add(waitEntry.data)
					}
				default:
					drained = true
				}
			}
		}
	}
}

func insert(q *priorityQueue, knownEntries map[t]*waitItem, entry *waitItem) {
	// update same data's time
	existing, exists := knownEntries[entry.data]
	if exists {
		if existing.readyAt.After(entry.readyAt) {
			existing.readyAt = entry.readyAt
			// reorder the queue's elements
			heap.Fix(q, existing.index)
		}

		return
	}

	heap.Push(q, entry)
	knownEntries[entry.data] = entry
}
