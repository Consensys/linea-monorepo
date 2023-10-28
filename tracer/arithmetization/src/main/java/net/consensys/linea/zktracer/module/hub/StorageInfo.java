/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package net.consensys.linea.zktracer.module.hub;

import java.util.HashMap;
import java.util.Map;

import net.consensys.linea.zktracer.EWord;
import org.hyperledger.besu.datatypes.Address;

public class StorageInfo {
  final Map<Address, Map<EWord, EWord>> originalStorageValues = new HashMap<>();

  public EWord getOriginalValueOrUpdate(Address address, EWord key, EWord value) {
    EWord r =
        this.originalStorageValues.getOrDefault(address, new HashMap<>()).putIfAbsent(key, value);
    if (r == null) {
      return value;
    }
    return r;
  }

  public EWord getOriginalValueOrUpdate(Address address, EWord key) {
    return this.getOriginalValueOrUpdate(address, key, EWord.ZERO);
  }
}
