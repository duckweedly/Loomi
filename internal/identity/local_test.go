package identity

import "testing"

func TestLocalIdentityIsStable(t *testing.T) {
	ident := LocalDevIdentity()
	if ident.UserID != "user_local_dev" {
		t.Fatalf("UserID = %q", ident.UserID)
	}
	if ident.DisplayName != "Local Developer" {
		t.Fatalf("DisplayName = %q", ident.DisplayName)
	}
	if ident.Source != "local_dev" {
		t.Fatalf("Source = %q", ident.Source)
	}
}

func TestLocalIdentityIgnoresExternalUserChoice(t *testing.T) {
	ident := ResolveLocalIdentity("user_other")
	if ident.UserID != "user_local_dev" {
		t.Fatalf("UserID = %q", ident.UserID)
	}
}
