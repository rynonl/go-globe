package nameme

// Generic client interface used as a wrapper for the underyling log client.
// See etcd_client.go for example implementatino
type LogClient interface {
  Get(path string) (*LogNode, error)
  Put(path string, value string) error
  PutDir(path string) error
  Watch(path string, watcher chan *LogNode, closer chan bool)
}

type LogNode struct {
  Key       string
  Value     string
  Children  []LogNode
}