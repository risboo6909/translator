package main

// Service is a Translator user.
type Service struct {
	translator Translator
}

func NewService(url string, formatter func(key requestCtx) string, cacheSizeBytes, cacheEntryExpireSec int) *Service {
	//t := newRandomTranslator(
	//	100*time.Millisecond,
	//	500*time.Millisecond,
	//	0.1,
	//)

	return &Service{
		translator: newTranslator(url, formatter, cacheSizeBytes, cacheEntryExpireSec),
	}
}
