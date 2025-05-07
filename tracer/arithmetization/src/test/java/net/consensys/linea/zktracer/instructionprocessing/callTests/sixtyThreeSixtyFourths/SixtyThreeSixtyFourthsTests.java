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

package net.consensys.linea.zktracer.instructionprocessing.callTests.sixtyThreeSixtyFourths;

import static com.google.common.base.Preconditions.checkArgument;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_G_CALL_STIPEND;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_G_CALL_VALUE;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_G_NEW_ACCOUNT;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_G_WARM_ACCESS;
import static net.consensys.linea.zktracer.Trace.PRC_BLAKE2F_SIZE;
import static net.consensys.linea.zktracer.Trace.WORD_SIZE;
import static net.consensys.linea.zktracer.module.hub.signals.TracedException.OUT_OF_GAS_EXCEPTION;
import static net.consensys.linea.zktracer.opcode.OpCode.CALL;
import static net.consensys.linea.zktracer.opcode.OpCode.MLOAD;
import static net.consensys.linea.zktracer.opcode.OpCode.POP;
import static net.consensys.linea.zktracer.precompiles.LowGasStipendPrecompileCallTests.computeExponentLog;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.generateModexpInput;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.getBLAKE2FCost;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.getECADDCost;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.getMODEXPCost;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.getPrecompileCost;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.prepareBlake2F;
import static net.consensys.linea.zktracer.precompiles.PrecompileUtils.prepareModexp;
import static org.hyperledger.besu.datatypes.Address.ALTBN128_ADD;
import static org.hyperledger.besu.datatypes.Address.ALTBN128_MUL;
import static org.hyperledger.besu.datatypes.Address.ALTBN128_PAIRING;
import static org.hyperledger.besu.datatypes.Address.BLAKE2B_F_COMPRESSION;
import static org.hyperledger.besu.datatypes.Address.ECREC;
import static org.hyperledger.besu.datatypes.Address.ID;
import static org.hyperledger.besu.datatypes.Address.MODEXP;
import static org.hyperledger.besu.datatypes.Address.RIPEMD160;
import static org.hyperledger.besu.datatypes.Address.SHA256;
import static org.junit.jupiter.api.Assertions.assertNotEquals;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.testing.ToyAccount;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

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

public class SixtyThreeSixtyFourthsTests {

  /*
  Cases to cover:

  * value = 0
  If precompileGasCost >= 2300 then we are interested in:
   * value = 1, targetAddressExists = false
   * value = 1, targetAddressExists = true

  Note: BLAKE2F and MODEXP requires a non-zero non-trivial input to have a cost greater than 2300.
  Other precompiles only require a proper call data size at most.

  See https://github.com/Consensys/linea-tracer/issues/1153 for additional documentation.
  */

  static final Bytes INFINITE_GAS = Bytes.fromHexString("ff".repeat(32));

  // BLAKE2F specific parameters
  static final int rLeadingByte = 0x09;
  static final int r = rLeadingByte << 8;

  // MODEXP specific parameters
  static final int bbs = 2;
  static final int ebs = 6;
  static final int mbs = 128;
  static final Bytes modexpInput = generateModexpInput(bbs, mbs, ebs);
  static final int exponentLog = computeExponentLog(modexpInput, 96 + bbs + ebs + mbs, bbs, ebs);
  static final Address codeOwnerAddress = Address.fromHexString("0xC0DE");
  // codeOwnerAccount owns the bytecode that will be given as input to MODEXP through EXTCODECOPY
  static final ToyAccount codeOwnerAccount =
      ToyAccount.builder()
          .balance(Wei.of(0))
          .nonce(1)
          .address(codeOwnerAddress)
          .code(modexpInput)
          .build();
  static final List<ToyAccount> additionalAccounts = List.of(codeOwnerAccount);

  // Cost of preCallProgram in different scenarios:
  // (address, transfersValue) -> gasCost
  static final Map<Address, Map<Boolean, Long>> preCallProgramGasMap =
      Stream.of(
              ECREC,
              SHA256,
              RIPEMD160,
              ID,
              MODEXP,
              ALTBN128_ADD,
              ALTBN128_MUL,
              ALTBN128_PAIRING,
              BLAKE2B_F_COMPRESSION)
          .collect(
              Collectors.toMap(
                  address -> address,
                  address ->
                      new HashMap<>() {
                        {
                          put(
                              false,
                              BytecodeRunner.of(preCallProgram(address, false, false, 0))
                                  .runOnlyForGasCost(
                                      address == MODEXP ? additionalAccounts : List.of()));
                          put(
                              true,
                              BytecodeRunner.of(preCallProgram(address, false, true, 0))
                                  .runOnlyForGasCost(
                                      address == MODEXP ? additionalAccounts : List.of()));
                        }
                      }));

