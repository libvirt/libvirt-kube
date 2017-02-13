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

package hypervisor

import (
	"time"

	"github.com/golang/glog"
	"github.com/libvirt/libvirt-go"
)

// Retry every 5 seconds for 30 seconds, then every 15 seconds
// for another minute, then every 60 seconds thereafter
var reconnectDelay = []int{
	5, 5, 5, 5, 5, 5, 15, 15, 15, 15, 60,
}

type ConnectEventType string

var (
	ConnectReady  ConnectEventType = "ready"
	ConnectFailed ConnectEventType = "failed"
)

type ConnectEvent struct {
	Type ConnectEventType
	URI  string
	Conn *libvirt.Connect
}

func lazyUnregister(conn *libvirt.Connect) {
	conn.UnregisterCloseCallback()
	conn.Close()
}

func connect(uri string, notify chan ConnectEvent) error {
	glog.V(1).Infof("Trying to connect to %s", uri)
	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		return err
	}

	err = conn.RegisterCloseCallback(func(conn *libvirt.Connect, reason libvirt.ConnectCloseReason) {
		glog.V(1).Infof("Notify about connection close %d", reason)
		conn.Ref()
		// TODO: figure out why we can't call UnregisterCloseCallback directly
		// here - libvirt is supposed to allow this but it hangs...
		go lazyUnregister(conn)
		notify <- ConnectEvent{
			Type: ConnectFailed,
			URI:  uri,
			Conn: nil,
		}

		go connector(uri, notify)
	})
	if err != nil {
		conn.Close()
		return err
	}

	notify <- ConnectEvent{
		Type: ConnectReady,
		URI:  uri,
		Conn: conn,
	}

	return nil
}

func connector(uri string, notify chan ConnectEvent) {
	glog.V(1).Infof("Starting connect loop for %s", uri)
	var delayIndex = 0
	for {
		err := connect(uri, notify)
		if err == nil {
			return
		}

		glog.V(1).Infof("Unable to connect to %s, retry in %d seconds: %s",
			uri, reconnectDelay[delayIndex], err)
		time.Sleep(time.Duration(reconnectDelay[delayIndex]) * time.Second)
		if delayIndex < (len(reconnectDelay) - 1) {
			delayIndex++
		}
	}
}

func OpenConnect(uri string, notify chan ConnectEvent) {
	go connector(uri, notify)
}

func eventloop() {
	for {
		libvirt.EventRunDefaultImpl()
	}
}

func init() {
	libvirt.EventRegisterDefaultImpl()
	go eventloop()
}
