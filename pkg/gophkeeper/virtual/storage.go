package virtual

import (
	"sync"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

type (
	resource struct {
		meta  string
		id    int
		owner string
		_type gophkeeper.ResourceType
	}
	piece struct {
		content []byte
	}
	blob struct {
		location string
	}
)

type storage struct {
	resources []resource
	blobs     []blob
	pieces    []piece

	mutex *sync.Mutex
}
