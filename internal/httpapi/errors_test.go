package httpapi

import (
	"net/http"
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestStatusForProductError(t *testing.T) {
	if got := statusForError(productdata.NewError(productdata.CodeInvalidRequest, "bad")); got != http.StatusBadRequest {
		t.Fatalf("invalid status = %d", got)
	}
	if got := statusForError(productdata.NewError(productdata.CodeThreadNotFound, "missing")); got != http.StatusNotFound {
		t.Fatalf("not found status = %d", got)
	}
	if got := statusForError(productdata.NewError(productdata.CodeRunNotFound, "missing")); got != http.StatusNotFound {
		t.Fatalf("run not found status = %d", got)
	}
	if got := statusForError(productdata.NewError(productdata.CodeActiveRunExists, "conflict")); got != http.StatusConflict {
		t.Fatalf("active run conflict status = %d", got)
	}
}
