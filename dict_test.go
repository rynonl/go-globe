package globe

import (
  "testing"
  "github.com/coreos/go-etcd/etcd"
)

var testCluster = []string{"http://127.0.0.1:4001"}
var rawClient = etcd.NewClient(testCluster)

func initNewTestDict(keyspace string) *Dict {
  rawClient.Delete(keyspace, true)
  rawClient.Set(keyspace + "/key1", "asdf", uint64(0))
  rawClient.Set(keyspace + "/key2", "lkjh", uint64(0))

  logClient := NewEtcdClient(testCluster)
  newD, err := NewDict(keyspace, logClient)

  if err != nil {
    panic(err)
  }

  return newD
}

func TestDictInitial(t *testing.T) {
  globeDict := initNewTestDict("TestDictInitial")

  if val, _ := globeDict.Get("key1"); val != "asdf" {
    t.Error("Expected key1 to be asdf but was", val)
  }
}

func TestDictWatchExisting(t *testing.T) {
  globeDict := initNewTestDict("TestDictWatchExisting")

  rawClient.Set("TestDictWatchExisting/key1", "new val", uint64(0))

  if val, _ := globeDict.Get("key1"); val != "new val" {
    t.Errorf("Expected key1 to be new val but was %s", val)
  }
}

func TestDictPut(t *testing.T) {
  globeDict := initNewTestDict("TestDictPut")

  globeDict.Put("key1", "new val")

  if val, _ := globeDict.Get("key1"); val != "new val" {
    t.Errorf("Expected key1 to be new val but was %s", val)
  }
}

func TestDictWatchMulti(t *testing.T) {
  globeDict := initNewTestDict("TestDictWatchMulti")

  rawClient.Set("TestDictWatchMulti/key1", "new val", uint64(0))

  if val, _ := globeDict.Get("key1"); val != "new val" {
    t.Errorf("Expected key1 to be new val but was %s", val)
  }

  rawClient.Set("TestDictWatchMulti/key1", "new val2", uint64(0))

  if val, _ := globeDict.Get("key1"); val != "new val2" {
    t.Errorf("Expected key1 to be new val2 but was %s", val)
  }
}