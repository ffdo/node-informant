package data

import "time"

var (
	possibleNodeIdKey = []string{"nodeinfo.node_id", "statistics.node_id", "neighbours.node_id"}
)

type NodeData struct {
	Root      interface{}
	timestamp time.Time
}

func (c NodeData) Set(path string, value interface{}) error {
	return set(c.Root, path, value)
}

func (c NodeData) Get(path string) (interface{}, error) {
	return get(c.Root, path)
}

func (c NodeData) Merge(other NodeData) error {
	var err error
	c.Root, err = merge(c.Root, other.Root)
	return err
}

func (c NodeData) Timestamp() time.Time {
	return c.timestamp
}

func (c NodeData) UpdateTimestamp() {
	c.timestamp = time.Now()
}

func (c NodeData) NodeId() string {
	for _, key := range possibleNodeIdKey {
		if nodeId, err := c.Get(key); err == nil {
			return nodeId.(string)
		}
	}
	return ""
}
