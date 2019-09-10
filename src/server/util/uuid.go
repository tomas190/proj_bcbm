package util

import (
	"github.com/rs/xid"
)

type UUID struct{}

func (uid *UUID) GenUUID() string {
	id := xid.New()
	return id.String()
}