  // Note: transferValue = false and cds = 0 as we are interested only in the cost of the
  // corresponding PUSHes here

  /**
   * Parameterized test for the ECADD precompile, that has a fixed cost of 150.
   *
   * @param gasLimit the gas limit for the transaction. It is either as much as needed for the ECADD
   *     call or slightly less.
   * @param insufficientGasForPrecompileExpected flag indicating if insufficient gas for ECADD is
   *     expected.
   */
  @ParameterizedTest
  @MethodSource("fixedCostEcAddTestSource")
  void fixedCostEcAddTest(long gasLimit, boolean insufficientGasForPrecompileExpected) {
    // Whenever transferValue = true, gas is enough
    // so we only test the case in which transferValue = false

    final BytecodeCompiler program = BytecodeCompiler.newProgram();

    program.immediate(preCallProgram(ALTBN128_ADD, false, false, 0)).op(CALL);

    final BytecodeRunner bytecodeRunner = BytecodeRunner.of(program);
    bytecodeRunner.run(gasLimit);

    assertNotEquals(
        OUT_OF_GAS_EXCEPTION,
        bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
  }

  static Stream<Arguments> fixedCostEcAddTestSource() {
    List<Arguments> arguments = new ArrayList<>();
    final long targetCalleeGas = getECADDCost();
    for (int cornerCase : List.of(0, -1)) {
      final long gasLimit =
          getGasLimit(
              targetCalleeGas + cornerCase,
              false,
              false,
              preCallProgramGasMap.get(ALTBN128_ADD).get(false));
      arguments.add(Arguments.of(gasLimit, cornerCase == -1));
    }
    return arguments.stream();
  }

  /**
   * Parameterized test for precompile calls where the cost is greater than or equal to the stipend
   * (every precompile except ECADD, as long as cds and inputs are properly selected).
   *
   * @param address the address of the precompile contract.
   * @param gasLimit the gas limit for the transaction. It is either as much as needed for the
   *     precompile call or slightly less.
   * @param insufficientGasForPrecompileExpected flag indicating if insufficient gas for precompile
   *     call is expected.
   * @param transfersValue flag indicating if the call to the precompile transfers value.
   * @param targetAddressExists flag indicating if the precompile target address exists at the
   *     moment of the final call.
   * @param cds the call data size.
   */
  @ParameterizedTest
  @MethodSource("costGEQStipendTest")
  void costGEQStipendTest(
      Address address,
      long gasLimit,
      boolean insufficientGasForPrecompileExpected,
      boolean transfersValue,
      boolean targetAddressExists,
      int cds) {
    final BytecodeCompiler program = BytecodeCompiler.newProgram();
    program.immediate(preCallProgram(address, transfersValue, targetAddressExists, cds)).op(CALL);

    final BytecodeRunner bytecodeRunner = BytecodeRunner.of(program);
    bytecodeRunner.run(gasLimit, address == MODEXP ? additionalAccounts : List.of());

    assertNotEquals(
        OUT_OF_GAS_EXCEPTION,
        bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
  }

  static Stream<Arguments> costGEQStipendTest() {
    List<Arguments> arguments = new ArrayList<>();
    for (Address address :
        List.of(
            ECREC,
            SHA256,
            RIPEMD160,
            ID,
            MODEXP,
            ALTBN128_MUL,
            ALTBN128_PAIRING,
            BLAKE2B_F_COMPRESSION)) {
      final int cds = getCallDataSize(address);
      final long targetCalleeGas =
          address == BLAKE2B_F_COMPRESSION
              ? getBLAKE2FCost(r)
              : address == MODEXP
                  ? getMODEXPCost(bbs, mbs, exponentLog)
                  : getPrecompileCost(address, cds);
      for (int cornerCase : List.of(0, -1)) {
        for (boolean transfersValue : List.of(true, false)) {
          for (boolean targetAddressExists : List.of(true, false)) {
            if (!transfersValue && targetAddressExists) {
              continue; // no need to test this case
            }
            final long gasLimit =
                getGasLimit(
                    targetCalleeGas + cornerCase,
                    transfersValue,
                    targetAddressExists,
                    preCallProgramGasMap.get(address).get(targetAddressExists));
            arguments.add(
                Arguments.of(
                    address, gasLimit, cornerCase == -1, transfersValue, targetAddressExists, cds));
          }
        }
      }
    }
    return arguments.stream();
  }

  // Support methods
  static Bytes preCallProgram(
      Address address, boolean transfersValue, boolean targetAddressExists, int cds) {
    return BytecodeCompiler.newProgram()
        .immediate(expandMemoryTo2048Words())
        .immediate(targetAddressExists ? successfullySummonIntoExistence(address) : Bytes.EMPTY)
        .immediate(
            address == MODEXP ? prepareModexp(modexpInput, 0, codeOwnerAddress) : Bytes.EMPTY)
        .immediate(address == BLAKE2B_F_COMPRESSION ? prepareBlake2F(rLeadingByte, 2) : Bytes.EMPTY)
        .immediate(pushCallArguments(INFINITE_GAS, address, cds, transfersValue))
        .compile();
  }

  static Bytes expandMemoryTo2048Words() {
    return expandMemoryTo(2048);
  }

  static Bytes expandMemoryTo(int words) {
    checkArgument(words >= 1);
    return BytecodeCompiler.newProgram().push((words - 1) * WORD_SIZE).op(MLOAD).op(POP).compile();
  }

  static Bytes successfullySummonIntoExistence(Address address) {
    return call(
        INFINITE_GAS,
        address,
        address == BLAKE2B_F_COMPRESSION
            ? PRC_BLAKE2F_SIZE
            : 0, // For BLAKE2F we need a meaningful cds for the call to succeed
        true);
  }

  static Bytes call(Bytes gas, Address address, int cds, boolean transfersValue) {
    return BytecodeCompiler.newProgram()
        .immediate(pushCallArguments(gas, address, cds, transfersValue))
        .op(CALL)
        .compile();
  }

  static Bytes pushCallArguments(Bytes gas, Address address, int cds, boolean transfersValue) {
    return BytecodeCompiler.newProgram()
        .push(0) // returnAtCapacity
        .push(0) // returnAtOffset
        .push(cds) // callDataSize
        .push(0) // callDataOffset
        .push(transfersValue ? 1 : 0) // value
        .push(address) // address
        .push(gas) // gas
        .compile();
  }

  /**
   * Calculates the gas limit for a transaction to a contract calling a callee (in this class, a
   * precompile), based on targetCallGas, transfersValue, targetAddressExists, and
   * preCallProgramGas.
   *
   * @param targetCalleeGas the gas given to the callee.
   * @param transfersValue flag indicating if the call to callee transfers value.
   * @param targetAddressExists flag indicating if the target address exists.
   * @param preCallProgramGas the gas cost of the preCallProgram.
   * @return the calculated gas limit for the transaction.
   */
  static long getGasLimit(
      long targetCalleeGas,
      boolean transfersValue,
      boolean targetAddressExists,
      long preCallProgramGas) {
    /* gasLimit = preCallProgramGasCost + gasPreCall
    /  let x = gasPreCall - gasUpFront = 64 * k + l
    /  with:
    /  x !≡ 63 % 64 => (x - 63) % 64 != 0
    /  k = floor(x / 64)
    /  l = x % 64
    /  63 * k + l + stipend = targetCalleeGas
    /  find gasLimit going backwards
    */
    final long stipend = transfersValue ? GAS_CONST_G_CALL_STIPEND : 0;
    checkArgument(targetCalleeGas >= stipend);
    final long l = (targetCalleeGas - stipend) % 63;
    final long k = (targetCalleeGas - stipend - l) / 63;
    checkArgument(targetCalleeGas == 63 * k + l + stipend);
    final long gasUpfront = getUpfrontGasCost(transfersValue, targetAddressExists);
    final long gasPreCall = 64 * k + l + gasUpfront;
    return preCallProgramGas + gasPreCall; // gasLimit
  }

  /**
   * Calculates the upfront gas cost of a CALL-type instruction based on whether it transfers value
   * and if the target address exists. It assumed target address is warm and there is no memory
   * expansion.
   *
   * @param transfersValue flag indicating if the call transfers value.
   * @param targetAddressExists flag indicating if the target address exists.
   * @return the upfront gas cost for the call.
   */
  static long getUpfrontGasCost(boolean transfersValue, boolean targetAddressExists) {
    // GAS_CONST_G_WARM_ACCESS = 100
    // GAS_CONST_G_COLD_ACCOUNT_ACCESS = 2600
    // GAS_CONST_G_CALL_VALUE = 9000
    // GAS_CONST_G_NEW_ACCOUNT = 25000
    return GAS_CONST_G_WARM_ACCESS
        + (transfersValue
            ? GAS_CONST_G_CALL_VALUE + (targetAddressExists ? 0 : GAS_CONST_G_NEW_ACCOUNT)
            : 0);
  }

  static int getCallDataSize(Address address) {
    if (address == SHA256 || address == RIPEMD160 || address == ID) {
      return 1024 * WORD_SIZE; // Ensures cost is greater than stipend
    } else if (address == MODEXP) {
      return 96 + bbs + ebs + mbs; // Ensures cost is greater than stipend with non-zero non-trivial
    } else if (address == BLAKE2B_F_COMPRESSION) {
      return PRC_BLAKE2F_SIZE; // Ensures cost is greater than stipend with non-zero non-trivial
      // input
    } else {
      return 0;
    }
  }
}
