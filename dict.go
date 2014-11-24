package globe

import (
  "errors"
  "fmt"
  "strings"
  "sync"
  "time"
)

// A Key-Value dictionary backed by an distributed log
type Dict struct {
  keyspace      string
  logClient     LogClient
  localStore    map[string]string
  closeChan     chan bool
  lock          *sync.Mutex
}

// Constructs and returns a new dictionary.
// The returned dict will be populated with the latest snapshot of the log
// and automatically update the local copy of the data as it changes remotely.
func NewDict(keyspace string, logClient LogClient) (*Dict, error) {

  watchChan := make(chan *LogNode)
  closeChan := make(chan bool)

  lock := &sync.Mutex{}

  newDict := &Dict{
    keyspace:     keyspace,
    logClient:    logClient,
    localStore:   map[string]string{},
    closeChan:    closeChan,
    lock:         lock,
  }

  // Checks to see if keyspace exists, creates it if not.
  // Fails silently if log unreachable
  newDict.validateDirExists()

  // Setup watch
  go newDict.logClient.Watch(keyspace, watchChan, closeChan)

  // Force initial update
  err := newDict.initLocal()

  // If the initial, blocking, init fails(due to log not being available)
  // do not panic, but continue trying to init in a non-blocking way. The map
  // will appear empty until a connection is established, but the program will
  // not crash.
  if err != nil {
    fmt.Println("globe | Error initializing dict. Retrying.")
    go func() {
      for {
        // Continues trying to validate/create keyspace. This fails silently
        // if log is not reachable
        newDict.validateDirExists()

        if ierr := newDict.initLocal(); ierr == nil {
          fmt.Println("globe | Successfully initialized dict after error")
          return
        } else {
          fmt.Println("globe | Error initializing dict. Retrying.")
        }
        time.Sleep(5 * time.Second)
      }
    }()
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

  return newDict, err
}

// Close the dictionary. It will no longer respond to log updates
func (d *Dict) Close() {
  close(d.closeChan)
}

// Get value for given key. Returns error if key is not found.
func (d *Dict) Get(key string) (string, error) {
  d.lock.Lock()
  val, exist := d.localStore[key]
  d.lock.Unlock()

  if !exist {
    return "", errors.New(fmt.Sprintf("globe | Key %s does not exist in globe Dict %s", key, d.keyspace))
  }

  return val, nil
}

// Put value at given key. Overwrites any existing value.
func (d *Dict) Put(key string, value string) error {
  // TODO: Have this optimistic consistency be optional
  d.setLocal(key, value)

  err := d.logClient.Put(d.keyspace + "/" + key, value)
  return err
}

// Checks if a directory node exists and creates it if not.
func (d *Dict) validateDirExists() error {
  existing, _ := d.logClient.Get(d.keyspace)

  if existing == nil {
    if err := d.logClient.PutDir(d.keyspace); err != nil {
      return err
    }
  }

  return nil
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
  d.lock.Lock()
  defer d.lock.Unlock()

  for key, val := range batch {
    d.localStore[localKey(key)] = val
  }
}

// Sets a single entry in the map
func (d *Dict) setLocal(key string, val string) {
  d.lock.Lock()
  defer d.lock.Unlock()

  d.localStore[localKey(key)] = val
}

// Strips the keyspace prefix from the key name
func localKey(fullKey string) string {
  splits := strings.Split(fullKey, "/")
  return splits[len(splits) - 1]
}