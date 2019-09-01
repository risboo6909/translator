Main features:

1. retry requests N times with exponential back off before failing with an error
2. cache request results for in the storage to avoid charges for the same queries (uses LRU cache)
3. deduplicate simultaneous queries for the same parameters to avoid charges for same query burst
