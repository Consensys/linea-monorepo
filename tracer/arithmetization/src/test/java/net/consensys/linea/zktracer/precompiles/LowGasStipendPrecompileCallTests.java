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

package net.consensys.linea.zktracer.precompiles;

import static net.consensys.linea.zktracer.Trace.GAS_CONST_G_CALL_STIPEND;
import static net.consensys.linea.zktracer.Trace.PRC_BLAKE2F_SIZE;
import static net.consensys.linea.zktracer.Trace.PRC_ECPAIRING_SIZE;
import static net.consensys.linea.zktracer.Trace.WORD_SIZE;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.generateModexpInput;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.getExpectedReturnAtCapacity;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.getPrecompileCost;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.prepareBlake2F;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.prepareModexp;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.prepareSha256Ripemd160Id;
import static org.hyperledger.besu.datatypes.Address.ALTBN128_ADD;
import static org.hyperledger.besu.datatypes.Address.ALTBN128_MUL;
import static org.hyperledger.besu.datatypes.Address.ALTBN128_PAIRING;
import static org.hyperledger.besu.datatypes.Address.BLAKE2B_F_COMPRESSION;
import static org.hyperledger.besu.datatypes.Address.ECREC;
import static org.hyperledger.besu.datatypes.Address.ID;
import static org.hyperledger.besu.datatypes.Address.MODEXP;
import static org.hyperledger.besu.datatypes.Address.RIPEMD160;
import static org.hyperledger.besu.datatypes.Address.SHA256;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.stream.Stream;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.oob.OobOperation;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

public class LowGasStipendPrecompileCallTests {

  // Enums for the different testing scenarios
  enum ValueCase {
    ZERO,
    NON_ZERO;

    boolean isZeroCase() {
      return this == ZERO;
    }

    boolean isNonZeroCase() {
      return this == NON_ZERO;
    }
  }

  enum GasCase {
    ZERO,
    ONE,
    COST_MINUS_ONE,
    COST,
    COST_PLUS_ONE;
  }

  // BLAKE2F specific parameters
  int rLeadingByte = 0;
  int r = 0;

  // MODEXP specific parameters
  int bbs; // Defined within the test case
  int ebs; // Defined within the test case
  int mbs; // Defined within the test case
  int exponentLog; // Computed within the test case
  final Address codeOwnerAddress = Address.fromHexString("0xC0DE");
  List<ToyAccount> additionalAccounts = new ArrayList<>(); // Re-assigned within the test case

