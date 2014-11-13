package nameme

import (
  "errors"
  "fmt"
  "github.com/coreos/go-etcd/etcd"
)

// Implements LogClient
type EtcdClient struct {
  underlying  *etcd.Client
}

// Constructor
func NewEtcdClient(cluster []string) *EtcdClient {

  actual := etcd.NewClient(cluster)

  return &EtcdClient{
    underlying: actual,
  }
}

// LogClient Implementation
func (e *EtcdClient) Get(path string) (*LogNode, error) {
  resp, err := e.underlying.Get(path, false, false)

  if err != nil { return nil, errors.New(fmt.Sprintf("nameme | etcd | %s", err.Error())) }

  node := etcdResponseToLogNode(resp)

  return node, nil
}

func (e *EtcdClient) Put(path string, value string) error {

  _, err := e.underlying.Set(path, value, uint64(0))

  return err
}

func (e *EtcdClient) PutDir(path string) error {

  _, err := e.underlying.SetDir(path, uint64(0))

  return err
}

func (e *EtcdClient) Watch(path string, watcher chan *LogNode, closer chan bool) {

  // Set watch with underlying client
  underlyingChan := make(chan *etcd.Response)
  go func() {
    e.underlying.Watch(path, 0, true, underlyingChan, closer)
  }()

  for {
    change := <-underlyingChan

    // underlying client returns a nil change when etcd cluster is unreachable
    if change != nil {
      // Convert Etcd response to nameme nodes
      ret := etcdResponseToLogNode(change)
      watcher <- ret
    } else {
      // TODO recovery mode.  re-attach watch when etcd back online.
    }
  }

}

// Etcd Response to LogNode
func etcdResponseToLogNode(resp *etcd.Response) *LogNode {
  // TODO recursive children
  childrenLen := 0
  if resp.Node.Nodes != nil {
    childrenLen = len(resp.Node.Nodes)
  }

  children := make([]LogNode, childrenLen)
  for i :=0; i < len(resp.Node.Nodes); i++ {
    children[i] = LogNode{
      Key       : resp.Node.Nodes[i].Key,
      Value     : resp.Node.Nodes[i].Value,
      Children  : nil,
    }
  }

  return &LogNode{
    Key       : resp.Node.Key,
    Value     : resp.Node.Value,
    Children  : children,
  }
}
