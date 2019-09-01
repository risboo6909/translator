package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/text/language"
)

const (
	cacheSizeBytes      = 5 * 1024 * 1024 // 5Mb cache
	cacheEntryExpireSec = 60 * 60         // 1 Hour
)

func main() {
	ctx := context.Background()
	rand.Seed(time.Now().UTC().UnixNano())

	s := NewService(
		"http://localhost:33333",
		func(r requestCtx) string {
			return fmt.Sprintf("from=%s&to=%s&text=%s", r.from, r.to, r.data)
		},
		cacheSizeBytes,
		cacheEntryExpireSec)

	go func() {
		fmt.Println(s.translator.Translate(ctx, language.English, language.Japanese, "test"))
	}()

	go func() {
		fmt.Println(s.translator.Translate(ctx, language.English, language.Japanese, "test"))
	}()

	fmt.Println(s.translator.Translate(ctx, language.English, language.Japanese, "trololo"))
	fmt.Println(s.translator.Translate(ctx, language.English, language.Japanese, "test"))
}
