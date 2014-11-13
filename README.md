# nameme

#### This was an afternoon hack as an excuse to play with Go. Caveat emptor!

Nameme is a distributed dictionary backed by a replicated log such as etcd.

Good for things like:
* Managing dynamic configuration
* Storing data available cluster-wide
* Data or computational sharding
* Service discovery
* Other stuff!

Bad for things that require multiple nodes to write to the same key at a high rate,
or really anything that needs super high throughput. This is mostly a convenience layer.

There is an abstraction interface for the log client so that any log can be used(ZooKeeper, consul, etc)
but currently only etcd is implemented.

Currently it only supports string -> string key value pairs because of no generics
in Go and I haven't explored other options for better value typing.

## Interface
#### NewDict(keyspace string, logClient LogClient) *Dict
Creates a new nameme.Dict and returns a pointer to it.
* *keyspace*: The keyspace in which to store data in the underlying log.
* *logClient*: The log client(see below)

#### Dict.Get(key string) (string, error)
Returns the *local* value at the given key.

#### Dict.Put(key string, value string) error
Updates the *local* value at the given key, updates the log, and returns any errors

#### NewEtcdClient(cluster []string) *EtcdClient
EtcdClient implements the LogClient interface.
* *cluster*: A list of host:port strings that represent the etcd cluster

## Example
```
client := NewEtcdClient([]string{"http://127.0.0.1:4001"})
myDict, err := NewDict("example", client)

myDict.Put("foo", "bar") // ignoring any error

val := myDict.Get("foo") // val == "bar"
```

## Running tests
Note that the tests assume a running etcd process at localhost:4001(the default)
and that they are non-deterministic because of race conditions and sleeps and I just
haven't made them better yet.