package gophkeeper

// ResourceType is type of resource stored.
type ResourceType int

const (
	// ResourceTypePiece is a resource type
	// indicating a resource of type Piece.
	ResourceTypePiece ResourceType = iota + 1

	// ResourceTypeBlob is a reource type
	// indicating a resource of type Blob.
	ResourceTypeBlob
)

// String returns string representation of ResourceType.
func (t ResourceType) String() string {
	switch t {
	case ResourceTypePiece:
		return "Piece"
	case ResourceTypeBlob:
		return "Blob"
	default:
		panic("unsupported ResourceType")
	}
}
