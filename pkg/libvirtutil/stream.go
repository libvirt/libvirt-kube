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

package libvirtutil

import (
	"github.com/libvirt/libvirt-go"
)

type StreamIO struct {
	stream *libvirt.Stream
	err    bool
}

func NewStreamIO(s *libvirt.Stream) *StreamIO {
	return &StreamIO{
		stream: s,
	}
}

func (s *StreamIO) Read(p []byte) (int, error) {
	n, err := s.stream.Recv(p)
	if err != nil {
		s.err = true
	}
	return n, err
}

func (s *StreamIO) Write(p []byte) (int, error) {
	n, err := s.stream.Send(p)
	if err != nil {
		s.err = true
	}
	return n, err
}

func (s *StreamIO) Close() error {
	if s.err {
		return s.stream.Abort()
	} else {
		return s.stream.Finish()
	}
}
