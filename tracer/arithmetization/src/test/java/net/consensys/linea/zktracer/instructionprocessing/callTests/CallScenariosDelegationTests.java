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

package net.consensys.linea.zktracer.instructionprocessing.callTests;

import static net.consensys.linea.zktracer.instructionprocessing.utilities.Calls.appendFullGasCall;

import java.util.ArrayList;
import java.util.List;
import java.util.function.BiFunction;
import java.util.function.Function;
import java.util.stream.Stream;
import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyExecutionEnvironmentV2;
import net.consensys.linea.testing.ToyTransaction;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@ExtendWith(UnitTestWatcher.class)
public class CallScenariosDelegationTests extends TracerTestBase {

  /*
    https://github.com/Consensys/linea-monorepo/issues/2470

                                                  ┌--------------┐
                                                  |this is where |
                                                  |the scenario  |
                                                  |should apply  |
  tx    --->    root ---[CALL]--->    caller   ---[CALL-type inst]--->    callee
                 |                      |                                   |
                 |                      |                                   |
                 |                      V                                   V
                 |                [?] delegt SMC 1                    [?] delegt SMC 2
                 |
                 |
                 |
                 └------[CALL]--->    caller   ---[CALL-type inst]--->    callee
                 |                      |                                   |
                 |                      |                                   |
                 |                      V                                   V
                 |                [?] delegt SMC 1                    [?] delegt SMC 2
                 |
                 |
                 |
                 └------[further test]
     */

