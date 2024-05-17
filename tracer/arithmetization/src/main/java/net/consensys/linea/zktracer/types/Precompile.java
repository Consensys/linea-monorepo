/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.types;

import java.util.Optional;

import lombok.RequiredArgsConstructor;
import org.hyperledger.besu.datatypes.Address;

@RequiredArgsConstructor
public enum Precompile {
  EC_RECOVER(Address.ECREC),
  SHA2_256(Address.SHA256),
  RIPEMD_160(Address.RIPEMD160),
  IDENTITY(Address.ID),
  MODEXP(Address.MODEXP),
  EC_ADD(Address.ALTBN128_ADD),
  EC_MUL(Address.ALTBN128_MUL),
  EC_PAIRING(Address.ALTBN128_PAIRING),
  BLAKE2F(Address.BLAKE2B_F_COMPRESSION);

  public final Address address;

  public static Optional<Precompile> maybeOf(Address a) {
    if (a.equals(Address.ECREC)) {
      return Optional.of(Precompile.EC_RECOVER);
    } else if (a.equals(Address.SHA256)) {
      return Optional.of(Precompile.SHA2_256);
    } else if (a.equals(Address.RIPEMD160)) {
      return Optional.of(Precompile.RIPEMD_160);
    } else if (a.equals(Address.ID)) {
      return Optional.of(Precompile.IDENTITY);
    } else if (a.equals(Address.MODEXP)) {
      return Optional.of(Precompile.MODEXP);
    } else if (a.equals(Address.ALTBN128_ADD)) {
      return Optional.of(Precompile.EC_ADD);
    } else if (a.equals(Address.ALTBN128_MUL)) {
      return Optional.of(Precompile.EC_MUL);
    } else if (a.equals(Address.ALTBN128_PAIRING)) {
      return Optional.of(Precompile.EC_PAIRING);
    } else if (a.equals(Address.BLAKE2B_F_COMPRESSION)) {
      return Optional.of(Precompile.BLAKE2F);
    }

    return Optional.empty();
  }
}
