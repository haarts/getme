GetMe downloads your favourite TV shows in the easiest way possible.

Just run:
```
$ getme "American Dad"
```

And you're done!

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

## Help

For more help run:

```
$ getme help
```

## Tools

The `tools` directory contains a Python 3 script to create a list of popular
shows. This is used to present the user with the most relevant search results.
As a regular user you don't need to use this file. A recent list of shows in
compiled in the binaries.
