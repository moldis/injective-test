# Home Assignment

We want to build a price service that provides its users with the latest BTC price

The source information can be found in this API
https://api.coindesk.com/v1/bpi/currentprice.json

## Part 1
Build a service that provides an endpoint where users can subscribe and get updates of BTC/USD price every 5s

On each update, the user should get these **exact** fields:
- `timedate`: Time date of the price
- `price`: The USD price of BTC

## Part 2
Users have complained that if they lose connection and reconnect after more than 5s, they lose
some updates, so now we want to allow the API user to pass a starting time date, so they can
receive all updates they may have missed before they're back in sync.

**Note**: As our source API doesn't provide historical data, you can assume that we will provide
this functionality since the moment we deploy this new version, no need to backfill data.

## Part 3
We've just expand operations to Europe, and those users would like to get the price in EUR,
update the API so that people can choose to receive either USD, EUR or both

## Part 4
The service is now very successful, and we want to scale it out and be able to run multiple instances transparently
to the user

## Part 5
We are in a very mature stage of the project, we want to make this project very robust, can you help us out explaining
how can we achieve this? Which architectural changes or patterns would it need?

**Note:** There's no need to code this part, you can edit this file and add your thoughts and ideas on what should we
do at this stage.

As an example, some of our concerns at this stage are:
- We want to make sure that it run smoothly and detect failures soon
- We're scared that our success will bring malicious actors that will attack us
- ... And even if traffic is legit, how can we survive a huge spike or mitigate the problem associated to it
- Sometimes our source data is not available, so users effectively stop receiving updates, which is very confusing
  for them as they see gaps of information even if they reconnect, can you suggest usability changes to deal with
  this scenario?

## Requirements
- [ ] Code must be written in Go
- [ ] While we don't want a 100% coverage, a reasonable test strategy is expected, although for time purposes, feel free
  to skip the implementation of certain testcases
- [ ] The source code should be provided as a git repository (either upload to a 3rd party provider or zip it including
  the .git folder)
- [ ] Build each part of the exercise incrementally: add a git tag for each part (you can have more commits)
- [ ] Either document your choices in the code or provide a file explaining your decisions

### Optional
- [ ] Dockerize the application
- [ ] Use gRPC for the streaming API
- [ ] Provide a CLI SDK to connect and consume the API