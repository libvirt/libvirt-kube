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

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/pflag"

	"libvirt.org/libvirt-kube/pkg/imagerepo"
)

var (
	connect = pflag.String("connect", "qemu:///system",
		"Libvirt connection URI")
	kubeconfig = pflag.String("kubeconfig", "", "Path to a kube config, if running outside cluster")

	reponame = pflag.String("reponame", "default", "Name of virtimagerepo resource to manage")
	repopath = pflag.String("repopath", "/srv/images", "Path to image repository mount point")

	streaminsecure = pflag.Bool("stream-insecure", false,
		"Run public streamer without TLS encryption")
	streamaddr = pflag.String("stream-addr", "0.0.0.0:80",
		"TCP address and port to stream on")
	streamtlscert = pflag.String("stream-tls-cert", "/etc/pki/virtkubeimagerepo/server-cert.pem",
		"Path to TLS public server cert PEM file")
	streamtlskey = pflag.String("stream-tls-key", "/etc/pki/virtkubeimagerepo/server-key.pem",
		"Path to TLS public server key PEM file")
	streamtlsca = pflag.String("stream-tls-ca", "/etc/pki/virtkubeimagerepo/server-ca.pem",
		"Path to TLS public server CA cert PEM file")
)

func loadTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	calist := x509.NewCertPool()

	ok := calist.AppendCertsFromPEM(ca)
	if !ok {
		return nil, fmt.Errorf("Error loading CA certs from %s", caFile)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    calist,
	}

	return config, nil
}

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	// Convince glog that we really have parsed CLI
	flag.CommandLine.Parse([]string{})

	var streamTLS *tls.Config
	if !*streaminsecure {
		var err error
		streamTLS, err = loadTLSConfig(*streamtlscert, *streamtlskey, *streamtlsca)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	svc, err := imagerepo.NewService(*connect, *streamaddr, *streaminsecure, streamTLS, *kubeconfig, *reponame, *repopath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = svc.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
