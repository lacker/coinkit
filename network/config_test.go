package network

import (
	"bytes"
	"testing"
)

func TestSerializingConfig(t *testing.T) {
	c := &Config{
		Nodes:     make(map[string]*Address),
		Threshold: 3,
	}
	c.Nodes["a"] = &Address{Host: "a", Port: 1}
	c.Nodes["b"] = &Address{Host: "b", Port: 2}
	c.Nodes["c"] = &Address{Host: "c", Port: 3}
	c.Nodes["d"] = &Address{Host: "d", Port: 4}

	s := c.Serialize()
	c2 := NewConfigFromSerialized(s)
	s2 := c2.Serialize()
	if bytes.Compare(s, s2) != 0 {
		t.Fatal("serialize-deserialize fail in config")
	}
}
