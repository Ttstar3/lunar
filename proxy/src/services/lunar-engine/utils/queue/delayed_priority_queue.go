package queue

import (
	"container/heap"
	"lunar/toolkit-core/clock"
	"lunar/toolkit-core/logging"
	"sync"
	"time"
)

var epochTime = time.Unix(0, 0)

type Strategy struct {
	WindowQuota int
	WindowSize  time.Duration
}

type DelayedPriorityQueue struct {
	strategy             Strategy
	currentWindowCounter int
	currentWindowEndTime time.Time
	requestCounts        map[int]int
	mutex                sync.RWMutex
	queue                PriorityQueue
	clock                clock.Clock
	cl                   logging.ContextLogger
}

func NewDelayedPriorityQueue(
	strategy Strategy,
	clock clock.Clock,
	cl logging.ContextLogger,
) *DelayedPriorityQueue {
	dpq := &DelayedPriorityQueue{ //nolint:exhaustruct
		strategy:      strategy,
		cl:            cl.WithComponent("delayed-priority-queue"),
		requestCounts: map[int]int{},
		clock:         clock,
	}

	heap.Init(&dpq.queue)
	dpq.ensureWindowIsUpdated()
	go dpq.process()

	return dpq
}

func (dpq *DelayedPriorityQueue) Enqueue(
	req *Request,
	ttl time.Duration,
) bool {
	dpq.mutex.Lock()

	dpq.cl.Logger.Trace().Str("requestID", req.ID).
		Msgf("Enqueueing request, currentWindowCounter: %d, windowQuota: %d",
			dpq.currentWindowCounter, dpq.strategy.WindowQuota)

	dpq.ensureWindowIsUpdated()

	// Requests are processed in current window, if quota allows for it
	if dpq.currentWindowCounter < dpq.strategy.WindowQuota {
		dpq.currentWindowCounter++
		dpq.mutex.Unlock()
		close(req.doneCh)
		dpq.cl.Logger.Trace().
			Str("requestId", req.ID).
			Msg("Request processed in current window")

		return true
	}

	dpq.cl.Logger.Trace().Str("requestID", req.ID).
		Msgf("Sending request to be processed in queue")
	heap.Push(&dpq.queue, req)
	dpq.requestCounts[req.priority]++

	dpq.mutex.Unlock()

	// Wait until request is processed or TTL expires
	select {
	case <-req.doneCh:
		dpq.cl.Logger.Trace().
			Str("requestID", req.ID).
			Msgf("Request processing completed")
		dpq.mutex.Lock()
		defer dpq.mutex.Unlock()
		dpq.requestCounts[req.priority]--
		return true
	case <-dpq.clock.After(ttl):
		dpq.cl.Logger.Trace().Str("requestID", req.ID).
			Msgf("Request TTLed (now: %+v, ttl: %+v)", dpq.clock.Now(), ttl)
		dpq.mutex.Lock()
		defer dpq.mutex.Unlock()
		dpq.requestCounts[req.priority]--
		return false
	}
}

func (dpq *DelayedPriorityQueue) Counts() map[int]int {
	dpq.mutex.RLock()
	defer dpq.mutex.RUnlock()
	return deepCopyMap(dpq.requestCounts)
}

func deepCopyMap(m map[int]int) map[int]int {
	result := map[int]int{}
	for k, v := range m {
		result[k] = v
	}
	return result
}

func (dpq *DelayedPriorityQueue) getTimeTillWindowEnd() time.Duration {
	dpq.mutex.RLock()
	res := dpq.currentWindowEndTime.Sub(dpq.clock.Now())
	dpq.mutex.RUnlock()

	return res
}

func (dpq *DelayedPriorityQueue) ensureWindowIsUpdated() {
	currentTime := dpq.clock.Now()
	elapsedTime := currentTime.Sub(epochTime)
	currentWindowStartTime := epochTime.Add(
		(elapsedTime / dpq.strategy.WindowSize) * dpq.strategy.WindowSize,
	)
	updatedWindowEndTime := currentWindowStartTime.Add(dpq.strategy.WindowSize)
	if updatedWindowEndTime.After(dpq.currentWindowEndTime) {
		dpq.currentWindowCounter = 0
		dpq.currentWindowEndTime = updatedWindowEndTime
	}
}

func (dpq *DelayedPriorityQueue) process() {
	for {
		<-dpq.clock.After(dpq.getTimeTillWindowEnd())
		dpq.mutex.Lock()
		dpq.ensureWindowIsUpdated()
		dpq.processQueueItems()
		dpq.mutex.Unlock()
	}
}

func (dpq *DelayedPriorityQueue) processQueueItems() {
	for dpq.queue.Len() > 0 &&
		dpq.currentWindowCounter < dpq.strategy.WindowQuota {
		req, valid := heap.Pop(&dpq.queue).(*Request)
		if !valid {
			dpq.cl.Logger.Error().
				Msg("Could not cast priorityQueue item as Request, " +
					"will not process")
			continue
		}
		dpq.cl.Logger.Trace().
			Str("requestID", req.ID).
			Msgf("Attempt to process queued request")
		select {
		case req.doneCh <- struct{}{}:
			close(req.doneCh)
			dpq.currentWindowCounter++
			dpq.cl.Logger.Trace().Str("requestID", req.ID).
				Msgf("notified successful request processing to req.doneCh")
		default:
			dpq.cl.Logger.Trace().Str("requestID", req.ID).
				Msgf("req.doneCh already closed")
		}
		dpq.cl.Logger.Trace().Msgf("request %s processed in queue", req.ID)
	}
}
