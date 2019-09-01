package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/text/language"

	"github.com/coocood/freecache"
	"github.com/pkg/errors"
)

const maxRetriesDelaySec = 8

type requestCtx struct {
	from, to language.Tag
	data     string
}

func (c requestCtx) asBytes() []byte {
	return []byte(fmt.Sprintf("%s-%s-%s", c.from, c.to, c.data))
}

type result struct {
	body []byte
	err  error
}

type MyTranslator struct {
	url       string
	formatter func(key requestCtx) string

	// there is no mention in problem description that backend may be scaled in future
	// therefor I suppose it runs as a single process on one machine and no external in-memory cache is needed
	cache               *freecache.Cache
	cacheEntryExpireSec int

	// queue to avoid many duplicate requests at once
	inProgress map[requestCtx][]chan result

	m sync.Mutex
}

func newTranslator(url string, formatter func(key requestCtx) string, cacheSizeBytes, cacheEntryExpireSec int) *MyTranslator {
	return &MyTranslator{
		url:                 url,
		formatter:           formatter,
		cache:               freecache.NewCache(cacheSizeBytes),
		cacheEntryExpireSec: cacheEntryExpireSec,
		inProgress:          make(map[requestCtx][]chan result),
	}
}

func doFetch(url string, timeout time.Duration) ([]byte, error) {

	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "error in doFetch")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error in doFetch")
	}

	return body, nil

}

func fetchWithBackOff(url string, r requestCtx) ([]byte, error) {

	var (
		retryDelay time.Duration = 1
		timeout    time.Duration = 1
		resp       []byte
		err        error
	)

	for retryDelay <= maxRetriesDelaySec {

		resp, err = doFetch(url, timeout*time.Second)

		if err != nil {

			log.Printf("Can't get translation for '%s' from %s to %s, reason: %v, retry in %d seconds",
				r.data, r.from, r.to, err, retryDelay)

			time.Sleep(retryDelay * time.Second)
			retryDelay *= 2

			// increase timeout by a second
			timeout++

		} else {
			break
		}

	}

	if err != nil {
		return nil, errors.Errorf("Unable to translate '%s' from %s to %s", r.data, r.from, r.to)
	}

	return resp, nil

}

func (t *MyTranslator) worker(key requestCtx) {

	resp, err := fetchWithBackOff(t.url+"?"+t.formatter(key), key)
	// don't update cache in case of any error to avoid invalid results until cache expiration
	if err == nil {
		t.cache.Set(key.asBytes(), resp, t.cacheEntryExpireSec)
	}

	// get all consumers waiting for translation
	t.m.Lock()
	consumers := t.inProgress[key]
	delete(t.inProgress, key)
	t.m.Unlock()

	// send response to all consumers
	for _, c := range consumers {
		c <- result{
			body: resp,
			err:  err,
		}
	}

}

func (t *MyTranslator) enqueue(key requestCtx) chan result {

	t.m.Lock()

	if _, ok := t.inProgress[key]; !ok {
		// if no fetcher spawned for the request - spawn one
		go t.worker(key)
	}

	// add new consumer for the request
	result := make(chan result)
	t.inProgress[key] = append(t.inProgress[key], result)

	t.m.Unlock()

	return result

}

func (t *MyTranslator) Translate(ctx context.Context, from, to language.Tag, data string) (string, error) {

	k := requestCtx{from, to, data}

	translated, err := t.cache.Get(k.asBytes())
	if err == nil {
		return string(translated), nil
	}

	r := <-t.enqueue(k)

	if r.err != nil {
		return "", r.err
	}

	return string(r.body), nil
}