  /**
   * Parameterized test for low gas stipend precompile call. In this family of tests we call every
   * precompile with hand-crafted gas parameters in the CALL. The frame itself is provided with
   * ample gas, but we choose the gas parameter to be equal to the execution cost of the precompile
   * contract (up to -1, 0, +1). This allows us to deliberately trigger failures in the precompiles
   * due to insufficient gas and test that, on the contrary, calls to precompiles with *exactly* the
   * right amount of gas are successful (given that other conditions are met in terms of the
   * contents of the call data).
   *
   * @param precompileAddress the address of the precompile contract.
   * @param valueCase the value case (zero or non-zero).
   * @param gasCase the gas case (zero, one, cost minus one, cost, cost plus one).
   * @param modexpCostGT200OrBlake2fRoundsGT0 flag indicating if the MODEXP cost is greater than 200
   *     or if the BLAKE2F rounds are greater than 0. It is ignored for other precompile contracts.
   */
  @ParameterizedTest
  @MethodSource("lowGasStipendPrecompileCallTestSource")
  void lowGasStipendPrecompileCallTest(
      Address precompileAddress,
      ValueCase valueCase,
      GasCase gasCase,
      boolean modexpCostGT200OrBlake2fRoundsGT0) {
    final BytecodeCompiler program = BytecodeCompiler.newProgram();

    // In order to actually trigger the insufficient we need to:
    // - Set a specific callDataSize for BLAKE2F and EC_PAIRING
    // - Set the r value of BLAKE2F to have precompileCost > callStipend
    // - Populate the memory with a large enough number of words for SHA256, RIPEMD160, and ID
    //   to have precompileCost > callStipend.
    final int value = valueCase.isZeroCase() ? 0 : 1;
    final int callDataSize; // depends on the called precompile
    final int callDataOffset = 0;

    // returnAtCapacity is defined below
    final int returnAtOffset = 13;

    // Prepare the arguments for the different precompile calls
    if (precompileAddress == BLAKE2B_F_COMPRESSION) {
      rLeadingByte = modexpCostGT200OrBlake2fRoundsGT0 ? 0x12 : 0;
      r = rLeadingByte << 8;
      callDataSize = PRC_BLAKE2F_SIZE;
      prepareBlake2F(program, rLeadingByte, callDataOffset + 2);
    } else if (precompileAddress == ALTBN128_PAIRING) {
      callDataSize = PRC_ECPAIRING_SIZE;
    } else if ((precompileAddress == SHA256
        || precompileAddress == RIPEMD160
        || precompileAddress == ID)) {
      final int nWords = 1024;
      callDataSize = nWords * WORD_SIZE; // This guarantees that precompileCost > callStipend
      prepareSha256Ripemd160Id(program, nWords, callDataOffset);
    } else if (precompileAddress == MODEXP) {
      bbs = modexpCostGT200OrBlake2fRoundsGT0 ? 1 : 2;
      ebs = modexpCostGT200OrBlake2fRoundsGT0 ? 6 : 3;
      mbs = modexpCostGT200OrBlake2fRoundsGT0 ? 25 : 4;
      callDataSize = 96 + bbs + ebs + mbs;

      final Bytes modexpInput = generateModexpInput(bbs, mbs, ebs);
      exponentLog = OobOperation.computeExponentLog(modexpInput, callDataSize, bbs, ebs);
      // codeOwnerAccount owns the bytecode that will be given as input to MODEXP through
      // EXTCODECOPY
      final ToyAccount codeOwnerAccount =
          ToyAccount.builder()
              .balance(Wei.of(0))
              .nonce(1)
              .address(codeOwnerAddress)
              .code(modexpInput)
              .build();
      additionalAccounts = List.of(codeOwnerAccount);

      prepareModexp(program, modexpInput, callDataOffset, codeOwnerAddress);
    } else {
      // ECADD, ECMUL, ECRECOVER cases
      callDataSize = 1; // This is an arbitrary value
    }

    // Set returnAtCapacity equal to the expected return size of the precompile call
    final int returnAtCapacity = getExpectedReturnAtCapacity(precompileAddress, callDataSize, mbs);

    // Compute the precompile cost
    final int precompileCost =
        getPrecompileCost(precompileAddress, callDataSize, bbs, mbs, exponentLog, r);

    // Compute the gas parameter of the CALL in the different testing scenarios
    int gas = getGas(gasCase, precompileCost);

    // In case funds are sent to the precompile contract (valueCase == NON_ZERO)
    // a gas stipend of 2300 is added to the precompile's frame initial gas.
    // We now deduce that gas stipend from the gas given to the transaction to trigger
    // insufficient gas for the precompile call in the non-trivial cases (COST_MINUS_ONE, COST,
    // COST_PLUS_ONE).
    // Note that we exclude the case of MODEXP as it is treated in a separate test
    // and the case of ECADD as it has a fixed gas cost of 150 < 2300.
    if (valueCase.isNonZeroCase()
        && (gasCase == GasCase.COST_MINUS_ONE
            || gasCase == GasCase.COST
            || gasCase == GasCase.COST_PLUS_ONE)
        && !precompileAddress.equals(ALTBN128_ADD)
        && !precompileAddress.equals(MODEXP)) {
      gas -= GAS_CONST_G_CALL_STIPEND;
    }

    // Common program for all precompile calls
    program
        .push(returnAtCapacity) // returnAtCapacity
        .push(returnAtOffset) // returnAtOffset
        .push(callDataSize) // callDataSize
        .push(callDataOffset) // callDataOffset
        .push(value) // value
        .push(precompileAddress) // address
        .push(gas) // gas
        .op(OpCode.CALL);
    final BytecodeRunner bytecodeRunner = BytecodeRunner.of(program);
    bytecodeRunner.run(61_000_000L, precompileAddress == MODEXP ? additionalAccounts : List.of());
    final Hub hub = bytecodeRunner.getHub();

    // Here we check if OOB detects the insufficient gas for the precompile call
    // and the precompile cost computed by OOB.
    // As the number of OOB operation required is variable, we look for it over all the operations.
    boolean insufficientGasForPrecompile =
        hub.oob().operations().getAll().stream()
            .anyMatch(OobOperation::isInsufficientGasForPrecompile);

    BigInteger precompileCostComputedByOOB =
        hub.oob().operations().getAll().stream()
            .map(OobOperation::getPrecompileCost)
            .filter(Objects::nonNull)
            .findFirst()
            .orElse(BigInteger.ZERO);

    // We assert that the precompileCost we compute here is the same as the one computed in OOB
    assertEquals(BigInteger.valueOf(precompileCost), precompileCostComputedByOOB);

    // We assert that the insufficientGasForPrecompile flag is set correctly in OOB
    if (gasCase == GasCase.COST
        || gasCase == GasCase.COST_PLUS_ONE
        || (precompileAddress.equals(BLAKE2B_F_COMPRESSION)
            && r == 0) // precompileCost is 0 so gas cannot be insufficient
        || (precompileAddress.equals(ALTBN128_ADD)
            && value > 0) // precompileCost is 150 but stipend is at least 2300 so gas cannot be
    // insufficient
    ) {
      assertFalse(insufficientGasForPrecompile);
    } else {
      assertTrue(insufficientGasForPrecompile);
    }
  }

