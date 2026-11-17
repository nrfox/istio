//go:build integ

// Copyright Istio Authors. All Rights Reserved.
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

package tracing

import (
	"strings"
	"testing"

	"istio.io/istio/pkg/test/framework/components/cluster"
	"istio.io/istio/pkg/test/framework/components/istio/ingress"
	"istio.io/istio/pkg/test/util/assert"
)

// fakeCluster implements only the cluster.Cluster methods used by ingressHTTPAddress.
type fakeCluster struct {
	cluster.Cluster
}

func (fakeCluster) Name() string {
	return "primary"
}

// fakeIngress implements only the ingress.Instance methods used by ingressHTTPAddress.
type fakeIngress struct {
	ingress.Instance
	addrs []string
}

// HTTPAddresses returns the configured addresses; ports are ignored by ingressHTTPAddress.
func (f fakeIngress) HTTPAddresses() ([]string, []int) {
	return f.addrs, nil
}

func (fakeIngress) Namespace() string {
	return "istio-system"
}

func (fakeIngress) ServiceName() string {
	return "istio-ingressgateway"
}

func (fakeIngress) Cluster() cluster.Cluster {
	return fakeCluster{}
}

func TestIngressHTTPAddress(t *testing.T) {
	cases := []struct {
		name    string
		addrs   []string
		want    string
		wantErr bool
	}{
		{
			// The ingress component returns nil slices when address lookup fails.
			name:    "nil addresses",
			wantErr: true,
		},
		{
			name:    "empty addresses",
			addrs:   []string{},
			wantErr: true,
		},
		{
			name:  "single address",
			addrs: []string{"10.0.0.1"},
			want:  "10.0.0.1",
		},
		{
			name:  "multiple addresses",
			addrs: []string{"10.0.0.1", "fd00::1"},
			want:  "10.0.0.1",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ingressHTTPAddress(fakeIngress{addrs: tt.addrs})
			assert.Equal(t, got, tt.want)
			if tt.wantErr {
				assert.Error(t, err)
				if !strings.Contains(err.Error(), "istio-system/istio-ingressgateway in cluster primary") {
					t.Fatalf("error %q does not identify the ingress gateway and cluster", err)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}
