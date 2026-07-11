package schedular

import (
	"context"
	"log"
	"time"

	"github.com/DGreegman/gohunt/internal/fetcher"
	"github.com/robfig/cron/v3"
)

// Start launches a cron scheduler that fetches jobs on a fixed interval.
// It runs in the background and does not block.

func Start(source fetcher.JobSource, store *fetcher.Store, spec string) (*cron.Cron, error){
	c := cron.New()

	_, err := c.AddFunc(spec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

		defer cancel()
		log.Println("Schedular: starting fetch")
		fetched, inserted, err := fetcher.FetchAndStore(ctx, source, store)
		if err != nil {
			log.Printf("Schedular: fetch failed %v", err)
			return
		}
		log.Printf("Schedular: fetched %d, inserted %d new", fetched, inserted)
	})
	if err != nil {
		return nil, err 
	}
	c.Start()
	return c, nil
}