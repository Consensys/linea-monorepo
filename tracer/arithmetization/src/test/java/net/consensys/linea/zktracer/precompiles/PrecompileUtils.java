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
import static net.consensys.linea.zktracer.Trace.GAS_CONST_BLS_G1_ADD;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_BLS_G2_ADD;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_BLS_MAP_FP2_TO_G2;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_BLS_MAP_FP_TO_G1;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_BLS_PAIRING_CHECK;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_BLS_PAIRING_CHECK_PAIR;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_ECADD;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_ECMUL;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_ECPAIRING;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_ECPAIRING_PAIR;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_ECRECOVER;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_IDENTITY;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_IDENTITY_WORD;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_MODEXP;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_POINT_EVALUATION;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_RIPEMD;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_RIPEMD_WORD;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_SHA2;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_SHA2_WORD;
import static net.consensys.linea.zktracer.Trace.OOB_INST_BLS_G1_MSM;
import static net.consensys.linea.zktracer.Trace.OOB_INST_BLS_G2_MSM;
import static net.consensys.linea.zktracer.Trace.Oob.G_QUADDIVISOR;
import static net.consensys.linea.zktracer.Trace.PRC_BLS_G1_MSM_MULTIPLICATION_COST;
import static net.consensys.linea.zktracer.Trace.PRC_BLS_G2_MSM_MULTIPLICATION_COST;
import static net.consensys.linea.zktracer.Trace.PRC_BLS_MULTIPLICATION_MULTIPLIER;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_G1_MSM;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_G2_MSM;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_PAIRING_CHECK;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_CALL_DATA_UNIT_SIZE___ECPAIRING;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_RETURN_DATA_SIZE___BLS_G1_ADD;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_RETURN_DATA_SIZE___BLS_G1_MSM;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_RETURN_DATA_SIZE___BLS_G2_ADD;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_RETURN_DATA_SIZE___BLS_G2_MSM;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_RETURN_DATA_SIZE___BLS_MAP_FP2_TO_G2;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_RETURN_DATA_SIZE___BLS_MAP_FP_TO_G1;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_RETURN_DATA_SIZE___BLS_PAIRING_CHECK;
import static net.consensys.linea.zktracer.Trace.PRECOMPILE_RETURN_DATA_SIZE___POINT_EVALUATION;
import static net.consensys.linea.zktracer.Trace.WORD_SIZE;
import static net.consensys.linea.zktracer.Trace.WORD_SIZE_MO;
import static net.consensys.linea.zktracer.instructionprocessing.callTests.Utilities.populateMemory;
import static org.hyperledger.besu.datatypes.Address.*;

import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.module.tables.bls.BlsRt;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;

