package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

/* -------------------------------------------------------------------------- */
/* --------------------------- Crazy HTTP Wrappers -------------------------- */
/* -------------------------------------------------------------------------- */

func wrapHandErr(hand func(r *http.Request) ([]byte, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, status, err := hand(r)
		if err != nil {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write(b)
	}
}

func wrapHand[Res any](hand func() (Res, int, error)) http.HandlerFunc {
	return wrapHandErr(func(r *http.Request) ([]byte, int, error) {
		res, status, err := hand()
		if err != nil {
			return nil, status, err
		}

		b, err := json.Marshal(res)
		if err != nil {
			return nil, status, fmt.Errorf("failed to marshal the response: %w", err)
		}
		return b, http.StatusOK, nil
	})
}

func wrapHandB[Req any, Res any](hand func(Req) (Res, int, error)) http.HandlerFunc {
	return wrapHandErr(func(r *http.Request) ([]byte, int, error) {
		// Get request body
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, http.StatusBadRequest, fmt.Errorf("failed to read the body: %w", err)
		}
		r.Body.Close()

		// Unmarshal the body
		var req Req
		if err = json.Unmarshal(b, &req); err != nil {
			return nil, http.StatusBadRequest, fmt.Errorf("failed to unmarshal the body: %w", err)
		}

		// Run the handler
		resp, status, err := hand(req)
		if err != nil {
			return nil, status, err
		}

		// Marshal the response
		b, err = json.Marshal(resp)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to marshal the response: %w", err)
		}

		return b, http.StatusOK, nil
	})
}
