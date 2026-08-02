package routing

import (
	"testing"
)

func TestRepositoryConstruction(t *testing.T) {
	repo := NewRepository(nil)
	if repo == nil {
		t.Fatal("NewRepository returned nil")
	}
}

func TestRouterConstruction(t *testing.T) {
	router := NewRouter(
		NewRepository(nil),
		nil,
		nil,
		nil,
		nil,
	)
	if router == nil {
		t.Fatal("NewRouter returned nil")
	}
	if router.fallbackEmail != nil {
		t.Error("expected nil fallbackEmail")
	}
	if router.fallbackSMS != nil {
		t.Error("expected nil fallbackSMS")
	}
}

func TestErrNoProviderForChannel(t *testing.T) {
	if ErrNoProviderForChannel == nil {
		t.Error("ErrNoProviderForChannel should not be nil")
	}
}
