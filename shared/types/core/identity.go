package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type ID string

func NewID(prefix string) ID {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return ID(fmt.Sprintf("%s%d", prefix, time.Now().UTC().UnixNano()))
	}
	return ID(prefix + hex.EncodeToString(bytes))
}

type Metadata struct {
	ID   ID
	Name string
	Description string
	Version     string
	Labels      map[string]string
}