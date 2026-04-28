package main

import (
	"context"
	"fmt"
	"time"
)

func StartPersistenceWorker(ctx context.Context, store MessageStore, queue <-chan ChatMessage, batchSize int, flushInterval time.Duration) {
	if store == nil || queue == nil || batchSize <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		batch := make([]ChatMessage, 0, batchSize)
		flush := func() {
			if len(batch) == 0 {
				return
			}
			if err := store.SaveMessages(ctx, batch); err != nil {
				fmt.Println("persist save err:", err)
			}
			batch = batch[:0]
		}

		for {
			select {
			case <-ctx.Done():
				flush()
				return
			case msg, ok := <-queue:
				if !ok {
					flush()
					return
				}
				batch = append(batch, msg)
				if len(batch) >= batchSize {
					flush()
				}
			case <-ticker.C:
				flush()
			}
		}
	}()
}
