package commands

import (
	"sync"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

type BatchItem struct {
	Ctx    *ext.Context
	Update *ext.Update
}

type BatchQueue struct {
	mu sync.Mutex

	items []BatchItem
	timer *time.Timer

	loadingCtx   *ext.Context
	loadingMsgID int
}

var (
	batchMutex sync.Mutex
	batches    = map[int64]*BatchQueue{}
)

const batchTimeout = 3 * time.Second

func getBatch(chatID int64) *BatchQueue {
	batchMutex.Lock()
	defer batchMutex.Unlock()

	if q, ok := batches[chatID]; ok {
		return q
	}

	q := &BatchQueue{}
	batches[chatID] = q

	return q
}

func enqueueBatch(ctx *ext.Context, u *ext.Update, chatID int64) {

	queue := getBatch(chatID)

	queue.mu.Lock()
	defer queue.mu.Unlock()

	// Primeiro arquivo do lote
	if len(queue.items) == 0 {

		queue.loadingCtx = ctx

		msg, err := ctx.SendMessage(chatID, &tg.MessagesSendMessageRequest{
			Message: "🕒 Agrupando arquivos do lote...",
		})

		if err == nil {
			queue.loadingMsgID = msg.ID
		}
	}

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
	loadingCtx := queue.loadingCtx
	loadingMsgID := queue.loadingMsgID

	queue.items = nil
	queue.loadingCtx = nil
	queue.loadingMsgID = 0

	if queue.timer != nil {
		queue.timer.Stop()
		queue.timer = nil
	}

	queue.mu.Unlock()

	if len(items) == 0 {
		return
	}

	processBatch(chatID, items, loadingCtx, loadingMsgID)
}
