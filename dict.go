package nameme

import (
  "errors"
  "fmt"
  "strings"
)

// A Key-Value dictionary backed by an distributed log
type Dict struct {
  keyspace      string
  logClient     LogClient
  localStore    map[string]string
}

// Constructs and returns a new dictionary.
// The returned dict will be populated with the latest snapshot of the log
// and automatically update the local copy of the data as it changes remotely.
func NewDict(keyspace string, logClient LogClient) (*Dict, error) {
  empty := make(map[string]string)

  newDict := &Dict{
    keyspace:     keyspace,
    logClient:    logClient,
    localStore:   empty,
  }

  watchChan := make(chan *LogNode)
  closeChan := make(chan bool)

  // Setup watch
  go newDict.logClient.Watch(keyspace, watchChan, closeChan)

  // Validate log node is as expected and force initial update
  if err := newDict.initLocal(); err != nil {
    return nil, err
  }

  // Start responding to log updates
  go func() {
    for {
      select {
        case node := <- watchChan:
          newDict.setLocal(node.Key, node.Value)
        case <- closeChan:
          close(watchChan)
          return
      }
    }
  }()

  return newDict, nil
}

// Close the dictionary. It will no longer respond to log updates
func (d *Dict) Close() {
  // TODO
}

// Get value for given key. Returns error if key is not found.
func (d *Dict) Get(key string) (string, error) {
  val, exist := d.localStore[key]

  if !exist {
    return "", errors.New(fmt.Sprintf("Key %s does not exist in Nameme Dict %s", key, d.keyspace))
  }

  return val, nil
}

// Put value at given key. Overwrites any existing value.
func (d *Dict) Put(key string, value string) error {
  err := d.logClient.Put(d.keyspace + "/" + key, value)
  return err
}

// A forced bootstrap of the local map. Used on instantiation only.
func (d *Dict) initLocal() error {
  existing, err := d.logClient.Get(d.keyspace)

  if err != nil {
    return err
  }

  if existing == nil {
    d.logClient.PutDir(d.keyspace)
  } else if (existing.Children != nil) && (len(existing.Children) > 0) {
    initialDict := d.buildDict(existing.Children)

    d.setBatchLocal(initialDict)
  }

  return nil
}

// Converts a slice of LogNodes to a dict
func (d *Dict) buildDict(nodes []LogNode) map[string]string {
  dict := make(map[string]string)

  for _, node := range nodes {
    dict[node.Key] = node.Value
  }

  return dict
}

// Merges a dict with the local version, overwriting existing keys
func (d *Dict) setBatchLocal(batch map[string]string) {
  for key, val := range batch {
    d.setLocal(key, val)
  }
}

// Sets a single entry in the map
func (d *Dict) setLocal(key string, val string) {
  d.localStore[localKey(key)] = val
}

// Strips the keyspace prefix from the key name
func localKey(fullKey string) string {
  splits := strings.Split(fullKey, "/")
  return splits[len(splits) - 1]
}