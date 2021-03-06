package feeder

import (
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-feeder/internal/domain"
)

// Service is the interface describing the feeder behavior
type Service interface {
	Start() error
	Stop()
	IsRunning() bool
}

type feederService struct {
	feeds    []domain.Feed
	targets  []domain.Target
	stopChan chan bool
	running  bool
	locker   sync.Locker
}

// NewFeeder is the factory function for feeder service
func NewFeeder(feeds []domain.Feed, targets []domain.Target) Service {
	return &feederService{
		feeds:    feeds,
		targets:  targets,
		stopChan: make(chan bool),
		running:  false,
		locker:   &sync.Mutex{},
	}
}

// Start observe all the feeds chan (using merge function)
// and push the results to all targets
func (t *feederService) Start() error {
	if t.IsRunning() {
		return errors.New("the feeder is already started")
	}

	t.running = true
	marketPriceChannel := merge(t.feeds...)

	for t.IsRunning() {
		select {
		case <-t.stopChan:
			t.running = false
			break
		case marketPrice := <-marketPriceChannel:
			log.Info("Market ", marketPrice.Market.BaseAsset[:4], "-", marketPrice.Market.QuoteAsset[:4], " | Base Price ", marketPrice.Price.BasePrice, " | Quote Price ", marketPrice.Price.QuotePrice)
			for index, target := range t.targets {
				target.Push(marketPrice)
				log.Debug("Pushed to target ", index)
			}
		}
	}

	return nil
}

func (t *feederService) Stop() {
	t.stopChan <- true
}

func (t *feederService) IsRunning() bool {
	t.locker.Lock()
	defer t.locker.Unlock()
	return t.running
}

// merge gathers several feeds into a unique channel
func merge(feeds ...domain.Feed) <-chan domain.MarketPrice {
	mergedChan := make(chan domain.MarketPrice)
	var wg sync.WaitGroup

	wg.Add(len(feeds))
	for _, feed := range feeds {
		c := feed.GetMarketPriceChan()
		go func(c <-chan domain.MarketPrice) {
			for marketPrice := range c {
				mergedChan <- marketPrice
			}
			wg.Done()
		}(c)
	}

	go func() {
		wg.Wait()
		close(mergedChan)
	}()

	return mergedChan
}
