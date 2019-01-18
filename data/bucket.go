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

// Makes a copy of this bucket with all of the provider data removed except provider IDs.
func (b *Bucket) StripProviderData() {
	ps := []*Provider{}
	for _, p := range b.Providers {
		ps = append(ps, &Provider{
			ID: p.ID,
		})
	}
	copy := new(Bucket)
	*copy = *b
	copy.Providers = ps
	return copy
}
