package httphandlers

import (
	"fmt"
	"net/http"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
)

type UnitOfWork struct {
	uowService UnitOfWorkService
}

func NewUnitOfWork(uowService UnitOfWorkService) *UnitOfWork {
	return &UnitOfWork{
		uowService: uowService,
	}
}

func (h *UnitOfWork) Submit(writer http.ResponseWriter, request *http.Request) {
	var uowContract manifestcontracts.UnitOfWork
	cb, err := ParseBodyAndHandleError(writer, request, &uowContract, true)
	if err != nil {
		cb()
		return
	}

	ctx := request.Context()
	if err = h.uowService.Submit(ctx, uowContract); err != nil {
		WriteResult(writer, http.StatusInternalServerError, ContentTypeJSON, GenericResponse{
			Message: "submission failed",
			Error:   fmt.Errorf("failed to submit unit-of-work: %w", err),
		})
		return
	}

	WriteResult(writer, http.StatusCreated, ContentTypeJSON, GenericResponse{
		Message: "ok",
	})
}
