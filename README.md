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

What GetMe **doesn't** do is actually download the files. That is the job of your
Bittorrent client. All most all clients support 'watch directories'. When a
torrent shows up in that directory the client will process it.

## Installation
Couldn't be simpler. No external dependancies, no hassle. Either grab on of the
binaries or build from source.

### Binaries
Check out the [releases](https://github.com/haarts/getme/releases).

### Bleeding edge
You need to have [Go](golang.org) installed. Then run:

```
$ go get github.com/haarts/getme
```

I've been using version 1.4 but I'm fairly certain every 1.x version of Go will
work.

## Usage

GetMe supports two modes. Adding shows/movies with `-a`. And updating them
with, you guessed it, `-u`.

Usually you'd added a couple of shows and then periodically run (cron anyone?)
with the update flag.

### First time
The first time that you run GetMe it will exit immediately because no config
file could be found. GetMe will create one for you. 
This file contains ONE line with a simple key value pair. This pair will tell
GetMe where the watch directory is of your favourite Torrent client. 
You really want to check if the directory is the correct one.

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


## Why?!

There are, of course, tools which do what GetMe does. Internet is great like
that. One notable example is [FlexGet](flexget.com). This is a great tool with
a great community. It has an endless list of features and options. Which is
exactly why I wrote GetMe. I just wanted the job to get done without needing to
care about all the nitty gritty details. If you _really_ want everything in
1080p from a specific release group GetMe is not for you. If you just want the
job done: `./getme -a 'Pioneer one`.
