package main

import (
	"fmt"
)

// Service is a Translator user.
type Service struct {
	translator Translator
}

func NewService() *Service {
	//t := newRandomTranslator(
	//	100*time.Millisecond,
	//	500*time.Millisecond,
	//	0.1,
	//)

	return &Service{
		translator: newTranslator("http://localhost:33333", func(key requestCtx) string {
			return fmt.Sprintf("from=%s&to=%s&text=%s", key.from, key.to, key.data)
		}, 10000),
	}
}