  static Stream<Arguments> lowGasStipendPrecompileCallTestSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (GasCase gasCase : GasCase.values()) {
      for (ValueCase valueCase : ValueCase.values()) {
        arguments.add(Arguments.of(ECREC, valueCase, gasCase, false));
        arguments.add(Arguments.of(SHA256, valueCase, gasCase, false));
        arguments.add(Arguments.of(RIPEMD160, valueCase, gasCase, false));
        arguments.add(Arguments.of(ID, valueCase, gasCase, false));
        arguments.add(Arguments.of(ALTBN128_ADD, valueCase, gasCase, false));
        arguments.add(Arguments.of(ALTBN128_MUL, valueCase, gasCase, false));
        arguments.add(Arguments.of(Address.ALTBN128_PAIRING, valueCase, gasCase, false));
        arguments.add(Arguments.of(BLAKE2B_F_COMPRESSION, valueCase, gasCase, false));
        arguments.add(Arguments.of(BLAKE2B_F_COMPRESSION, valueCase, gasCase, true));
      }
      // The NON_ZERO for MODEXP case will be treated in a separate test
      arguments.add(Arguments.of(MODEXP, ValueCase.ZERO, gasCase, false));
      arguments.add(Arguments.of(MODEXP, ValueCase.ZERO, gasCase, true));
    }
    return arguments.stream();
  }

  // Support methods
  /**
   * Computes the gas based on the gas parameter and precompile cost.
   *
   * @param gasCase the gas case.
   * @param precompileCost the precompile cost.
   * @return the computed gas.
   */
  private static int getGas(GasCase gasCase, int precompileCost) {
    return switch (gasCase) {
      case ZERO -> 0;
      case ONE -> 1;
      case COST_MINUS_ONE -> precompileCost - 1;
      case COST -> precompileCost;
      case COST_PLUS_ONE -> precompileCost + 1;
    };
  }
}
