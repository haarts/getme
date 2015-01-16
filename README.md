GetMe downloads your favourite TV shows in the easiest way possible.
[![GoDoc](https://godoc.org/github.com/haarts/getme?status.svg)](https://godoc.org/github.com/haarts/getme) 

Just run:
```
$ getme -a "Pioneer one"
```

And you're done!

## What?

GetMe allows you to:

1. Find shows
2. Download the appropriate torrents
3. Continue to follow the show

All from a simple CLI interface.

## Installation
Couldn't be simpler. No external dependancies, no hassle. Either grab on of the
binaries or build from source.

### Binaries
Check out the releases.

### Bleeding edge
You need to have [Go](golang.org) installed. Then run:

```
$ go get github.com/haarts/getme
```

I've been using version 1.4 but I'm fairly certain every 1.x version of Go will
work.

## Usage

Currently GetMe supports two modes. Adding shows/movies with `-a`. And updating
them with, you guessed it, `-u`.

Usually you'd added a couple of shows and then periodically run (cron anyone?)
with the update flag.

## Help

For more help (there isn't any but what the heck) run:

```
$ getme -h
```

## Tools

The `tools` directory contains a Python 3 script to create a list of popular
shows. This is used to present the user with the most relevant search results.
As a regular user you don't need to use this file. A recent list of shows in
compiled in the binaries.