  static final KeyPair senderKeyPair = new SECP256K1().generateKeyPair();
  static final Address senderAddress =
      Address.extract(Hash.hash(senderKeyPair.getPublicKey().getEncodedBytes()));
  static final ToyAccount senderAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(10))
          .nonce(42)
          .address(senderAddress)
          .keyPair(senderKeyPair)
          .build();

  static final ToyAccount rootAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(2))
          .nonce(67)
          .address(Address.fromHexString("0x40010000"))
          .build();

  static final ToyAccount callerAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(3))
          .nonce(69)
          .address(Address.fromHexString("0xCA77E400"))
          .build();

  static final ToyAccount calleeAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(4))
          .nonce(90)
          .address(Address.fromHexString("0xCA77EE00"))
          .build();

  static final ToyAccount smcAccount1 =
      ToyAccount.builder()
          .balance(Wei.fromEth(5))
          .nonce(101)
          .address(Address.fromHexString("0xDDCA77E4"))
          .build();

  static final ToyAccount smcAccount2 =
      ToyAccount.builder()
          .balance(Wei.fromEth(6))
          .nonce(666)
          .address(Address.fromHexString("0xDDCA77EE"))
          .build();

  BiFunction<ToyAccount, RevertType, BytecodeCompiler> conditionalCallProgram =
      (targetAccount, revertType) ->
          BytecodeCompiler.newProgram(chainConfig)
              .push(0)
              .op(OpCode.SLOAD) // LOOP_DEPTH_CURRENT
              .push(3) // LOOP_DEPTH_MAX
              .op(OpCode.GT) // LOOP_DEPTH_MAX > LOOP_DEPTH_CURRENT
              .push(10)
              .op(OpCode.JUMPI) // if LOOP_DEPTH_CURRENT < LOOP_DEPTH_MAX jump to JUMPDEST else STOP
              .op(OpCode.STOP)
              .op(OpCode.JUMPDEST) // PC = 10
              .push(0)
              .op(OpCode.SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(OpCode.SSTORE) // increment LOOP_DEPTH_CURRENT by 1
              // execute the call
              .apply(
                  program ->
                      appendFullGasCall(
                          program, OpCode.CALL, targetAccount.getAddress(), 0, 0, 0, 0, 0))
              // preparing for a potential revert
              .push(0)
              .push(0)
              .op(revertType == RevertType.TERMINATES_ON_REVERT ? OpCode.REVERT : OpCode.STOP);

  BiFunction<ToyAccount, RevertType, BytecodeCompiler> callProgram =
      (targetAccount, revertType) ->
          BytecodeCompiler.newProgram(chainConfig)
              .push(0)
              .op(OpCode.SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(OpCode.SSTORE) // increment LOOP_DEPTH_CURRENT by 1
              // execute the call
              .apply(
                  program ->
                      appendFullGasCall(
                          program, OpCode.CALL, targetAccount.getAddress(), 0, 0, 0, 0, 0))
              // preparing for a potential revert
              .push(0)
              .push(0)
              .op(revertType == RevertType.TERMINATES_ON_REVERT ? OpCode.REVERT : OpCode.STOP);

  private record RootProgramParams(
      CallScenarioFragment.CallScenario callScenario, OpCode callOpCode, RevertType revertType) {}

  Function<RootProgramParams, BytecodeCompiler> rootProgram =
      par ->
          BytecodeCompiler.newProgram(chainConfig)
              .push(0)
              .op(OpCode.SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(OpCode.SSTORE) // increment LOOP_DEPTH_CURRENT by 1
              // execute the call
              .apply(
                  program ->
                      appendFullGasCall(
                          program, par.callOpCode, callerAccount.getAddress(), 0, 0, 0, 0, 0))
              // execute the call again
              .apply(
                  program ->
                      appendFullGasCall(
                          program, par.callOpCode, callerAccount.getAddress(), 0, 0, 0, 0, 0))
              .push(rootAccount.getAddress())
              .op(OpCode.BALANCE)
              .op(OpCode.POP)
              .push(callerAccount.getAddress())
              .op(OpCode.BALANCE)
              .op(OpCode.POP)
              .push(calleeAccount.getAddress())
              .op(OpCode.BALANCE)
              .op(OpCode.POP)
              // preparing for a potential revert
              .push(0)
              .push(0)
              .op(par.revertType == RevertType.TERMINATES_ON_REVERT ? OpCode.REVERT : OpCode.STOP);

  public enum CallerType {
    DELEGATED,
    SMC // already tested
  }

  public enum CalleeType {
    // the first few we don't really care about: they don't lead to execution
    // DELEGATED_TO_NON_EXISTENT,
    // DELEGATED_TO_EMPTY_CODE_ACCOUNT,
    // DELEGATED_TO_PRC,
    // DELEGATED_TO_SELF,
    // most relevant cases
    DELEGATED_TO_ROOT,
    DELEGATED_TO_CALLER,
    DELEGATED_TO_SMC,
    SMC // already tested
  }

  // this should apply per smart contract
  public enum RevertType {
    TERMINATES_ON_REVERT,
    TERMINATES_ON_NON_REVERT;
  }

  // this should apply uniformly to all smart contracts
  public enum LoopType {
    INFINITE_LOOP,
    EXIT_EARLY;
  }

  @ParameterizedTest
  @MethodSource("callaDelegationTestSource")
  public void callDelegationTest(
      CallerType callerType,
      CalleeType calleeType,
      RevertType rootCodeRevertType,
      RevertType callerCodeRevertType,
      RevertType calleeCodeRevertType,
      LoopType loopType,
      TestInfo testInfo) {
    BiFunction<ToyAccount, RevertType, BytecodeCompiler> actualCallProgram =
        loopType == LoopType.EXIT_EARLY ? conditionalCallProgram : callProgram;

    rootAccount.setCode(actualCallProgram.apply(callerAccount, rootCodeRevertType).compile());

    switch (callerType) {
      case DELEGATED -> {
        callerAccount.delegateTo(smcAccount1);
        smcAccount1.setCode(actualCallProgram.apply(calleeAccount, callerCodeRevertType).compile());
      }
      case SMC ->
          callerAccount.setCode(
              actualCallProgram.apply(calleeAccount, callerCodeRevertType).compile());
    }

    switch (calleeType) {
      case DELEGATED_TO_ROOT -> calleeAccount.delegateTo(rootAccount);
      case DELEGATED_TO_CALLER -> calleeAccount.delegateTo(callerAccount);
      case DELEGATED_TO_SMC -> {
        calleeAccount.delegateTo(smcAccount2);
        smcAccount2.setCode(
            actualCallProgram
                .apply(callerAccount, calleeCodeRevertType)
                .compile()); // This could be a call to anything
      }
      case SMC ->
          calleeAccount.setCode(
              actualCallProgram
                  .apply(callerAccount, calleeCodeRevertType)
                  .compile()); // This could be a call to anything
    }

    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(rootAccount)
            .keyPair(senderKeyPair)
            .gasLimit(100_000L)
            .build();

    ToyExecutionEnvironmentV2 toyExecutionEnvironmentV2 =
        ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
            .accounts(
                List.of(
                    senderAccount,
                    rootAccount,
                    callerAccount,
                    calleeAccount,
                    smcAccount1,
                    smcAccount2))
            .transaction(tx)
            .build();
    toyExecutionEnvironmentV2.run();
  }

  static Stream<Arguments> callaDelegationTestSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (CallerType callerType : CallerType.values()) {
      for (CalleeType calleeType : CalleeType.values()) {
        for (RevertType rootCodeRevertType : RevertType.values()) {
          for (RevertType callerCodeRevertType : RevertType.values()) {
            for (RevertType calleeCodeRevertType : RevertType.values()) {
              for (LoopType loopType : LoopType.values()) {
                arguments.add(
                    Arguments.of(
                        callerType,
                        calleeType,
                        rootCodeRevertType,
                        callerCodeRevertType,
                        calleeCodeRevertType,
                        loopType));
              }
            }
          }
        }
      }
    }
    return arguments.stream();
  }
}
