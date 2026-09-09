package cron

import (
	"log"
	"time"

	"eventifyApi/services"
)

type Worker struct {
	orderService services.OrderService
	stopChan     chan struct{}
}

func NewWorker(orderService services.OrderService) *Worker {
	return &Worker{
		orderService: orderService,
		stopChan:     make(chan struct{}),
	}
}

func (w *Worker) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	log.Printf("[Cron Worker] Started: checking for expired orders every %v", interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				w.executeOrderExpiry()
			case <-w.stopChan:
				ticker.Stop()
				log.Println("[Cron Worker] Stopped")
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	close(w.stopChan)
}

func (w *Worker) executeOrderExpiry() {
	affected, err := w.orderService.ExpireStaleOrders()
	if err != nil {
		log.Printf("[Cron Worker] Error expiring orders: %v", err)
		return
	}
	if affected > 0 {
		log.Printf("[Cron Worker] Successfully expired %d stale pending order(s) and released quota", affected)
	}
}
