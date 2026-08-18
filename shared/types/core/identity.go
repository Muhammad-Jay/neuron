package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type ID string

// NewID generates a lexicographically sortable identifier: a fixed-width
// millisecond timestamp followed by random bytes. The timestamp prefix keeps
// key-ordered listings - such as event history - in occurrence order without
// requiring a separate sort field.
func NewID(prefix string) ID {
	timestamp := fmt.Sprintf("%012x", time.Now().UTC().UnixMilli())
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return ID(prefix + timestamp + fallbackSuffix())
	}
	return ID(prefix + timestamp + hex.EncodeToString(bytes))
}

func fallbackSuffix() string {
	return fmt.Sprintf("%024x", time.Now().UTC().UnixNano())
}

type Metadata struct {
	ID          ID
	Name        string
	Description string
	Version     string
	Labels      map[string]string
}
