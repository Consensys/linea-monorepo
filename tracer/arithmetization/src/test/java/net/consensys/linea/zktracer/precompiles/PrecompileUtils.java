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

package net.consensys.linea.zktracer.precompiles;

import static net.consensys.linea.zktracer.Trace.GAS_CONST_BLAKE2_PER_ROUND;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_ECADD;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_ECMUL;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_ECPAIRING;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_ECPAIRING_PAIR;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_ECRECOVER;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_IDENTITY;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_IDENTITY_WORD;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_MODEXP;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_RIPEMD;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_RIPEMD_WORD;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_SHA2;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_SHA2_WORD;
import static net.consensys.linea.zktracer.Trace.Oob.G_QUADDIVISOR;
import static net.consensys.linea.zktracer.Trace.PRC_ECPAIRING_SIZE;
import static net.consensys.linea.zktracer.Trace.WORD_SIZE;
import static net.consensys.linea.zktracer.Trace.WORD_SIZE_MO;
import static org.hyperledger.besu.datatypes.Address.*;

import org.hyperledger.besu.datatypes.Address;

public class PrecompileUtils {

  /**
   * Computes the precompile cost based on the precompile address, arguments size, and r value in
   * case of BLAKE2F.
   *
   * @param precompileAddress the address of the precompile contract.
   * @param cds the call data size for SHA256, RIPEMD160, ID, and MODEXP. For other precompile
   *     contracts this value is ignored.
   * @param bbs the base byte size for MODEXP. For other precompile contracts, this value is
   *     ignored.
   * @param mbs the modulus byte size for MODEXP. For other precompile contracts, this value is
   *     ignored.
   * @param exponentLog the logarithm of the exponent for MODEXP. For other precompile contracts,
   *     this value is ignored.
   * @param r the number of rounds for BLAKE2F. For other precompile contracts, this value is
   *     ignored.
   * @return the computed precompile cost.
   */
  public static int getPrecompileCost(
      Address precompileAddress, int cds, int bbs, int mbs, int exponentLog, int r) {
    if (precompileAddress.equals(ECREC)) {
      return getECRECCost();
    } else if (precompileAddress.equals(SHA256)) {
      return getSHA256Cost(cds);
    } else if (precompileAddress.equals(RIPEMD160)) {
      return getRIPEMD160Cost(cds);
    } else if (precompileAddress.equals(ID)) {
      return getIDCost(cds);
    } else if (precompileAddress.equals(MODEXP)) {
      return getMODEXPCost(bbs, mbs, exponentLog);
    } else if (precompileAddress.equals(ALTBN128_ADD)) {
      return getECADDCost();
    } else if (precompileAddress.equals(ALTBN128_MUL)) {
      return getECMULCost();
    } else if (precompileAddress.equals(ALTBN128_PAIRING)) {
      return getECPAIRINGCost(cds);
    } else if (precompileAddress.equals(BLAKE2B_F_COMPRESSION)) {
      return getBLAKE2FCost(r);
    } else {
      throw new IllegalArgumentException("Unknown precompile address");
    }
  }

  private static int getECRECCost() {
    return GAS_CONST_ECRECOVER;
  }

  private static int getSHA256Cost(int cds) {
    final int words = (cds + WORD_SIZE_MO) / WORD_SIZE;
    return GAS_CONST_SHA2 + words * GAS_CONST_SHA2_WORD;
  }

  private static int getRIPEMD160Cost(int cds) {
    final int words = (cds + WORD_SIZE_MO) / WORD_SIZE;
    return GAS_CONST_RIPEMD + words * GAS_CONST_RIPEMD_WORD;
  }

  private static int getIDCost(int cds) {
    final int words = (cds + WORD_SIZE_MO) / WORD_SIZE;
    return GAS_CONST_IDENTITY + words * GAS_CONST_IDENTITY_WORD;
  }

  static int getMODEXPCost(int bbs, int mbs, int exponentLog) {
    final int fOfMax = ((Math.max(bbs, mbs) + 7) / 8) * ((Math.max(bbs, mbs) + 7) / 8);
    final int bigNumerator = fOfMax * Math.max(exponentLog, 1);
    final int bigQuotient = bigNumerator / G_QUADDIVISOR;
    return Math.max(GAS_CONST_MODEXP, bigQuotient);
  }

  private static int getECADDCost() {
    return GAS_CONST_ECADD;
  }

  private static int getECMULCost() {
    return GAS_CONST_ECMUL;
  }

  private static int getECPAIRINGCost(int cds) {
    return GAS_CONST_ECPAIRING + GAS_CONST_ECPAIRING_PAIR * (cds / PRC_ECPAIRING_SIZE);
  }

  private static int getBLAKE2FCost(int r) {
    return GAS_CONST_BLAKE2_PER_ROUND * r;
  }
}
