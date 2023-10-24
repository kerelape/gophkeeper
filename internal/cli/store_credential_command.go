package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/kerelape/gophkeeper/internal/stack"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

type storeCredentialCommand struct {
	gophkeeper gophkeeper.Gophkeeper
}

var _ command = (*storeCredentialCommand)(nil)

// Description implements command.
func (s *storeCredentialCommand) Description() string {
	return "Store username-password pair."
}

// Execute implements command.
func (s *storeCredentialCommand) Execute(ctx context.Context, args stack.Stack[string]) (bool, error) {
	if len(args) > 0 {
		return false, errors.New("expected 0 arguments")
	}
	gophkeeperIdentity, authenticateError := authenticate(ctx, s.gophkeeper)
	if authenticateError != nil {
		return true, authenticateError
	}
	description, descriptionError := description(ctx)
	if descriptionError != nil {
		return true, descriptionError
	}
	vaultPassword, vaultPasswordError := vaultPassword(ctx)
	if vaultPasswordError != nil {
		return true, vaultPasswordError
	}
	identity := identity{
		origin: gophkeeperIdentity,
	}
	username, password, credentialError := credential(ctx)
	if credentialError != nil {
		return true, credentialError
	}
	resource := credentialResource{
		description: description,
		username:    username,
		password:    password,
	}
	rid, storeError := identity.StoreCredential(ctx, resource, vaultPassword)
	if storeError != nil {
		return true, storeError
	}
	fmt.Printf("Successfully stored credential.\n")
	fmt.Printf("RID of the newly stored resource is %d.\n", rid)
	return true, nil
}

// Help implements command.
func (s *storeCredentialCommand) Help() string {
	return ""
}
