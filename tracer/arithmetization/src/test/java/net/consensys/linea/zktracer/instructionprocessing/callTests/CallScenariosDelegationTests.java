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

import static net.consensys.linea.testing.BytecodeRunner.MAX_GAS_LIMIT;
import static net.consensys.linea.zktracer.instructionprocessing.utilities.Calls.appendCall;
import static net.consensys.linea.zktracer.instructionprocessing.utilities.Calls.appendFullGasCall;
import static net.consensys.linea.zktracer.instructionprocessing.utilities.Calls.appendInsufficientBalanceCall;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment.CallScenario.CALL_ABORT_WILL_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment.CallScenario.CALL_ABORT_WONT_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment.CallScenario.CALL_EOA_SUCCESS_WILL_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment.CallScenario.CALL_EOA_SUCCESS_WONT_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment.CallScenario.CALL_EXCEPTION;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment.CallScenario.CALL_SMC_FAILURE_WILL_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment.CallScenario.CALL_SMC_FAILURE_WONT_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment.CallScenario.CALL_SMC_SUCCESS_WILL_REVERT;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.CallScenarioFragment.CallScenario.CALL_SMC_SUCCESS_WONT_REVERT;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;
import kotlin.jvm.functions.Function2;
import kotlin.jvm.functions.Function4;
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

  final KeyPair senderKeyPair = new SECP256K1().generateKeyPair();
  final Address senderAddress =
      Address.extract(Hash.hash(senderKeyPair.getPublicKey().getEncodedBytes()));
  final ToyAccount senderAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(10))
          .nonce(42)
          .address(senderAddress)
          .keyPair(senderKeyPair)
          .build();

  final ToyAccount rootAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(2))
          .nonce(67)
          .address(Address.fromHexString("0x40010000"))
          .build();

  final ToyAccount callerAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(3))
          .nonce(69)
          .address(Address.fromHexString("0xCA77E400"))
          .build();

  final ToyAccount calleeAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(4))
          .nonce(90)
          .address(Address.fromHexString("0xCA77EE00"))
          .build();

  final ToyAccount smcAccount1 =
      ToyAccount.builder()
          .balance(Wei.fromEth(5))
          .nonce(101)
          .address(Address.fromHexString("0xDDCA77E4"))
          .build();

  final ToyAccount smcAccount2 =
      ToyAccount.builder()
          .balance(Wei.fromEth(6))
          .nonce(666)
          .address(Address.fromHexString("0xDDCA77EE"))
          .build();

  Function4<CallScenarioFragment.CallScenario, ToyAccount, LoopType, RevertType, BytecodeCompiler>
      callProgram =
          (callScenario, targetAccount, loopType, revertType) ->
              BytecodeCompiler.newProgram(chainConfig)
                  .immediate(
                      loopType == LoopType.EXIT_EARLY,
                      BytecodeCompiler.newProgram(chainConfig)
                          .push(0)
                          .op(OpCode.SLOAD) // LOOP_DEPTH_CURRENT
                          .push(3) // LOOP_DEPTH_MAX
                          .op(OpCode.GT) // LOOP_DEPTH_MAX > LOOP_DEPTH_CURRENT
                          .push(10)
                          .op(
                              OpCode
                                  .JUMPI) // if LOOP_DEPTH_CURRENT < LOOP_DEPTH_MAX jump to JUMPDEST
                          // else STOP
                          .op(OpCode.STOP)
                          .op(OpCode.JUMPDEST) // PC = 10
                          .compile())
                  .push(0)
                  .op(OpCode.SLOAD)
                  .push(1)
                  .op(OpCode.ADD)
                  .push(0)
                  .op(OpCode.SSTORE) // increment LOOP_DEPTH_CURRENT by 1
                  // execute the call
                  .apply(
                      program ->
                          switch (callScenario) {
                            case CALL_EXCEPTION ->
                                // triggers a staticFault as the target account tries to execute
                                // SSTORE
                                appendFullGasCall(
                                    program,
                                    OpCode.STATICCALL,
                                    targetAccount.getAddress(),
                                    0,
                                    0,
                                    0,
                                    0,
                                    0);
                            case CALL_ABORT_WILL_REVERT, CALL_ABORT_WONT_REVERT ->
                                appendInsufficientBalanceCall(
                                    program, OpCode.CALL, targetAccount.getAddress(), 0, 0, 0, 0);
                            default ->
                                appendFullGasCall(
                                    program,
                                    OpCode.CALL,
                                    targetAccount.getAddress(),
                                    0,
                                    0,
                                    0,
                                    0,
                                    0);
                          })
                  // preparing for a potential revert
                  .push(0)
                  .push(0)
                  .op(revertType == RevertType.TERMINATES_ON_REVERT ? OpCode.REVERT : OpCode.STOP);

  Function2<LoopType, RevertType, BytecodeCompiler> rootProgram =
      (loopType, revertType) ->
          BytecodeCompiler.newProgram(chainConfig)
              .immediate(
                  loopType == LoopType.EXIT_EARLY,
                  BytecodeCompiler.newProgram(chainConfig)
                      .push(0)
                      .op(OpCode.SLOAD) // LOOP_DEPTH_CURRENT
                      .push(3) // LOOP_DEPTH_MAX
                      .op(OpCode.GT) // LOOP_DEPTH_MAX > LOOP_DEPTH_CURRENT
                      .push(10)
                      .op(OpCode.JUMPI) // if LOOP_DEPTH_CURRENT < LOOP_DEPTH_MAX jump to JUMPDEST
                      // else STOP
                      .op(OpCode.STOP)
                      .op(OpCode.JUMPDEST) // PC = 10
                      .compile())
              .push(0)
              .op(OpCode.SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(OpCode.SSTORE) // increment LOOP_DEPTH_CURRENT by 1
              // execute 2 identical calls to callerAccount
              .apply(
                  program ->
                      appendCall(
                          program, OpCode.CALL, 100_000, callerAccount.getAddress(), 0, 0, 0, 0, 0),
                  2)
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
              .op(revertType == RevertType.TERMINATES_ON_REVERT ? OpCode.REVERT : OpCode.STOP);

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
    SMC; // already tested

    public boolean isDelegatedEOA() {
      return this == DELEGATED_TO_ROOT || this == DELEGATED_TO_CALLER || this == DELEGATED_TO_SMC;
    }

    public boolean isSmc() {
      return this == SMC;
    }
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
  @MethodSource("callScenariosDelegationTestsSource")
  public void callScenariosDelegationTests(
      CallScenarioFragment.CallScenario callScenario,
      CallerType callerType,
      CalleeType calleeType,
      RevertType rootCodeRevertType,
      RevertType calleeCodeRevertType,
      LoopType loopType,
      TestInfo testInfo) {
    /*
     Scenarios and how they are triggered:
     CALL_EXCEPTION: caller STATICCALLs to the callee which tries to execute SSTORE
     CALL_ABORT_WILL_REVERT: caller CALLs to the callee with insufficient balance to cover value transfer, then reverts
     CALL_ABORT_WONT_REVERT: caller CALLs to the callee with insufficient balance to cover value transfer, but doesn't revert
     CALL_EOA_SUCCESS_WILL_REVERT: caller CALLs to an EOA which successfully processes the call but then reverts
     CALL_EOA_SUCCESS_WONT_REVERT: caller CALLs to an EOA which successfully processes the call and doesn't revert
     CALL_SMC_FAILURE_WILL_REVERT: caller CALLs to an SMC which fails (e.g., by executing the INVALID opcode) and reverts
     CALL_SMC_FAILURE_WONT_REVERT: caller CALLs to an SMC which fails (e.g., by executing the INVALID opcode) but doesn't revert
     CALL_SMC_SUCCESS_WILL_REVERT: caller CALLs to an SMC which successfully processes the call but then reverts
     CALL_SMC_SUCCESS_WONT_REVERT: caller CALLs to an SMC which successfully processes the call and doesn't revert
    */
    rootAccount.setCode(rootProgram.invoke(loopType, rootCodeRevertType).compile());

    switch (callerType) {
      case DELEGATED -> {
        callerAccount.delegateTo(smcAccount1);
        smcAccount1.setCode(
            callProgram
                .invoke(
                    callScenario,
                    calleeAccount,
                    loopType,
                    callScenario.isWillRevertCallScenario()
                        ? RevertType.TERMINATES_ON_REVERT
                        : RevertType.TERMINATES_ON_NON_REVERT)
                .compile());
      }
      case SMC ->
          callerAccount.setCode(
              callProgram
                  .invoke(
                      callScenario,
                      calleeAccount,
                      loopType,
                      callScenario.isWillRevertCallScenario()
                          ? RevertType.TERMINATES_ON_REVERT
                          : RevertType.TERMINATES_ON_NON_REVERT)
                  .compile());
    }

    // CALL_EOA scenarios are implicitly covered by DELEGATED_TO_ROOT, DELEGATED_TO_CALLER and
    // DELEGATED_TO_SMC cases
    switch (calleeType) {
      case DELEGATED_TO_ROOT -> calleeAccount.delegateTo(rootAccount);
      case DELEGATED_TO_CALLER -> calleeAccount.delegateTo(callerAccount);
      case DELEGATED_TO_SMC -> {
        calleeAccount.delegateTo(smcAccount2);
        smcAccount2.setCode(
            callScenario.isFailureCallScenario()
                ? BytecodeCompiler.newProgram(chainConfig).op(OpCode.INVALID).compile()
                : callProgram
                    .invoke(callScenario, callerAccount, loopType, calleeCodeRevertType)
                    .compile()); // This could be a call to anything
      }
      case SMC -> {
        calleeAccount.setCode(
            callScenario.isFailureCallScenario()
                ? BytecodeCompiler.newProgram(chainConfig).op(OpCode.INVALID).compile()
                : callProgram
                    .invoke(callScenario, callerAccount, loopType, calleeCodeRevertType)
                    .compile()); // This could be a call to anything
      }
    }

    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(rootAccount)
            .keyPair(senderKeyPair)
            .gasLimit(MAX_GAS_LIMIT)
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

  static Stream<Arguments> callScenariosDelegationTestsSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (CallScenarioFragment.CallScenario callScenario :
        new CallScenarioFragment.CallScenario[] {
          CALL_EXCEPTION,
          CALL_ABORT_WILL_REVERT,
          CALL_ABORT_WONT_REVERT,
          CALL_EOA_SUCCESS_WILL_REVERT,
          CALL_EOA_SUCCESS_WONT_REVERT,
          CALL_SMC_FAILURE_WILL_REVERT,
          CALL_SMC_FAILURE_WONT_REVERT,
          CALL_SMC_SUCCESS_WILL_REVERT,
          CALL_SMC_SUCCESS_WONT_REVERT,
        }) {
      for (CalleeType calleeType : CalleeType.values()) {
        if (callScenario.isEoaCallScenario() && !calleeType.isDelegatedEOA()
            || callScenario.isSmcCallScenario() && !calleeType.isSmc()) {
          continue;
          // if scenario is CALL_EOA then callee must be a (delegated) EOA
          // if scenario is CALL_SMC then callee must be an SMC
        }
        for (CallerType callerType : CallerType.values()) {
          for (RevertType rootCodeRevertType : RevertType.values()) {
            for (RevertType calleeCodeRevertType : RevertType.values()) {
              for (LoopType loopType : LoopType.values()) {
                arguments.add(
                    Arguments.of(
                        callScenario,
                        callerType,
                        calleeType,
                        rootCodeRevertType,
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
