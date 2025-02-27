/*
 * Copyright 2022 CloudWeGo Authors
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
 */

package manager

import "fmt"

type Options struct {
	XDSSvrConfig *XDSServerConfig
	DumpPath     string
}

func (o *Options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

type Option struct {
	F func(o *Options)
}

func DefaultOptions() *Options {
	return &Options{
		XDSSvrConfig: &XDSServerConfig{
			SvrAddr: IstiodAddr,
			SvrName: IstiodSvrName,
			XDSAuth: false,
		},
		DumpPath: defaultDumpPath,
	}
}

func NewOptions(opts []Option) *Options {
	o := DefaultOptions()
	o.Apply(opts)
	return o
}

func CheckXDSSvrConfig(cfg *XDSServerConfig) error {
	if cfg.SvrAddr == "" {
		return fmt.Errorf("[XDS] Option: xDS server address should be specified")
	}
	if cfg.SvrName == "" {
		// set default server name
		cfg.SvrName = IstiodSvrName
	}
	return nil
}
