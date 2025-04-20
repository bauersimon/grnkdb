# GRNKdb

Gronkh game database @ [grnkdb.dev](https://grnkdb.dev/).

## What does this do?

Attempt to semi-automate scraping and cataloging all video games that [Gronkh](https://www.youtube.com/gronkh) ever played. The website is updated every Saturday with respect to the newest YouTube videos.

## How do I use it?

- Check out [grnkdb.dev](https://grnkdb.dev/).
- Download the list at [grnkdb.dev/data.csv](https://grnkdb.dev/data.csv).
- Try the scraper locally by cloning the repo and running `go run main.go` (requires [Go](https://go.dev/) and a [YouTube API](https://developers.google.com/youtube/v3/getting-started) key).

## What state are we at?

Proof-of-concept / alpha!

- Currently only covers the main YouTube channel (side-channels and [gronkh.tv](https://gronkh.tv/) missing).
- Scraping algorithm is work-in-progress so the list looks quite ugly.
- Website contains the bare minimum (i.e. no search functionality, etc...).
