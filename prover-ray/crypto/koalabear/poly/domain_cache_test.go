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

import "testing"

func TestDomainCacheReusesDomainsBySize(t *testing.T) {
	var cache DomainCache
	d8 := cache.Get(8)
	if got := cache.Get(8); got != d8 {
		t.Fatal("DomainCache did not reuse domain for same size")
	}
	if got := cache.Get(16); got == d8 {
		t.Fatal("DomainCache reused domain for different size")
	}
}

func TestNilDomainCacheGet(t *testing.T) {
	var cache *DomainCache
	if got := cache.Get(8); got == nil {
		t.Fatal("nil DomainCache returned nil domain")
	}
}
