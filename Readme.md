Run server from test_server folder and run translator in another terminal window to see how it works.

test_server is a python script which which reverses its input and responds with the random delay from 0 to 5 seconds.

Main features:

1. retry requests N times with exponential back off before failing with an error
2. cache request results for in the storage to avoid charges for the same queries (I use freecache library for caching, assuming there will be only one instance of the application running because there is no opposite requirement in the assignment, in case of many instances running, external in-memory storage should be used instead of internal cache, e.g. Redis, etc)
3. deduplicate simultaneous queries for the same parameters to avoid charges for same query burst
