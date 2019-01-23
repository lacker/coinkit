package data

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
)

type Bucket struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`

	// Measured in megabytes
	Size uint32 `json:"size"`

	Providers ProviderArray `json:"providers"`
}

var validBucketName = regexp.MustCompile("^[-a-zA-Z0-9]*$")

// Bucket names must be valid domain names, and can't begin or end with hyphen
func IsValidBucketName(s string) bool {
	if !validBucketName.MatchString(s) {
		return false
	}
	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return false
	}
	return true
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
func (b *Bucket) StripProviderData() *Bucket {
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

func (b *Bucket) CheckEqual(other *Bucket) error {
	if b == nil && other == nil {
		return nil
	}
	if b == nil || other == nil {
		return fmt.Errorf("b != other. b is %+v, other is %+v", b, other)
	}
	if b.Name != other.Name {
		return fmt.Errorf("name %s != name %s", b.Name, other.Name)
	}
	if b.Owner != other.Owner {
		return fmt.Errorf("owner %s != owner %s", b.Owner, other.Owner)
	}
	if b.Size != other.Size {
		return fmt.Errorf("size %d != size %d", b.Size, other.Size)
	}
	return b.Providers.CheckEqual(other.Providers)
}

// Value and Scan let a BucketArray map to a sql text[] with just bucket names
type BucketArray []*Bucket

// TODO: escape things properly
func (bs BucketArray) Value() (driver.Value, error) {
	panic("TODO")
}

// TODO: unescape things properly
func (bs *BucketArray) Scan(src interface{}) error {
	panic("TODO")
}

func (bs BucketArray) CheckEqual(other BucketArray) error {
	if len(bs) != len(other) {
		return fmt.Errorf("len %d != len %d", len(bs), len(other))
	}
	for i, b := range bs {
		err := b.CheckEqual(other[i])
		if err != nil {
			return err
		}
	}
	return nil
}
