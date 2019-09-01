package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/text/language"

	"github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

const maxRetriesDelaySec = 8

type requestCtx struct {
	from, to language.Tag
	data     string
}

type result struct {
	text string
	err  error
}

type MyTranslator struct {
	url           string
	paramsBuilder func(key requestCtx) string

	// there is no mention in problem description that backend may be scaled in future
	// therefor I suppose it runs as a single process on one machine and no external in-memory cache is needed
	cache *lru.ARCCache

	// queue to avoid many duplicate requests at once
	inProgress map[requestCtx][]chan result

	m sync.Mutex
}

func newTranslator(url string, paramsBuilder func(key requestCtx) string, cacheCapacity int) *MyTranslator {
	c, _ := lru.NewARC(cacheCapacity)
	return &MyTranslator{
		url:           url,
		paramsBuilder: paramsBuilder,
		cache:         c,
		inProgress:    make(map[requestCtx][]chan result),
	}
}

func doFetch(url string, timeout time.Duration) (string, error) {

	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "error in doFetch")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "error in doFetch")
	}

	return string(body), nil

}

func fetchWithBackOff(url string, key requestCtx) (string, error) {

	var (
		retryDelay time.Duration = 1
		timeout    time.Duration = time.Second
		resp       string
		err        error
	)

	for retryDelay <= maxRetriesDelaySec {

		resp, err = doFetch(url, timeout)

		if err != nil {

			log.Printf("Can't get translation for '%s' from %s to %s, reason: %v, retry in %d seconds",
				key.data, key.from, key.to, err, 5)

			time.Sleep(retryDelay * time.Second)
			retryDelay *= 2

			// increase timeout by a second
			timeout++

		} else {
			break
		}

	}

	if err != nil {
		return "", errors.Errorf("Unable to translate '%s' from %s to %s", key.data, key.from, key.to)
	}

	return resp, nil

}

func (t *MyTranslator) worker(key requestCtx) {

	resp, err := fetchWithBackOff(t.url+"?"+t.paramsBuilder(key), key)
	if err == nil {
		t.cache.Add(key, resp)
	}

	// get all consumers waiting for translation
	t.m.Lock()
	consumers := t.inProgress[key]
	delete(t.inProgress, key)
	t.m.Unlock()

	// send response to all consumers
	for _, c := range consumers {
		c <- result{
			text: resp,
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

	translated, found := t.cache.Get(k)
	if found {
		return translated.(string), nil
	}

	r := <-t.enqueue(k)

	if r.err != nil {
		return "", r.err
	}

	return r.text, nil
}
