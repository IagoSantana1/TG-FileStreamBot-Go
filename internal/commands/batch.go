package commands

import (
	"sync"
	"time"

	"github.com/celestix/gotgproto/ext"
)

type BatchItem struct {
	Ctx    *ext.Context
	Update *ext.Update
}

type BatchQueue struct {
	mu sync.Mutex

	items []BatchItem

	timer *time.Timer
}

var (
	batchMutex sync.Mutex
	batches    = map[int64]*BatchQueue{}
)

const batchTimeout = 5 * time.Second

func getBatch(chatID int64) *BatchQueue {
	batchMutex.Lock()
	defer batchMutex.Unlock()

	q, ok := batches[chatID]
	if ok {
		return q
	}

	q = &BatchQueue{}
	batches[chatID] = q

	return q
}

func enqueueBatch(ctx *ext.Context, u *ext.Update, chatID int64) {

	queue := getBatch(chatID)

	queue.mu.Lock()
	defer queue.mu.Unlock()

	queue.items = append(queue.items, BatchItem{
		Ctx:    ctx,
		Update: u,
	})

	if queue.timer != nil {
		queue.timer.Stop()
	}

	queue.timer = time.AfterFunc(batchTimeout, func() {
		flushBatch(chatID, queue)
	})
}

func flushBatch(chatID int64, queue *BatchQueue) {

	queue.mu.Lock()

	items := queue.items

	queue.items = nil

	if queue.timer != nil {
		queue.timer.Stop()
		queue.timer = nil
	}

	queue.mu.Unlock()

	if len(items) == 0 {
		return
	}

	processBatch(chatID, items)
}
