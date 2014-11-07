package nameme

import "testing"
import "time"

// TODO: Mock things so we don't need active etcd cluster and yucky sleep

// Expects etcd to be running on the local host under port 4001
func TestEtcdClient(t *testing.T) {
  client := NewEtcdClient(testCluster)

  rawClient.Set("test", "value", uint64(1))

  val, err := client.Get("test")

  if err != nil {
    t.Error(err)
  }

  if val.Value != "value" {
    t.Errorf("Expected %s to be 'value'", val.Value)
  }
  client = nil
}

func TestEtcdWatch(t *testing.T) {
  client := NewEtcdClient(testCluster)
  watchChan := make(chan *LogNode)

  go client.Watch("watchtest", watchChan, nil)

  // Janky way to give our watch time to be set
  time.Sleep(1000 * time.Millisecond)

  rawClient.Set("watchtest", "newvalue", uint64(10))

  change := <- watchChan

  if change.Value != "newvalue" {
    t.Errorf("Expected %s to be 'newvalue'", change.Value)
  }
  client = nil
}
