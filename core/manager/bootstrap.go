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

import (
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	v3core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	PodNamespace   = "POD_NAMESPACE"
	PodName        = "POD_NAME"
	InstanceIP     = "INSTANCE_IP"
	IstiodAddr     = "istiod.istio-system.svc:15010"
	KitexXdsDomain = "KITEX_XDS_DOMAIN"
	// use json to marshal it.
	KitexXdsMetas     = "KITEX_XDS_METAS"
	IstiodSvrName     = "istiod.istio-system.svc"
	IstioVersion      = "ISTIO_VERSION"
	metaSeparator     = ","
	keyValueSeparator = "="
)

type BootstrapConfig struct {
	node      *v3core.Node
	xdsSvrCfg *XDSServerConfig
}

type XDSServerConfig struct {
	SvrName         string        // The name of the xDS server
	SvrAddr         string        // The address of the xDS server
	XDSAuth         bool          // If this xDS enable the authentication of xDS stream
	NDSNotRequired  bool          // required by default for Istio
	FetchXDSTimeout time.Duration // timeout for fecth xds, default to 1s
}

// GetFetchXDSTimeout get timeout.
func (xsc XDSServerConfig) GetFetchXDSTimeout() time.Duration {
	if xsc.FetchXDSTimeout == 0 {
		return defaultXDSFetchTimeout
	}
	return xsc.FetchXDSTimeout
}

func parseMetaEnvs(envs, istioVersion string) *structpb.Struct {
	defaultMeta := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			IstioVersion: {
				Kind: &structpb.Value_StringValue{StringValue: istioVersion},
			},
		},
	}
	if len(envs) == 0 {
		return defaultMeta
	}

	pbmeta := &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}
	err := protojson.Unmarshal([]byte(envs), pbmeta)
	if err != nil {
		klog.Warnf("[Kitex] XDS meta info is invalid %s, error %v", envs, err)
		return defaultMeta
	}
	return pbmeta
}

func nodeId(podIP, podName, namespace, nodeDomain string) string {
	//"sidecar~" + podIP + "~" + podName + "." + namespace + "~" + namespace + ".svc." + domain,
	return fmt.Sprintf("sidecar~%s~%s.%s~%s.svc.%s", podIP, podName, namespace, namespace, nodeDomain)
}

// newBootstrapConfig constructs the bootstrapConfig
func newBootstrapConfig(config *XDSServerConfig) (*BootstrapConfig, error) {
	// Get info from env
	// Info needed for construct the nodeID
	namespace := os.Getenv(PodNamespace)
	if namespace == "" {
		return nil, fmt.Errorf("[XDS] Bootstrap, POD_NAMESPACE is not set in env")
	}
	podName := os.Getenv(PodName)
	if podName == "" {
		return nil, fmt.Errorf("[XDS] Bootstrap, POD_NAME is not set in env")
	}
	podIP := os.Getenv(InstanceIP)
	if podIP == "" {
		return nil, fmt.Errorf("[XDS] Bootstrap, INSTANCE_IP is not set in env")
	}
	if config.FetchXDSTimeout == 0 {
		config.FetchXDSTimeout = defaultXDSFetchTimeout
	}
	// specify the version of istio in case of the canary deployment of istiod
	istioVersion := os.Getenv(IstioVersion)
	nodeDomain := os.Getenv(KitexXdsDomain)
	if nodeDomain == "" {
		nodeDomain = "cluster.local"
	}

	metaEnvs := os.Getenv(KitexXdsMetas)

	return &BootstrapConfig{
		node: &v3core.Node{
			Id:       nodeId(podIP, podName, namespace, nodeDomain),
			Metadata: parseMetaEnvs(metaEnvs, istioVersion),
		},
		xdsSvrCfg: config,
	}, nil
}
