package nameme

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