package postgres

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"log"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
	"golang.org/x/crypto/bcrypt"
)

// Identity is a postgres identity.
type Identity struct {
	Connection       *pgx.Conn
	PasswordEncoding *base64.Encoding
	BlobsDir         string

	Username string
}

var _ gophkeeper.Identity = (*Identity)(nil)

// StorePiece implements Identity.
func (i *Identity) StorePiece(ctx context.Context, piece gophkeeper.Piece, password string) (gophkeeper.ResourceID, error) {
	if err := i.comparePassword(ctx, password); err != nil {
		return -1, errors.Join(err, gophkeeper.ErrBadCredential)
	}

	var transaction, transactionError = i.Connection.Begin(ctx)
	if transactionError != nil {
		return -1, transactionError
	}
	defer transaction.Rollback(context.Background())

	insertPieceResult := transaction.QueryRow(ctx, `INSERT INTO pieces(content) VALUES($1) RETURNING id`, piece.Content)
	var id int
	if err := insertPieceResult.Scan(&id); err != nil {
		return -1, err
	}
	insertResourceResult := transaction.QueryRow(
		ctx,
		`INSERT INTO resources(meta, resource, type, owner) VALUES($1, $2, $3, $4) RETURNING id`,
		piece.Meta, id, (int)(gophkeeper.ResourceTypePiece), i.Username,
	)
	var rid int64
	if err := insertResourceResult.Scan(&rid); err != nil {
		return -1, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return -1, err
	}

	return (gophkeeper.ResourceID)(rid), nil
}

// RestorePiece implements Identity.
func (i *Identity) RestorePiece(ctx context.Context, rid gophkeeper.ResourceID, password string) (gophkeeper.Piece, error) {
	if err := i.comparePassword(ctx, password); err != nil {
		return gophkeeper.Piece{}, errors.Join(err, gophkeeper.ErrBadCredential)
	}

	var (
		meta    string
		content []byte
	)
	var queryResourceResult = i.Connection.QueryRow(
		ctx,
		`SELECT meta, resource FROM resources WHERE id = $1 AND owner = $2 AND type = $3`,
		(int64)(rid), i.Username, (int)(gophkeeper.ResourceTypePiece),
	)
	var id int
	if err := queryResourceResult.Scan(&meta, &id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return gophkeeper.Piece{}, gophkeeper.ErrResourceNotFound
		}
		return gophkeeper.Piece{}, err
	}
	var queryPieceResult = i.Connection.QueryRow(
		ctx,
		`SELECT content FROM pieces WHERE id = $1`,
		id,
	)
	if err := queryPieceResult.Scan(&content); err != nil {
		return gophkeeper.Piece{}, err
	}

	var piece = gophkeeper.Piece{
		Meta:    meta,
		Content: content,
	}
	return piece, nil
}

// StoreBlob implements Identity.
func (i *Identity) StoreBlob(ctx context.Context, blob gophkeeper.Blob, password string) (gophkeeper.ResourceID, error) {
	defer blob.Content.Close()
	if err := i.comparePassword(ctx, password); err != nil {
		return -1, errors.Join(err, gophkeeper.ErrBadCredential)
	}

	var location = path.Join(i.BlobsDir, uuid.New().String())
	var file, createError = os.Create(location)
	if createError != nil {
		return -1, createError
	}

	if _, err := bufio.NewWriter(file).ReadFrom(blob.Content); err != nil {
		log.Printf("failed to write file: %s\n", err.Error())
		if err := file.Close(); err != nil {
			log.Printf("failed to close file: %s\n", err.Error())
		}
		if err := os.Remove(location); err != nil {
			log.Printf("failed to remove file: %s\n", err.Error())
		}
		return -1, err
	}
	if err := file.Close(); err != nil {
		log.Printf("failed to close file: %s\n", err.Error())
		return -1, err
	}

	var transaction, transactionError = i.Connection.Begin(ctx)
	if transactionError != nil {
		return -1, transactionError
	}
	defer transaction.Rollback(context.Background())

	var (
		blobID int
		rid    int64
	)

	var insertBlobResult = transaction.QueryRow(
		ctx,
		`INSERT INTO blobs(location) VALUES($1) RETURNING id`,
		location,
	)
	if err := insertBlobResult.Scan(&blobID); err != nil {
		return -1, err
	}

	var insertResourceResult = transaction.QueryRow(
		ctx,
		`INSERT INTO resources(meta, owner, type, resource) VALUES($1, $2, $3, $4) RETURNING id`,
		blob.Meta, i.Username, gophkeeper.ResourceTypeBlob, blobID,
	)
	if err := insertResourceResult.Scan(&rid); err != nil {
		return -1, err
	}

	if err := transaction.Commit(ctx); err != nil {
		return -1, err
	}

	return (gophkeeper.ResourceID)(rid), nil
}

