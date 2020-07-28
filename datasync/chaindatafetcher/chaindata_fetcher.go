package chaindatafetcher

import (
	"github.com/klaytn/klaytn/blockchain"
	"github.com/klaytn/klaytn/event"
	"github.com/klaytn/klaytn/log"
	"github.com/klaytn/klaytn/networks/p2p"
	"github.com/klaytn/klaytn/networks/rpc"
	"github.com/klaytn/klaytn/node"
	"sync"
)

var logger = log.NewModuleLogger(log.ChainDataFetcher)

type ChainDataFetcher struct {
	config *ChainDataFetcherConfig

	blockchain *blockchain.BlockChain

	chainCh  chan blockchain.ChainEvent
	chainSub event.Subscription

	reqCh  chan uint64 // TODO-ChainDataFetcher add logic to insert new requests from APIs to this channel
	resCh  chan uint64
	stopCh chan struct{}

	numHandlers int

	wg sync.WaitGroup
}

func NewChainDataFetcher(ctx *node.ServiceContext, cfg *ChainDataFetcherConfig) (*ChainDataFetcher, error) {
	return &ChainDataFetcher{
		config:      cfg,
		chainCh:     make(chan blockchain.ChainEvent, cfg.BlockChannelSize),
		reqCh:       make(chan uint64, cfg.RequestChSize),
		resCh:       make(chan uint64, cfg.ResponseChSize),
		stopCh:      make(chan struct{}),
		numHandlers: cfg.NumHandlers,
	}, nil
}

func (f *ChainDataFetcher) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

func (f *ChainDataFetcher) APIs() []rpc.API {
	// TODO-ChainDataFetcher add APIs to start or stop chaindata fetcher
	return []rpc.API{}
}

func (f *ChainDataFetcher) Start(server p2p.Server) error {
	// launch multiple goroutines to handle new blocks
	for i := 0; i < f.numHandlers; i++ {
		f.wg.Add(1)
		go func() {
			defer f.wg.Done()
			f.handler()
		}()
	}

	// subscribe chain head event
	f.chainSub = f.blockchain.SubscribeChainEvent(f.chainCh)
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		f.loop()
	}()

	return nil
}

func (f *ChainDataFetcher) Stop() error {
	f.chainSub.Unsubscribe()
	close(f.stopCh)

	logger.Info("wait for all goroutines to be terminated...")
	f.wg.Wait()
	logger.Info("terminated all goroutines for chaindatafetcher")
	return nil
}

func (f *ChainDataFetcher) Components() []interface{} {
	return nil
}

func (f *ChainDataFetcher) SetComponents(components []interface{}) {
	for _, component := range components {
		switch v := component.(type) {
		case *blockchain.BlockChain:
			f.blockchain = v
		}
	}
}

func (f *ChainDataFetcher) handler() {
	for {
		select {
		case <-f.stopCh:
			logger.Info("stopped a handler")
			return
		case req := <-f.reqCh:
			// TODO-ChainDataFetcher do handle new request
			f.resCh <- req
		}
	}
}

func (f *ChainDataFetcher) loop() {
	for {
		select {
		case <-f.stopCh:
			logger.Info("stopped main loop for chaindatafetcher")
			return
		case ev := <-f.chainCh:
			num := ev.Block.NumberU64()
			f.reqCh <- num
			logger.Info("request to handle new block", "blockNumber", num)
		case res := <-f.resCh:
			f.updateCheckpoint(res)
			logger.Info("processed requested block", "blockNumber", res)
		}
	}
}

func (f *ChainDataFetcher) updateCheckpoint(num uint64) {
	// TODO-ChainDataFetcher add logic to update new checkpoint
}
