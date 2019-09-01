# PaulCamper offline challenge

## Application process

To apply you solution in the most suitable way please use following process:
* create a git repository
* commit all initial files in this repo
* iterate on the solution and commit changes step by step (one commit per logical change set) with descriptive one line comments
* send resulting repo to your PaulCamper tech contact person in a 24 hours

You will find 3 parts of the task below in this Readme file. Parts are different in complexity.
They can be treated as separate and implemented one by one starting from top to bottom.
Patial solutions also accepted.

All commits withing 24 hours after challenge been shared are considered as a part of accepted solution.

## Domain

We are working on the application which uses external translation service developed by 3rd party. We have translation service interface (Translator) and implementation (randomTranslator) which is used for testing.

We know that translation service calls:
* can take long time
* sometimes fails
* costs us money

Also translations known to change sometimes but business approved caching them for at least N hours.

## Task

Implement solution that will properly handle external translation service
1. retry requests N times with exponential back off before failing with an error
2. cache request results for in the storage to avoid charges for the same queries (use simplest inmemory storage)
3. deduplicate simultaneous queries for the same parameters to avoid charges for same query burst

Cover new functionality with tests.

## Source code

translator.go and main.go should not be modified. Please use service.go and any new files for the solution.