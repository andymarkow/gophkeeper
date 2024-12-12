package objrepo

import (
	"fmt"
)

// ObjectInfo represents an object info.
type ObjectInfo struct {
	name     string
	crc32c   string
	location string
	size     int64
}

// NewObjectInfo creates a new object info.
func NewObjectInfo(name, crc32c, location string, size int64) (*ObjectInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("object name is empty")
	}

	return &ObjectInfo{
		name:     name,
		crc32c:   crc32c,
		location: location,
		size:     size,
	}, nil
}

// Name returns the name of the object.
func (o *ObjectInfo) Name() string {
	return o.name
}

// CRC32C returns the CRC32C checksum of the object.
func (o *ObjectInfo) CRC32C() string {
	return o.crc32c
}

// Location returns the download url of the object.
func (o *ObjectInfo) Location() string {
	return o.location
}

// Size returns the size of the object.
func (o *ObjectInfo) Size() int64 {
	return o.size
}
