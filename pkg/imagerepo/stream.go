/*
 * This file is part of the libvirt-kube project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2017 Red Hat, Inc.
 *
 */

package imagerepo

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang/glog"
)

// XXX kind of ugly to include the http status code as an return param
// but mapping error -> status codes in a fine grained manner is fugly
// too
type VolumeIOResolver interface {
	UploadVolume(imagerepo, imagefile, token string) (io.WriteCloser, uint64, int, error)

	DownloadVolume(imagerepo, imagefile, token string) (io.ReadCloser, uint64, string, int, error)
}

type VolumeStreamer struct {
	mux        *http.ServeMux
	insecure   bool
	server     *http.Server
	ioresolver VolumeIOResolver
}

func NewVolumeStreamer(listenAddr string, insecure bool, tlsConfig *tls.Config, ioresolver VolumeIOResolver) *VolumeStreamer {

	s := &VolumeStreamer{
		mux:        http.NewServeMux(),
		insecure:   insecure,
		ioresolver: ioresolver,
	}

	s.server = &http.Server{
		Addr:      listenAddr,
		TLSConfig: tlsConfig,
		Handler:   s.mux,
	}

	s.mux.HandleFunc("/stream/", s.handle)

	return s
}

func (s *VolumeStreamer) handle(res http.ResponseWriter, req *http.Request) {
	bits := strings.Split(req.URL.Path, "/")
	if len(bits) != 5 {
		// bits[0] -> ""
		// bits[1] -> "stream"
		// bits[2] -> repo name
		// bits[3] -> image name
		// bits[4] -> token
		glog.V(1).Infof("Unknown image file %s, only %d bits %shh", req.URL.Path, len(bits), bits)
		http.Error(res, "Unknown image file", http.StatusNotFound)
		return
	}

	// XXX spawn goroutine. safety ?

	switch req.Method {
	case http.MethodGet:
		glog.V(1).Infof("Try download repo=%s file=%s token=%s", bits[2], bits[3], bits[4])
		volio, length, filename, code, err := s.ioresolver.DownloadVolume(bits[2], bits[3], bits[4])

		if err != nil {
			glog.V(1).Infof("Unable to download volume code=%d msg=%s", code, err)
			http.Error(res, "Unable to download volume", code)
			return
		}

		glog.V(1).Infof("Running download %p %s %d", volio, filename, length)
		res.Header().Set("Content-Type", "application/octet-stream")
		res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		res.Header().Set("Content-Length", fmt.Sprintf("%d", length))

		res.WriteHeader(http.StatusOK)

		copied, err := io.Copy(res, volio)
		if err != nil {
			glog.V(1).Infof("Aborted sending prematurely %s", err)
			return
		}
		if uint64(copied) != length {
			glog.V(1).Infof("Volume was too short %d, expected %d", copied, length)
			return
		}
		volio.Close()

	case http.MethodPut:
		glog.V(1).Infof("Try upload repo=%s file=%s token=%s", bits[2], bits[3], bits[4])
		volio, length, code, err := s.ioresolver.UploadVolume(bits[2], bits[3], bits[4])

		if err != nil {
			glog.V(1).Infof("Unable to upload volume code=%d msg=%s", code, err)
			http.Error(res, "Unable to upload volume", code)
			return
		}

		glog.V(1).Infof("Running upload %p %s %d", volio, length)
		if req.ContentLength == -1 {
			http.Error(res, fmt.Sprintf("Volume length required %d", length), http.StatusLengthRequired)
			return
		}

		if uint64(req.ContentLength) != length {
			glog.V(1).Infof("Incorrect volume length %d, expected %d", req.ContentLength, length)
			http.Error(res, fmt.Sprintf("Incorrect volume length, expected %d", length), http.StatusRequestEntityTooLarge)
			return
		}

		copied, err := io.Copy(volio, req.Body)
		if err != nil {
			glog.V(1).Infof("Aborted recving prematurely %s", err)
			http.Error(res, "Unable to save volume", http.StatusInternalServerError)
			return
		}
		if uint64(copied) != length {
			http.Error(res, "Unable to save volume", http.StatusInternalServerError)
			return
		}
		volio.Close()

		res.Header().Set("Content-Type", "text/plain")
		res.Header().Set("Content-Length", "3")
		res.WriteHeader(http.StatusOK)
		io.WriteString(res, "OK\n")

	default:
		glog.V(1).Infof("Rejecting request for unsupported method %s", req.Method)
		http.Error(res, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
}

func (s *VolumeStreamer) Run(done chan error) {
	err := s.Serve()
	done <- err
}

func (s *VolumeStreamer) Serve() error {
	if s.insecure {
		glog.V(1).Infof("Listening HTTP on %s", s.server.Addr)
		err := s.server.ListenAndServe()
		if err != nil {
			return err
		}
	} else {
		glog.V(1).Infof("Listening HTTPS on %s", s.server.Addr)
		err := s.server.ListenAndServeTLS("", "")
		if err != nil {
			return err
		}
	}
	return nil
}
