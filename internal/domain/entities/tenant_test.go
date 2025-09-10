package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTenantApplyTenantUpdate_NoChanges(t *testing.T) {
	now := time.Now().UTC().Add(-time.Hour)
	tenant := &Tenant{
		ID:        uuid.New(),
		Name:      "Acme",
		PublicURL: "https://public.acme.io",
		AdminURL:  "https://admin.acme.io",
		CreatedAt: now,
		UpdatedAt: now,
	}

	changed := tenant.ApplyTenantUpdate("Acme", "https://public.acme.io", "https://admin.acme.io")

	if changed {
		t.Fatalf("expected changed=false when inputs are identical")
	}
	if !tenant.UpdatedAt.Equal(now) {
		t.Fatalf("expected UpdatedAt unchanged; got %v want %v", tenant.UpdatedAt, now)
	}
}

func TestTenantApplyTenantUpdate_UpdateSomeFields(t *testing.T) {
	now := time.Now().UTC().Add(-time.Hour)
	tenant := &Tenant{
		ID:        uuid.New(),
		Name:      "Acme",
		PublicURL: "https://public.acme.io",
		AdminURL:  "https://admin.acme.io",
		CreatedAt: now,
		UpdatedAt: now,
	}

	changed := tenant.ApplyTenantUpdate("Acme Corp", "", "https://admin-new.acme.io")

	if !changed {
		t.Fatalf("expected changed=true when any field changes")
	}
	if tenant.Name != "Acme Corp" {
		t.Fatalf("expected Name updated; got %s", tenant.Name)
	}
	if tenant.PublicURL != "https://public.acme.io" {
		t.Fatalf("expected PublicURL unchanged when empty input; got %s", tenant.PublicURL)
	}
	if tenant.AdminURL != "https://admin-new.acme.io" {
		t.Fatalf("expected AdminURL updated; got %s", tenant.AdminURL)
	}
	if !tenant.UpdatedAt.After(now) {
		t.Fatalf("expected UpdatedAt to be refreshed; got %v <= %v", tenant.UpdatedAt, now)
	}
}
