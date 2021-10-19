package keel

import (
	"errors"
	"io"
	"net/http"

	"github.com/tkeel-io/tkeel/pkg/openapi"
)

func (k *Keel) Route(e *openapi.APIEvent) {
	// check path.
	path := e.HTTPReq.RequestURI
	next, err := checkRoutePath(path)
	if err != nil {
		log.Error(err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}
	if !next {
		e.Write([]byte("succ."))
		return
	}

	pID, err := auth(e)
	if err != nil {
		log.Errorf("error auth: %s", err)
		http.Error(e, err.Error(), http.StatusBadRequest)
		return
	}

	log.Debugf("route pID(%s) requst %s", pID, e.HTTPReq.RequestURI)

	// find upstream plugin.
	upPluginID, endpoint, err := getUpstreamPlugin(e.HTTPReq.Context(), pID, path)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.Errorf("not found(%s)", path)
			http.Error(e, err.Error(), http.StatusNotFound)
			return
		}
		log.Errorf("error request(%s): %s", path, err)
		http.Error(e, err.Error(), http.StatusBadRequest)
		return
	}

	// check upstream plugin.
	err = checkPluginStatus(e.HTTPReq.Context(), upPluginID)
	if err != nil {
		log.Errorf("error check plugin(%s) status: %s", upPluginID, err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}

	resp, err := proxy(e.HTTPReq, e.ResponseWriter, upPluginID, endpoint)
	if err != nil {
		log.Errorf("error proxy: %s", err)
		http.Error(e, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	copyHeader(e.Header(), resp.Header)
	respBodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("error get response body: %s", err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}
	if _, err = e.Write(respBodyByte); err != nil {
		log.Errorf("error response write: %s", err)
		return
	}
	log.Debugf("route succ.")
}
