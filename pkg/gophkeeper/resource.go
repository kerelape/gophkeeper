package gophkeeper

type (
	// ResourceID is id of a resource.
	ResourceID int64

	// Resource is a resource information.
	Resource struct {
		ID   ResourceID
		Type ResourceType
		Meta string
	}
)
