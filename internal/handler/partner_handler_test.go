package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type partnerDeleteServiceStub struct {
	service.PartnerService
	deletePermanent func(uint) error
}

func (s partnerDeleteServiceStub) DeletePermanent(id uint) error {
	return s.deletePermanent(id)
}

func TestPartnerDeletePermanentResponse(t *testing.T) {
	for _, test := range []struct {
		name    string
		id      string
		err     error
		status  int
		message string
		called  bool
	}{
		{"success", "7", nil, http.StatusOK, "Partner berhasil dihapus permanen", true},
		{"has piutang", "7", entity.ErrPartnerHasPiutang, http.StatusConflict, entity.ErrPartnerHasPiutang.Error(), true},
		{"not archived", "7", entity.ErrPartnerNotArchived, http.StatusConflict, entity.ErrPartnerNotArchived.Error(), true},
		{"other reference", "7", entity.ErrPartnerReferenced, http.StatusConflict, entity.ErrPartnerReferenced.Error(), true},
		{"missing partner", "7", gorm.ErrRecordNotFound, http.StatusNotFound, "Partner tidak ditemukan", true},
		{"database failure", "7", errors.New("internal database error"), http.StatusInternalServerError, "Gagal menghapus partner secara permanen", true},
		{"zero id", "0", nil, http.StatusBadRequest, "ID tidak valid", false},
		{"negative id", "-1", nil, http.StatusBadRequest, "ID tidak valid", false},
		{"invalid id", "abc", nil, http.StatusBadRequest, "ID tidak valid", false},
		{"overflow id", "4294967296", nil, http.StatusBadRequest, "ID tidak valid", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			called := false
			h := NewPartnerHandler(partnerDeleteServiceStub{deletePermanent: func(id uint) error {
				called = true
				if id != 7 {
					t.Fatalf("wrong target: %d", id)
				}
				return test.err
			}})
			router := gin.New()
			router.POST("/master/partner/:id/delete-permanent", h.DeletePermanent)
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/master/partner/"+test.id+"/delete-permanent", nil)
			router.ServeHTTP(response, request)
			var body struct {
				Status  bool   `json:"status"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if response.Code != test.status || body.Status != (test.status == http.StatusOK) || body.Message != test.message || called != test.called {
				t.Fatalf("unexpected response: code=%d body=%+v serviceCalled=%v", response.Code, body, called)
			}
		})
	}
}
