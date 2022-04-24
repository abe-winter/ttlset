# ttlset

Is a stateful REST server that exposes a set datatype with per-key TTL. It's designed for hosting a persistent counter with per-item expiration semantics.

## status

I'm the only one using this. would be very surprised if something like this doesn't exist in an established cache DB. Ideally this would be a [redis module](https://redis.io/docs/modules/) I think.

## usage

- `go run ./cmd/ttlsetd`
- `go build ./cmd/ttlsetd`
- `go test ./pkg/ttlset`

(todo: publish a swagger spec + some examples)

in the meantime check out ttlsetd/main.go, which defines the REST routes, or curl.sh, which exercises the API. File a ticket if you want to use this and have questions.

## configs

(to grep for these, grep for lowercase w/out `TTLSET_` prefix)

- `TTLSET_TICKER_INTERVAL=10s` would set the ttl cull sweep to run every 10 seconds
- `TTLSET_DEFAULT_TTL=1m` would set the default item ttl of every set to 1 minute
- `TTLSET_TREE_ORDER=3` sets default btree order (= how many children per node I think)
- `TTLSET_ACCOUNTS="key1:rw key2:ro"` would create two accounts, with key1 + key2 as API keys, one with read-write permissions, the other read-only

## security

This has accounts + roles (see TTLSET_ACCOUNTS above) but doesn't set gin/contrib/secure and is generally assumed to be running inside a firewall (or better yet on my laptop). If you need this to be internet-facing file an issue.

# boring internals

## TtlSet datastructure

- Byval `map[string]TtlValue` (basically `map[string]Time`). the string key here is a set key (i.e. the thing you're storing the set)
- Bytime is a [gods](https://pkg.go.dev/github.com/emirpasic/gods) `btree.Tree` which stores `Time -> []string` (and the `[]string` is an array of keys for that time; unless you have lots of same-instant writes, most slices will be length-1)

Bytime is used to quickly find the oldest keys so that TTL sweeps are quick.

## approach to locking

is sort of bananas

each TtlSet has:
- `valLock RWMutex` managing reads and edits to Byval (the forward map)
- `timeLock Mutex` protecting access to Bytime (the reverse btree). this isn't an RWMutex because most accesses to the tree will involve a write

plus globally there's `keyMutexes`, an array of mutexes indexed with the `getMutex` function using the hash of the string keys.

the point of the mutex array is:
1. to provide transactional locking around the two-phase edit to a key's entry in the forward map (`map[string]Time`, more or less) and the reverse map (btree of `Time` -> `[]key`)
1. while enabling a higher degree of concurrency than locking on the whole TtlSet

I feel uncomfortable with the complexity here, and would like some kind of declarative proof or model of it to establish 1) that it increases concurrency in practice and 2) that it protects what it's supposed to protect.

simple summary is maybe:
- valLock + timeLock each protect access to a datastructure
- keyMutexes protects the two-phase commit i.e. a specific key that exists in both datastructures

Inability to time out locks and backpressure clients is a problem.
