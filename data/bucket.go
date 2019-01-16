package data

import (
	"fmt"
)

type Bucket struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`

	// Measured in megabytes
	Size uint32 `json:"size"`

	Providers ProviderArray `json:"providers"`
}

func (b *Bucket) String() string {
	return fmt.Sprintf("bucket:%s, size:%d", b.Name, b.Size)
}

func (b *Bucket) RemoveProvider(id uint64) {
	providers := []*Provider{}
	for _, p := range b.Providers {
		if p.ID != id {
			providers = append(providers, p)
		}
	}
	b.Providers = providers
}
