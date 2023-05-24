// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package queue

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"

	"istio.io/pkg/log"
)

type RagTask struct {
	Task  func() error
	Type  string
	Start time.Time
}

// Task to be performed.
type Task func() error

// Instance of work tickets processed using a rate-limiting loop
type Instance interface {
	// Push a task.
	Push(task *RagTask)
	// Run the loop until a signal on the channel
	Run(<-chan struct{})

	// Closed returns a chan that will be signaled when the Instance has stopped processing tasks.
	Closed() <-chan struct{}

	IncrementType(typeObj string) int

	DecrementType(typeObj string) int
}

type queueImpl struct {
	delay           time.Duration
	tasks           []*RagTask
	cond            *sync.Cond
	closing         bool
	closed          chan struct{}
	closeOnce       *sync.Once
	id              string
	TypeCounterMap  map[string]int
	typeSyncCounter sync.Map
	queueLastInfoed time.Time
}

func (q *queueImpl) printQueue() {
	log.Infof("Starting evaluation of queue:")
	q.typeSyncCounter.Range(func(key, value interface{}) bool {
		log.Infof("%s ->  %d", key.(string), value.(int))
		return true
	})
	log.Infof("Finished evaluation of queue:")
}

func (q *queueImpl) IncrementType(typeObj string) int {
	return 0
}

func (q *queueImpl) DecrementType(typeObj string) int {
	return 0
}

// NewQueue instantiates a queue with a processing function
func NewQueue(errorDelay time.Duration) Instance {
	return NewQueueWithID(errorDelay, rand.String(10))
}

func NewQueueWithID(errorDelay time.Duration, name string) Instance {
	return &queueImpl{
		delay:          errorDelay,
		tasks:          make([]*RagTask, 0),
		closing:        false,
		closed:         make(chan struct{}),
		closeOnce:      &sync.Once{},
		cond:           sync.NewCond(&sync.Mutex{}),
		id:             name,
		TypeCounterMap: map[string]int{},
	}
}

func (q *queueImpl) Push(item *RagTask) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if !q.closing {
		q.tasks = append(q.tasks, item)
	}
	q.cond.Signal()
}

func (q *queueImpl) Closed() <-chan struct{} {
	return q.closed
}

// get blocks until it can return a task to be processed. If shutdown = true,
// the processing go routine should stop.
func (q *queueImpl) get() (task *RagTask, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	// wait for closing to be set, or a task to be pushed
	for !q.closing && len(q.tasks) == 0 {
		q.cond.Wait()
	}

	if q.closing {
		// We must be shutting down.
		return nil, true
	}
	task = q.tasks[0]
	// Slicing will not free the underlying elements of the array, so explicitly clear them out here
	q.tasks[0] = nil
	q.tasks = q.tasks[1:]
	return task, false
}

func (q *queueImpl) processNextItem() bool {
	// Wait until there is a new item in the queue
	task, shuttingdown := q.get()
	if shuttingdown {
		return false
	}

	// Run the task.
	queue.With(nameTag.Value(task.Type)).Record(float64(len(q.tasks)))
	log.Infof("Dequeuing task %s , process time %d, queue length: %d", task.Type, time.Since(task.Start).Microseconds(), len(q.tasks))

	if task.Task == nil {
		log.Infof("Task came to be nil with type %s", task.Type)
		return true
	}

	if err := task.Task(); err != nil {
		delay := q.delay
		log.Infof("Work item handle failed (%v), retry after delay %v", err, delay)
		time.AfterFunc(delay, func() {
			q.Push(task)
		})
	}
	return true
}

func (q *queueImpl) Run(stop <-chan struct{}) {
	log.Debugf("started queue %s", q.id)
	defer func() {
		q.closeOnce.Do(func() {
			log.Debugf("closed queue %s", q.id)
			close(q.closed)
		})
	}()
	go func() {
		<-stop
		q.cond.L.Lock()
		q.cond.Signal()
		q.closing = true
		q.cond.L.Unlock()
	}()
	//ticker := time.NewTicker(20 * time.Second)
	//printMap := func() {
	//	log.Infof("Starting evaluation of queue:")
	//	q.typeSyncCounter.Range(func(key, value interface{}) bool {
	//		log.Infof("%s ->  %d", key.(string), value.(int))
	//		return true
	//	})
	//	log.Infof("Finished evaluation of queue:")
	//	fmt.Println()
	//}

	//go func() {
	//	for {
	//		select {
	//		case <-ticker.C:
	//			printMap()
	//		case <-stop:
	//			ticker.Stop()
	//			return
	//		}
	//	}
	//}()

	for q.processNextItem() {
	}
}
