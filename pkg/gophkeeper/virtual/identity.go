package virtual

import (
	"bufio"
	"context"
	"os"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

// Identity is a virtual identity.
type Identity struct {
	identity

	blobsDir string

	storage *storage
}

var _ gophkeeper.Identity = (*Identity)(nil)

// StorePiece implements gophkeeper.Identity.
func (i *Identity) StorePiece(_ context.Context, origin gophkeeper.Piece, password string) (gophkeeper.ResourceID, error) {
	i.storage.mutex.Lock()
	defer i.storage.mutex.Unlock()

	if password != i.password {
		return -1, gophkeeper.ErrBadCredential
	}

	i.storage.pieces = append(
		i.storage.pieces,
		piece{
			content: origin.Content,
		},
	)
	i.storage.resources = append(
		i.storage.resources,
		resource{
			id:    len(i.storage.pieces) - 1,
			_type: gophkeeper.ResourceTypePiece,
			meta:  origin.Meta,
			owner: i.username,
		},
	)
	return (gophkeeper.ResourceID)(len(i.storage.resources) - 1), nil
}

// RestorePiece implements gophkeeper.Identity.
func (i *Identity) RestorePiece(_ context.Context, rid gophkeeper.ResourceID, password string) (gophkeeper.Piece, error) {
	i.storage.mutex.Lock()
	defer i.storage.mutex.Unlock()

	if password != i.password {
		return gophkeeper.Piece{}, gophkeeper.ErrBadCredential
	}

	if !((int)(rid) < len(i.storage.resources)) {
		return gophkeeper.Piece{}, gophkeeper.ErrResourceNotFound
	}

	var resource = i.storage.resources[rid]
	if resource.owner != i.username || resource._type != gophkeeper.ResourceTypePiece {
		return gophkeeper.Piece{}, gophkeeper.ErrResourceNotFound
	}

	var piece = gophkeeper.Piece{
		Meta:    resource.meta,
		Content: i.storage.pieces[resource.id].content,
	}
	return piece, nil
}

// StoreBlob implements gophkeeper.Identity.
func (i *Identity) StoreBlob(_ context.Context, origin gophkeeper.Blob, password string) (gophkeeper.ResourceID, error) {
	i.storage.mutex.Lock()
	defer i.storage.mutex.Unlock()
	defer origin.Content.Close()

	if password != i.password {
		return -1, gophkeeper.ErrBadCredential
	}

	var dir, dirError = os.MkdirTemp(i.blobsDir, "blobs-*")
	if dirError != nil {
		return -1, dirError
	}
	var file, fileError = os.CreateTemp(dir, "blob-*")
	if fileError != nil {
		return -1, fileError
	}
	if _, err := bufio.NewWriter(file).ReadFrom(origin.Content); err != nil {
		return -1, err
	}
	if err := file.Close(); err != nil {
		return -1, err
	}

	i.storage.blobs = append(
		i.storage.blobs,
		blob{
			location: file.Name(),
		},
	)
	i.storage.resources = append(
		i.storage.resources,
		resource{
			meta:  origin.Meta,
			id:    len(i.storage.blobs) - 1,
			owner: i.username,
			_type: gophkeeper.ResourceTypeBlob,
		},
	)

	return (gophkeeper.ResourceID)(len(i.storage.resources) - 1), nil
}

// RestoreBlob implements gophkeeper.Identity.
func (i *Identity) RestoreBlob(_ context.Context, rid gophkeeper.ResourceID, password string) (gophkeeper.Blob, error) {
	i.storage.mutex.Lock()
	defer i.storage.mutex.Unlock()

	if password != i.password {
		return gophkeeper.Blob{}, gophkeeper.ErrBadCredential
	}

	if !((int)(rid) < len(i.storage.resources)) {
		return gophkeeper.Blob{}, gophkeeper.ErrResourceNotFound
	}

	var resource = i.storage.resources[rid]
	if resource.owner != i.username || resource._type != gophkeeper.ResourceTypeBlob {
		return gophkeeper.Blob{}, gophkeeper.ErrResourceNotFound
	}

	var file, fileError = os.Open(i.storage.blobs[resource.id].location)
	if fileError != nil {
		return gophkeeper.Blob{}, fileError
	}

	var blob = gophkeeper.Blob{
		Meta:    resource.meta,
		Content: file,
	}
	return blob, nil
}

// Delete implements gophkeeper.Identity.
func (i *Identity) Delete(_ context.Context, rid gophkeeper.ResourceID) error {
	i.storage.mutex.Lock()
	defer i.storage.mutex.Unlock()

	if !((int)(rid) < len(i.storage.resources)) {
		return gophkeeper.ErrResourceNotFound
	}

	if i.storage.resources[rid].owner != i.username {
		return gophkeeper.ErrResourceNotFound
	}

	i.storage.resources[rid].owner = ""

	return nil
}

// List implements gophkeeper.Identity.
func (i *Identity) List(_ context.Context) ([]gophkeeper.Resource, error) {
	i.storage.mutex.Lock()
	defer i.storage.mutex.Unlock()

	var resources = make([]gophkeeper.Resource, 0)
	for rid, resource := range i.storage.resources {
		if resource.owner != i.username {
			continue
		}
		resources = append(
			resources,
			gophkeeper.Resource{
				ID:   (gophkeeper.ResourceID)(rid),
				Type: resource._type,
				Meta: resource.meta,
			},
		)
	}

	return resources, nil
}
