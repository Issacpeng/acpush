// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/coreos/rkt/rkt/config"

	"github.com/appc/acpush/libacpush"
)

var (
	flagDebug           = flag.Bool("debug", false, "Enables debug messages")
	flagInsecure        = flag.Bool("insecure", false, "Permits unencrypted traffic")
	flagSystemConfigDir = flag.String("system-conf", "/usr/lib/rkt", "Directory for system configuration")
	flagLocalConfigDir  = flag.String("local-conf", "/etc/rkt", "Directory for local configuration")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "acpush [OPTIONS] IMAGE SIGNATURE URL\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) < 3 {
		usage()
		return
	}

	if flagInsecure == nil {
		fmt.Fprintln(os.Stderr, "Insecure flag unset?")
		os.Exit(1)
	}
	if flagDebug == nil {
		fmt.Fprintln(os.Stderr, "Debug flag unset?")
		os.Exit(1)
	}
	if flagSystemConfigDir == nil {
		fmt.Fprintln(os.Stderr, "System config dir unset?")
		os.Exit(1)
	}
	if flagLocalConfigDir == nil {
		fmt.Fprintln(os.Stderr, "Local config dir unset?")
		os.Exit(1)
	}

	conf, err := config.GetConfigFrom(*flagSystemConfigDir, *flagLocalConfigDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	err = libacpush.Uploader{
		Acipath:  args[0],
		Ascpath:  args[1],
		Uri:      args[2],
		Insecure: *flagInsecure,
		Debug:    *flagDebug,
		SetHTTPHeaders: func(r *http.Request) {
			if r.URL == nil {
				return
			}
			headerer, ok := conf.AuthPerHost[r.URL.Host]
			if !ok {
				if *flagDebug {
					fmt.Fprintf(os.Stderr, "No auth present in config for domain %s.\n", r.URL.Host)
				}
				return
			}
			header := headerer.Header()
			for k, v := range header {
				r.Header[k] = append(r.Header[k], v...)
			}
		},
	}.Upload()

	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		os.Exit(1)
	}
	if *flagDebug {
		fmt.Fprintln(os.Stderr, "Upload successful")
	}
}
