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

import java.util.List;

public class Address {
  private static final List<org.hyperledger.besu.datatypes.Address> precompileAddress =
      List.of(
          org.hyperledger.besu.datatypes.Address.ECREC,
          org.hyperledger.besu.datatypes.Address.SHA256,
          org.hyperledger.besu.datatypes.Address.RIPEMD160,
          org.hyperledger.besu.datatypes.Address.ID,
          org.hyperledger.besu.datatypes.Address.MODEXP,
          org.hyperledger.besu.datatypes.Address.ALTBN128_ADD,
          org.hyperledger.besu.datatypes.Address.ALTBN128_MUL,
          org.hyperledger.besu.datatypes.Address.ALTBN128_PAIRING,
          org.hyperledger.besu.datatypes.Address.BLAKE2B_F_COMPRESSION);

  public static boolean isPrecompile(org.hyperledger.besu.datatypes.Address to) {
    return precompileAddress.contains(to);
  }
}
