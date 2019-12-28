# Name

Leon Chou

# Building

`$ cd collatz && go build .`
`$ cd collatz-cachinng && go build .`

# Running

`$ cd collatz && go build . && ./collatz`
`$ cd collatz-cachinng && go build . && ./collatz-caching`

# performance difference

```
collatz-caching $ time go run .
Longest sequence starts at 8400511, length 686
go run .  61.15s user 18.21s system 275% cpu 28.849 total
collatz-caching $ time go run .
Longest sequence starts at 8400511, length 686
go run .  58.70s user 15.19s system 270% cpu 27.351 total
```

```
collatz $ time go run .
Longest sequence starts at 8400511, length 685
go run .  26.10s user 7.63s system 290% cpu 11.623 total
collatz $ time go run .
Longest sequence starts at 8400511, length 685
go run .  25.94s user 8.81s system 291% cpu 11.935 total
```

There is a noticeable difference between the cached collatz and the uncached collatz runtimes.
Noticeably the cached version is actually twice as slow.

The issues probably arise from attempting to access the cache (locks), as well as the fact that we
iterate from smallest to largest.

This makes it very easy to miss the cache.

# Channel buffering

For non-cached, I used unbuffered channels so that threads wouldn't waste time doing work that they didn't need to do.

For cached, I initially attempted a mutex with a single buffered `chan bool`, but ended up using a `sync.RWLock` instead
in an attempt to improve performance.

# To further improve performance

Some things to consider that could affect the performance of the cached version

Maybe the Collatz length could be calculated in a way that better utilizes the cache?

Maybe it could be calculated from higher numbers first?

More workers would probably help.