package cache

import (
	"context"
	"sync"

	"github.com/thought-machine/please/src/core"
)

// An asyncCache is a wrapper around a Cache interface that handles incoming
// store requests asynchronously and attempts to return immediately.
// The requests are handled on an internal queue, if that fills up then
// incoming requests will start to block again until it empties.
// Retrieval requests are still handled synchronously.
type asyncCache struct {
	requests  chan cacheRequest
	realCache core.Cache
	wg        sync.WaitGroup
}

// A cacheRequest models an incoming cache request on our queue.
type cacheRequest struct {
	target   *core.BuildTarget
	metadata *core.BuildMetadata
	key      []byte
	files    []string
	file     string
}

func newAsyncCache(realCache core.Cache, config *core.Configuration) core.Cache {
	c := &asyncCache{
		requests:  make(chan cacheRequest),
		realCache: realCache,
	}
	c.wg.Add(config.Cache.Workers)
	for i := 0; i < config.Cache.Workers; i++ {
		go c.run(context.Background())
	}
	return c
}

func (c *asyncCache) Store(ctx context.Context, target *core.BuildTarget, key []byte, metadata *core.BuildMetadata, files []string) {
	c.requests <- cacheRequest{
		target:   target,
		metadata: metadata,
		key:      key,
		files:    files,
	}
}

func (c *asyncCache) Retrieve(ctx context.Context, target *core.BuildTarget, key []byte, files []string) *core.BuildMetadata {
	return c.realCache.Retrieve(ctx, target, key, files)
}

func (c *asyncCache) Clean(ctx context.Context, target *core.BuildTarget) {
	c.realCache.Clean(ctx, target)
}

func (c *asyncCache) CleanAll(ctx context.Context) {
	c.realCache.CleanAll(ctx)
}

func (c *asyncCache) Shutdown() {
	log.Info("Shutting down cache workers...")
	close(c.requests)
	c.wg.Wait()
	log.Debug("Shut down all cache workers")
}

// run implements the actual async logic.
func (c *asyncCache) run(ctx context.Context) {
	for r := range c.requests {
		c.realCache.Store(ctx, r.target, r.key, r.metadata, r.files)
	}
	c.wg.Done()
}