// RestoreBlob implements Identity.
func (i *Identity) RestoreBlob(ctx context.Context, rid gophkeeper.ResourceID, password string) (gophkeeper.Blob, error) {
	if err := i.comparePassword(ctx, password); err != nil {
		return gophkeeper.Blob{}, errors.Join(err, gophkeeper.ErrBadCredential)
	}

	var selectResourceResult = i.Connection.QueryRow(
		ctx,
		`SELECT meta, resource FROM resources WHERE id = $1 AND owner = $2`,
		(int64)(rid), i.Username,
	)
	var (
		meta   string
		blobID int
	)
	if err := selectResourceResult.Scan(&meta, &blobID); err != nil {
		return gophkeeper.Blob{}, err
	}

	var location string
	var selectBlobResult = i.Connection.QueryRow(
		ctx,
		`SELECT location FROM blobs WHERE id = $1`,
		blobID,
	)
	if err := selectBlobResult.Scan(&location); err != nil {
		return gophkeeper.Blob{}, err
	}

	var file, fileError = os.Open(location)
	if fileError != nil {
		return gophkeeper.Blob{}, fileError
	}

	var blob = gophkeeper.Blob{
		Meta:    meta,
		Content: file,
	}
	return blob, nil
}

// Delete implements Identity.
func (i *Identity) Delete(ctx context.Context, rid gophkeeper.ResourceID) error {
	var transaction, transactionError = i.Connection.Begin(ctx)
	if transactionError != nil {
		return transactionError
	}
	defer transaction.Rollback(context.Background())

	var deleteResourceResult = transaction.QueryRow(
		ctx,
		`DELETE FROM resources WHERE id = $1 AND owner = $2 RETURNING type, resource`,
		(int64)(rid), i.Username,
	)
	var (
		resourceType int
		resourceID   int
	)
	if err := deleteResourceResult.Scan(&resourceType, &resourceID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return gophkeeper.ErrResourceNotFound
		}
		return err
	}

	switch (gophkeeper.ResourceType)(resourceType) {
	case gophkeeper.ResourceTypePiece:
		_, err := transaction.Exec(
			ctx,
			`DELETE FROM pieces WHERE id = $1`,
			resourceID,
		)
		if err != nil {
			return err
		}
	case gophkeeper.ResourceTypeBlob:
		var deleteResult = transaction.QueryRow(
			ctx,
			`DELETE FROM blobs WHERE id = $1 RETURNING location`,
			resourceID,
		)
		var location string
		if err := deleteResult.Scan(&location); err != nil {
			return err
		}
		if err := os.Remove(location); err != nil {
			return err
		}
	default:
		return errors.New("unknown resource type")
	}

	if err := transaction.Commit(ctx); err != nil {
		return err
	}
	return nil
}

// List implements Identity.
func (i *Identity) List(ctx context.Context) ([]gophkeeper.Resource, error) {
	var selectResourcesResult, selectResourcesResultError = i.Connection.Query(
		ctx,
		`SELECT id, type, meta FROM resources WHERE owner = $1`,
		i.Username,
	)
	if selectResourcesResultError != nil {
		return nil, selectResourcesResultError
	}
	defer selectResourcesResult.Close()
	var resources []gophkeeper.Resource
	for selectResourcesResult.Next() {
		if err := selectResourcesResult.Err(); err != nil {
			return nil, err
		}
		var resource gophkeeper.Resource
		if err := selectResourcesResult.Scan(&resource.ID, &resource.Type, &resource.Meta); err != nil {
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func (i *Identity) comparePassword(ctx context.Context, password string) error {
	var row = i.Connection.QueryRow(
		ctx,
		`SELECT password FROM identities WHERE username = $1`,
		i.Username,
	)
	var encodedPassword string
	if err := row.Scan(&encodedPassword); err != nil {
		var pgerr pgconn.PgError
		if errors.As(err, (any)(&pgerr)) {
			return gophkeeper.ErrBadCredential
		}
		return err
	}

	var decodedPassword, decodePasswordError = i.PasswordEncoding.DecodeString(encodedPassword)
	if decodePasswordError != nil {
		return decodePasswordError
	}
	if err := bcrypt.CompareHashAndPassword(decodedPassword, ([]byte)(password)); err != nil {
		return errors.Join(gophkeeper.ErrBadCredential, err)
	}
	return nil
}
