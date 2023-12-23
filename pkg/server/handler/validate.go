package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/acorn-io/brent/pkg/apierror"
	"github.com/acorn-io/brent/pkg/parse"
	"github.com/acorn-io/brent/pkg/types"
	"github.com/acorn-io/schemer"
	"github.com/acorn-io/schemer/validation"
)

const (
	csrfCookie = "CSRF"
	csrfHeader = "X-API-CSRF"
)

func validateAction(request *types.APIRequest) (*schemas.Action, error) {
	if request.Action == "" || request.Link != "" || request.Method != http.MethodPost {
		return nil, nil
	}

	if err := request.AccessControl.CanAction(request, request.Schema, request.Action); err != nil {
		return nil, err
	}

	actions := request.Schema.CollectionActions
	if request.Name != "" {
		actions = request.Schema.ResourceActions
	}

	action, ok := actions[request.Action]
	if !ok {
		return nil, apierror.NewAPIError(validation.InvalidAction, fmt.Sprintf("Invalid action: %s", request.Action))
	}

	return &action, nil
}

func checkCSRF(apiOp *types.APIRequest) error {
	if !parse.IsBrowser(apiOp.Request, false) {
		return nil
	}

	cookie, err := apiOp.Request.Cookie(csrfCookie)
	if errors.Is(err, http.ErrNoCookie) {
		bytes := make([]byte, 5)
		_, err := rand.Read(bytes)
		if err != nil {
			return apierror.WrapAPIError(err, validation.ServerError, "Failed in CSRF processing")
		}

		cookie = &http.Cookie{
			Name:   csrfCookie,
			Value:  hex.EncodeToString(bytes),
			Path:   "/",
			Secure: true,
		}

		http.SetCookie(apiOp.Response, cookie)
	} else if err != nil {
		return apierror.NewAPIError(validation.InvalidCSRFToken, "Failed to parse cookies")
	} else if apiOp.Method != http.MethodGet {
		/*
		 * Very important to use apiOp.Method and not apiOp.Request.Method. The client can override the HTTP method with _method
		 */
		if cookie.Value == apiOp.Request.Header.Get(csrfHeader) {
			// Good
		} else if cookie.Value == apiOp.Request.URL.Query().Get(csrfCookie) {
			// Good
		} else {
			return apierror.NewAPIError(validation.InvalidCSRFToken, "Invalid CSRF token")
		}
	}

	return nil
}