public class PrecompileUtils extends TracerTestBase {

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
    } else if (precompileAddress.equals(KZG_POINT_EVAL)) {
      return getPointEvaluationCost();
    } else if (precompileAddress.equals(BLS12_G1ADD)) {
      return getBlsG1AddCost();
    } else if (precompileAddress.equals(BLS12_G1MULTIEXP)) {
      return getBlsG1MsmCost(cds);
    } else if (precompileAddress.equals(BLS12_G2ADD)) {
      return getBlsG2AddCost();
    } else if (precompileAddress.equals(BLS12_G2MULTIEXP)) {
      return getBlsG2MsmCost(cds);
    } else if (precompileAddress.equals(BLS12_PAIRING)) {
      return getBlsPairingCheckCost(cds);
    } else if (precompileAddress.equals(BLS12_MAP_FP_TO_G1)) {
      return getBlsMapFpToG1Cost();
    } else if (precompileAddress.equals(BLS12_MAP_FP2_TO_G2)) {
      return getBlsMapFp2ToG2Cost();
    } else {
      throw new IllegalArgumentException("Unknown precompile address");
    }
  }

  public static int getPrecompileCost(Address precompileAddress, int cds) {
    return getPrecompileCost(precompileAddress, cds, 0, 0, 0, 0);
  }

  private static int words(int sizeInBytes) {
    return (sizeInBytes + WORD_SIZE_MO) / WORD_SIZE;
  }

  public static int getECRECCost() {
    return GAS_CONST_ECRECOVER;
  }

  public static int getSHA256Cost(int cds) {
    return GAS_CONST_SHA2 + words(cds) * GAS_CONST_SHA2_WORD;
  }

  public static int getRIPEMD160Cost(int cds) {
    return GAS_CONST_RIPEMD + words(cds) * GAS_CONST_RIPEMD_WORD;
  }

  public static int getIDCost(int cds) {
    return GAS_CONST_IDENTITY + words(cds) * GAS_CONST_IDENTITY_WORD;
  }

  public static int getMODEXPCost(int bbs, int mbs, int exponentLog) {
    final int fOfMax = ((Math.max(bbs, mbs) + 7) / 8) * ((Math.max(bbs, mbs) + 7) / 8);
    final int bigNumerator = fOfMax * Math.max(exponentLog, 1);
    final int bigQuotient = bigNumerator / G_QUADDIVISOR;
    return Math.max(GAS_CONST_MODEXP, bigQuotient);
  }

  public static int getECADDCost() {
    return GAS_CONST_ECADD;
  }

  public static int getECMULCost() {
    return GAS_CONST_ECMUL;
  }

  public static int getECPAIRINGCost(int cds) {
    return GAS_CONST_ECPAIRING
        + GAS_CONST_ECPAIRING_PAIR * (cds / PRECOMPILE_CALL_DATA_UNIT_SIZE___ECPAIRING);
  }

  public static int getBLAKE2FCost(int r) {
    return GAS_CONST_BLAKE2_PER_ROUND * r;
  }

  public static int getPointEvaluationCost() {
    return GAS_CONST_POINT_EVALUATION;
  }

  public static int getBlsG1AddCost() {
    return GAS_CONST_BLS_G1_ADD;
  }

  public static int getBlsG1MsmCost(int cds) {
    final int numInputs = cds / PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_G1_MSM;
    final int discount = BlsRt.getMsmDiscount(OOB_INST_BLS_G1_MSM, numInputs);
    return (numInputs * PRC_BLS_G1_MSM_MULTIPLICATION_COST * discount)
        / PRC_BLS_MULTIPLICATION_MULTIPLIER;
  }

  public static int getBlsG2AddCost() {
    return GAS_CONST_BLS_G2_ADD;
  }

  public static int getBlsG2MsmCost(int cds) {
    final int numInputs = cds / PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_G2_MSM;
    final int discount = BlsRt.getMsmDiscount(OOB_INST_BLS_G2_MSM, numInputs);
    return (numInputs * PRC_BLS_G2_MSM_MULTIPLICATION_COST * discount)
        / PRC_BLS_MULTIPLICATION_MULTIPLIER;
  }

  public static int getBlsPairingCheckCost(int cds) {
    return GAS_CONST_BLS_PAIRING_CHECK
        + GAS_CONST_BLS_PAIRING_CHECK_PAIR
            * (cds / PRECOMPILE_CALL_DATA_UNIT_SIZE___BLS_PAIRING_CHECK);
  }

  public static int getBlsMapFpToG1Cost() {
    return GAS_CONST_BLS_MAP_FP_TO_G1;
  }

  public static int getBlsMapFp2ToG2Cost() {
    return GAS_CONST_BLS_MAP_FP2_TO_G2;
  }

  // Methods to prepare inputs for certain precompiles

  // BLAKE2F
  static void prepareBlake2F(BytecodeCompiler program, int rLeadingByte, int offset) {
    program
        .push(rLeadingByte) // For simplicity, we only set the first byte of r
        .push(offset) // offset
        .op(OpCode.MSTORE8);
  }

  public static Bytes prepareBlake2F(int rLeadingByte, int offset) {
    BytecodeCompiler program = BytecodeCompiler.newProgram(chainConfig);
    prepareBlake2F(program, rLeadingByte, offset);
    return program.compile();
  }

  // SHA256, RIPEMD160, ID
  static void prepareSha256Ripemd160Id(BytecodeCompiler program, int nWords, int offset) {
    populateMemory(program, nWords, offset);
  }

  static Bytes prepareSha256Ripemd160Id(int nWords, int offset) {
    BytecodeCompiler program = BytecodeCompiler.newProgram(chainConfig);
    prepareSha256Ripemd160Id(program, nWords, offset);
    return program.compile();
  }

  // MODEXP
  public static Bytes generateModexpInput(int bbs, int mbs, int ebs) {
    final Bytes32 bbsPadded = Bytes32.leftPad(Bytes.of(bbs));
    final Bytes32 ebsPadded = Bytes32.leftPad(Bytes.of(ebs));
    final Bytes32 mbsPadded = Bytes32.leftPad(Bytes.of(mbs));
    final Bytes bem =
        Bytes.fromHexString("0x" + "aa".repeat(bbs) + "ff".repeat(ebs) + "bb".repeat(mbs));
    return Bytes.concatenate(bbsPadded, ebsPadded, mbsPadded, bem);
  }

  static void prepareModexp(
      BytecodeCompiler program, Bytes modexpInput, int targetOffset, Address codeOwnerAddress) {
    // Copy to targetOffset the code of codeOwnerAccount
    program
        .push(codeOwnerAddress)
        .op(OpCode.EXTCODESIZE) // size
        .push(0) // offset
        .push(targetOffset) // targetOffset
        .push(codeOwnerAddress) // address
        .op(OpCode.EXTCODECOPY);
  }

  public static Bytes prepareModexp(Bytes modexpInput, int targetOffset, Address codeOwnerAddress) {
    BytecodeCompiler program = BytecodeCompiler.newProgram(chainConfig);
    prepareModexp(program, modexpInput, targetOffset, codeOwnerAddress);
    return program.compile();
  }

  /**
   * Computes the expected returnAtCapacity based on the precompile address, and callDataSize in the
   * case of ID.
   *
   * @param precompileAddress the address of the precompile contract.
   * @param callDataSize the call data size. Beyond the case of ID, this value is ignored.
   * @param mbs the modulo byte size. Beyond the case of MODEXP, this value is ignored.
   * @return the computed return at capacity.
   */
  static int getExpectedReturnAtCapacity(Address precompileAddress, int callDataSize, int mbs) {
    if (precompileAddress.equals(ECREC)
        || precompileAddress.equals(SHA256)
        || precompileAddress.equals(RIPEMD160)
        || precompileAddress.equals(ALTBN128_PAIRING)) {
      return WORD_SIZE;
    } else if (precompileAddress.equals(ALTBN128_ADD)
        || precompileAddress.equals(ALTBN128_MUL)
        || precompileAddress.equals(BLAKE2B_F_COMPRESSION)) {
      return 2 * WORD_SIZE;
    } else if (precompileAddress.equals(MODEXP)) {
      return mbs;
    } else if (precompileAddress.equals(ID)) {
      return callDataSize;
    } else if (precompileAddress.equals(KZG_POINT_EVAL)) {
      return PRECOMPILE_RETURN_DATA_SIZE___POINT_EVALUATION;
    } else if (precompileAddress.equals(BLS12_G1ADD)) {
      return PRECOMPILE_RETURN_DATA_SIZE___BLS_G1_ADD;
    } else if (precompileAddress.equals(BLS12_G1MULTIEXP)) {
      return PRECOMPILE_RETURN_DATA_SIZE___BLS_G1_MSM;
    } else if (precompileAddress.equals(BLS12_G2ADD)) {
      return PRECOMPILE_RETURN_DATA_SIZE___BLS_G2_ADD;
    } else if (precompileAddress.equals(BLS12_G2MULTIEXP)) {
      return PRECOMPILE_RETURN_DATA_SIZE___BLS_G2_MSM;
    } else if (precompileAddress.equals(BLS12_PAIRING)) {
      return PRECOMPILE_RETURN_DATA_SIZE___BLS_PAIRING_CHECK;
    } else if (precompileAddress.equals(BLS12_MAP_FP_TO_G1)) {
      return PRECOMPILE_RETURN_DATA_SIZE___BLS_MAP_FP_TO_G1;
    } else if (precompileAddress.equals(BLS12_MAP_FP2_TO_G2)) {
      return PRECOMPILE_RETURN_DATA_SIZE___BLS_MAP_FP2_TO_G2;
    } else {
      throw new IllegalArgumentException("Unknown precompile address");
    }
  }
}
