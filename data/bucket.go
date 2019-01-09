package data

import (
	"fmt"
)

type Bucket struct {
	Name  string
	Owner string
	Size  uint32
}

func (b *Bucket) String() string {
	return fmt.Sprintf("bucket:%s, size:%d", b.Name, b.Size)
}
