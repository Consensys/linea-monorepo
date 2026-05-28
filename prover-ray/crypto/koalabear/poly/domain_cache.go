// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package poly

import (
	"sync"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
)

// DomainCache memoizes FFT domains by cardinality. Safe for concurrent use.
type DomainCache struct {
	mu      sync.Mutex
	domains map[uint64]*fft.Domain
}

// Get returns the FFT domain of cardinality n, creating it on first use. Calling
// Get on a nil cache is valid and behaves like fft.NewDomain(n).
func (c *DomainCache) Get(n uint64) *fft.Domain {
	if c == nil {
		return fft.NewDomain(n)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.domains == nil {
		c.domains = make(map[uint64]*fft.Domain)
	}
	if d, ok := c.domains[n]; ok {
		return d
	}
	d := fft.NewDomain(n)
	c.domains[n] = d
	return d
}
