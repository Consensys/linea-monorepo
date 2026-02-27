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

import static net.consensys.linea.zktracer.instructionprocessing.utilities.Calls.appendCall;
import static net.consensys.linea.zktracer.instructionprocessing.utilities.Calls.appendFullGasCall;
import static net.consensys.linea.zktracer.instructionprocessing.utilities.Calls.appendInsufficientBalanceCall;
import static net.consensys.linea.zktracer.opcode.OpCode.CALL;
import static net.consensys.linea.zktracer.opcode.OpCode.GAS;
import static net.consensys.linea.zktracer.opcode.OpCode.GT;
import static net.consensys.linea.zktracer.opcode.OpCode.JUMPDEST;
import static net.consensys.linea.zktracer.opcode.OpCode.JUMPI;
import static net.consensys.linea.zktracer.opcode.OpCode.POP;
import static net.consensys.linea.zktracer.opcode.OpCode.REVERT;
import static net.consensys.linea.zktracer.opcode.OpCode.SLOAD;
import static net.consensys.linea.zktracer.opcode.OpCode.SSTORE;
import static net.consensys.linea.zktracer.opcode.OpCode.STATICCALL;
import static net.consensys.linea.zktracer.opcode.OpCode.STOP;

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

/**
 * Tests for CALL scenarios with EIP-7702 delegation at the caller and/or callee level.
 *
 * <p>https://github.com/Consensys/linea-monorepo/issues/2470
 *
 * <p>Structure:
 *
 * <pre>
 *                                                 ┌--------------┐
 *                                                 |this is where |
 *                                                 |the scenario  |
 *                                                 |should apply  |
 * tx  →  root ---[CALL]--→  caller  ---[CALL-type inst]--→  callee
 *                              |                                |
 *                              | ?                              | ?
 *                              V                                V
 *                             SMC1                             SMC2
 * </pre>
 *
 * <p>The root calls the caller with a fixed gas budget, then calls the caller a second time, and
 * finally checks the balances of root, caller, and callee. The caller executes scenario-specific
 * code that calls the callee. Both the caller and callee may be either direct SMCs or EIP-7702
 * delegated accounts.
 */
@ExtendWith(UnitTestWatcher.class)
public class CallScenariosDelegationTests extends TracerTestBase {

  // ── Fixed gas provided by root to the inner call ──────────────────────────
  static final int ROOT_CALL_GAS = 100_000;

  // ── Accounts ──────────────────────────────────────────────────────────────

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

  // smcAccount1 is the delegate target when callerAccount is DELEGATED
  static final ToyAccount smcAccount1 =
      ToyAccount.builder()
          .balance(Wei.fromEth(5))
          .nonce(101)
          .address(Address.fromHexString("0xDDCA77E4"))
          .build();

  // smcAccount2 is the delegate target when calleeAccount is DELEGATED_TO_SMC
  static final ToyAccount smcAccount2 =
      ToyAccount.builder()
          .balance(Wei.fromEth(6))
          .nonce(666)
          .address(Address.fromHexString("0xDDCA77EE"))
          .build();

  // ── Enums ─────────────────────────────────────────────────────────────────

  public enum CallerType {
    DELEGATED,
    SMC
  }

  public enum CalleeType {
    // Uncomment to test non-execution cases:
    // DELEGATED_TO_NON_EXISTENT,
    // DELEGATED_TO_EMPTY_CODE_ACCOUNT,
    // DELEGATED_TO_PRC,
    // DELEGATED_TO_SELF,
    DELEGATED_TO_ROOT,
    DELEGATED_TO_CALLER,
    DELEGATED_TO_SMC,
    SMC
  }

  // Per-contract: does this execution frame terminate via REVERT or not?
  public enum RevertType {
    TERMINATES_ON_REVERT,
    TERMINATES_ON_NON_REVERT
  }

  // Applied uniformly to all contracts that support looping
  public enum LoopType {
    INFINITE_LOOP,
    EXIT_EARLY
  }

  // ── Root-program parameters ────────────────────────────────────────────────

  private record RootProgramParams(OpCode callOpCode, RevertType revertType) {}

  // ── Program builders ──────────────────────────────────────────────────────

  /**
   * Root program: calls {@code callerAccount} twice with a fixed gas budget, then reads the
   * balances of root, caller, and callee, then optionally reverts.
   *
   * <p>The double-call followed by balance checks mirrors "option 2" from the issue description.
   */
  Function<RootProgramParams, BytecodeCompiler> rootProgram =
      par ->
          BytecodeCompiler.newProgram(chainConfig)
              // first call to callerAccount
              .apply(
                  program ->
                      appendCall(
                          program,
                          par.callOpCode,
                          ROOT_CALL_GAS,
                          callerAccount.getAddress(),
                          0,
                          0,
                          0,
                          0,
                          0))
              .op(POP) // discard return value
              // second call to callerAccount
              .apply(
                  program ->
                      appendCall(
                          program,
                          par.callOpCode,
                          ROOT_CALL_GAS,
                          callerAccount.getAddress(),
                          0,
                          0,
                          0,
                          0,
                          0))
              .op(POP) // discard return value
              // balance checks: root, caller, callee
              .push(rootAccount.getAddress())
              .op(OpCode.BALANCE)
              .op(POP)
              .push(callerAccount.getAddress())
              .op(OpCode.BALANCE)
              .op(POP)
              .push(calleeAccount.getAddress())
              .op(OpCode.BALANCE)
              .op(POP)
              // optional revert
              .push(0)
              .push(0)
              .op(par.revertType == RevertType.TERMINATES_ON_REVERT ? REVERT : STOP);

  /**
   * Caller program (INFINITE_LOOP variant): increments a storage counter on every invocation and
   * makes a CALL to {@code targetAccount}.
   */
  BiFunction<ToyAccount, RevertType, BytecodeCompiler> callProgram =
      (targetAccount, revertType) ->
          BytecodeCompiler.newProgram(chainConfig)
              .push(0)
              .op(SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(SSTORE)
              .apply(
                  program ->
                      appendFullGasCall(
                          program, CALL, targetAccount.getAddress(), 0, 0, 0, 0, 0))
              .push(0)
              .push(0)
              .op(revertType == RevertType.TERMINATES_ON_REVERT ? REVERT : STOP);

  /**
   * Caller program (EXIT_EARLY variant): checks a loop-depth counter and exits early when the
   * counter reaches 3; otherwise increments the counter and makes a CALL to {@code targetAccount}.
   *
   * <p>JUMPDEST is at PC = 10 (header: 2+1+2+1+2+1+1 = 10 bytes before JUMPDEST).
   */
  BiFunction<ToyAccount, RevertType, BytecodeCompiler> conditionalCallProgram =
      (targetAccount, revertType) ->
          BytecodeCompiler.newProgram(chainConfig)
              .push(0)
              .op(SLOAD) // LOOP_DEPTH_CURRENT
              .push(3) // LOOP_DEPTH_MAX
              .op(GT) // MAX > CURRENT
              .push(10) // jump destination (JUMPDEST at PC 10)
              .op(JUMPI) // conditional jump
              .op(STOP) // early exit
              .op(JUMPDEST) // PC 10
              .push(0)
              .op(SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(SSTORE)
              .apply(
                  program ->
                      appendFullGasCall(
                          program, CALL, targetAccount.getAddress(), 0, 0, 0, 0, 0))
              .push(0)
              .push(0)
              .op(revertType == RevertType.TERMINATES_ON_REVERT ? REVERT : STOP);

  // ── Scenario-specific caller program builders ─────────────────────────────

  /**
   * Caller program for the CALL_ABORT (balance) scenario. The caller attempts to transfer
   * (balance + 1) to the callee, which causes the call to abort due to insufficient balance.
   */
  BiFunction<ToyAccount, RevertType, BytecodeCompiler> abortBalanceCallerProgram =
      (targetAccount, revertType) ->
          BytecodeCompiler.newProgram(chainConfig)
              .push(0)
              .op(SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(SSTORE)
              .apply(
                  program ->
                      appendInsufficientBalanceCall(
                          program, CALL, 1000, targetAccount.getAddress(), 0, 0, 0, 0))
              .op(POP) // discard the 0 return value from the aborted call
              .push(0)
              .push(0)
              .op(revertType == RevertType.TERMINATES_ON_REVERT ? REVERT : STOP);

  /**
   * Caller program for the EXIT_EARLY variant of the CALL_ABORT (balance) scenario.
   *
   * <p>JUMPDEST is at PC = 10.
   */
  BiFunction<ToyAccount, RevertType, BytecodeCompiler> conditionalAbortBalanceCallerProgram =
      (targetAccount, revertType) ->
          BytecodeCompiler.newProgram(chainConfig)
              .push(0)
              .op(SLOAD) // LOOP_DEPTH_CURRENT
              .push(3) // LOOP_DEPTH_MAX
              .op(GT)
              .push(10)
              .op(JUMPI)
              .op(STOP)
              .op(JUMPDEST)
              .push(0)
              .op(SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(SSTORE)
              .apply(
                  program ->
                      appendInsufficientBalanceCall(
                          program, CALL, 1000, targetAccount.getAddress(), 0, 0, 0, 0))
              .op(POP)
              .push(0)
              .push(0)
              .op(revertType == RevertType.TERMINATES_ON_REVERT ? REVERT : STOP);

  /**
   * Caller program for the CALL_EOA scenario. The caller transfers some value to the callee. When
   * the callee has no code (pure EOA) this triggers {@code CALL_EOA_SUCCESS_WILL_REVERT} or {@code
   * CALL_EOA_SUCCESS_WONT_REVERT}. When the callee is delegated or has code the ZK tracer is
   * exercised with a non-zero value call to a delegated account.
   */
  BiFunction<ToyAccount, RevertType, BytecodeCompiler> eoaCallerProgram =
      (targetAccount, revertType) ->
          BytecodeCompiler.newProgram(chainConfig)
              .push(0)
              .op(SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(SSTORE)
              .apply(
                  program ->
                      appendCall(
                          program,
                          CALL,
                          ROOT_CALL_GAS,
                          targetAccount.getAddress(),
                          13, // non-zero value (EOA transfer)
                          0,
                          0,
                          0,
                          0))
              .op(POP)
              .push(0)
              .push(0)
              .op(revertType == RevertType.TERMINATES_ON_REVERT ? REVERT : STOP);

  /**
   * Caller program (EXIT_EARLY) for the CALL_EOA scenario.
   *
   * <p>JUMPDEST is at PC = 10.
   */
  BiFunction<ToyAccount, RevertType, BytecodeCompiler> conditionalEoaCallerProgram =
      (targetAccount, revertType) ->
          BytecodeCompiler.newProgram(chainConfig)
              .push(0)
              .op(SLOAD)
              .push(3)
              .op(GT)
              .push(10)
              .op(JUMPI)
              .op(STOP)
              .op(JUMPDEST)
              .push(0)
              .op(SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(SSTORE)
              .apply(
                  program ->
                      appendCall(
                          program,
                          CALL,
                          ROOT_CALL_GAS,
                          targetAccount.getAddress(),
                          13,
                          0,
                          0,
                          0,
                          0))
              .op(POP)
              .push(0)
              .push(0)
              .op(revertType == RevertType.TERMINATES_ON_REVERT ? REVERT : STOP);

  /**
   * Caller program for the CALL_SMC scenario. The caller makes a zero-value call to the callee
   * (which must have code). The callee's code determines whether this results in a success or
   * failure scenario.
   */
  BiFunction<ToyAccount, RevertType, BytecodeCompiler> smcCallerProgram =
      (targetAccount, revertType) ->
          BytecodeCompiler.newProgram(chainConfig)
              .push(0)
              .op(SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(SSTORE)
              .apply(
                  program ->
                      appendFullGasCall(
                          program, CALL, targetAccount.getAddress(), 0, 0, 0, 0, 0))
              .op(POP)
              .push(0)
              .push(0)
              .op(revertType == RevertType.TERMINATES_ON_REVERT ? REVERT : STOP);

  /**
   * Caller program (EXIT_EARLY) for the CALL_SMC scenario.
   *
   * <p>JUMPDEST is at PC = 10.
   */
  BiFunction<ToyAccount, RevertType, BytecodeCompiler> conditionalSmcCallerProgram =
      (targetAccount, revertType) ->
          BytecodeCompiler.newProgram(chainConfig)
              .push(0)
              .op(SLOAD)
              .push(3)
              .op(GT)
              .push(10)
              .op(JUMPI)
              .op(STOP)
              .op(JUMPDEST)
              .push(0)
              .op(SLOAD)
              .push(1)
              .op(OpCode.ADD)
              .push(0)
              .op(SSTORE)
              .apply(
                  program ->
                      appendFullGasCall(
                          program, CALL, targetAccount.getAddress(), 0, 0, 0, 0, 0))
              .op(POP)
              .push(0)
              .push(0)
              .op(revertType == RevertType.TERMINATES_ON_REVERT ? REVERT : STOP);

  /**
   * Caller program for the CALL_EXCEPTION (staticx) scenario. When the root uses STATICCALL the
   * caller runs in a static context. Attempting a CALL with a non-zero value inside a static
   * context triggers a STATICX exception at the CALL opcode, which the hub classifies as
   * CALL_EXCEPTION.
   *
   * <p>No SSTORE is used here because SSTORE would itself trigger STATICX before reaching the CALL
   * instruction.
   */
  Function<ToyAccount, BytecodeCompiler> exceptionStaticxCallerProgram =
      targetAccount ->
          BytecodeCompiler.newProgram(chainConfig)
              // CALL with value = 1 inside static context → STATICX at this CALL opcode
              .push(0) // RAC
              .push(0) // RAO
              .push(0) // CDS
              .push(0) // CDO
              .push(1) // value = 1 (non-zero → triggers STATICX in static context)
              .push(targetAccount.getAddress())
              .op(GAS)
              .op(CALL);

  // ── Helpers ───────────────────────────────────────────────────────────────

  /**
   * Sets up the callee account code based on {@code calleeType}. For SMC / DELEGATED_TO_SMC the
   * callee runs the provided {@code calleeProgram}. For DELEGATED_TO_ROOT / DELEGATED_TO_CALLER the
   * callee simply delegates to those already-configured accounts.
   */
  private void setupCallee(
      CalleeType calleeType,
      BiFunction<ToyAccount, RevertType, BytecodeCompiler> calleeProgram,
      RevertType calleeCodeRevertType) {
    switch (calleeType) {
      case DELEGATED_TO_ROOT -> calleeAccount.delegateTo(rootAccount);
      case DELEGATED_TO_CALLER -> calleeAccount.delegateTo(callerAccount);
      case DELEGATED_TO_SMC -> {
        calleeAccount.delegateTo(smcAccount2);
        smcAccount2.setCode(calleeProgram.apply(callerAccount, calleeCodeRevertType).compile());
      }
      case SMC -> calleeAccount.setCode(
          calleeProgram.apply(callerAccount, calleeCodeRevertType).compile());
    }
  }

  /** Simple callee code used in SMC tests: just stops or reverts. */
  private BytecodeCompiler simpleCalleeCode(RevertType revertType) {
    return BytecodeCompiler.newProgram(chainConfig)
        .push(0)
        .push(0)
        .op(revertType == RevertType.TERMINATES_ON_REVERT ? REVERT : STOP);
  }

  /** Builds and executes the transaction. */
  private void runTest(TestInfo testInfo) {
    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(rootAccount)
            .keyPair(senderKeyPair)
            .gasLimit(1_000_000L)
            .build();

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
        .build()
        .run();
  }

  // ── Tests ─────────────────────────────────────────────────────────────────

  /**
   * Generic CALL delegation test: root calls caller (which calls callee); both caller and callee
   * may be direct SMCs or EIP-7702 delegated accounts. Supports optional looping via
   * LoopType-controlled programs.
   */
  @ParameterizedTest
  @MethodSource("fullDelegationTestSource")
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

    rootAccount.setCode(
        rootProgram.apply(new RootProgramParams(CALL, rootCodeRevertType)).compile());

    switch (callerType) {
      case DELEGATED -> {
        callerAccount.delegateTo(smcAccount1);
        smcAccount1.setCode(actualCallProgram.apply(calleeAccount, callerCodeRevertType).compile());
      }
      case SMC -> callerAccount.setCode(
          actualCallProgram.apply(calleeAccount, callerCodeRevertType).compile());
    }

    setupCallee(calleeType, actualCallProgram, calleeCodeRevertType);
    runTest(testInfo);
  }

  /**
   * CALL_ABORT (balance deficit) delegation test: the caller attempts to transfer more than its
   * balance to the callee, triggering a CALL_ABORT scenario in the hub.
   */
  @ParameterizedTest
  @MethodSource("noCalleeRevertDelegationTestSource")
  public void callAbortBalanceDelegationTest(
      CallerType callerType,
      CalleeType calleeType,
      RevertType rootCodeRevertType,
      RevertType callerCodeRevertType,
      LoopType loopType,
      TestInfo testInfo) {

    BiFunction<ToyAccount, RevertType, BytecodeCompiler> actualCallerProgram =
        loopType == LoopType.EXIT_EARLY
            ? conditionalAbortBalanceCallerProgram
            : abortBalanceCallerProgram;

    rootAccount.setCode(
        rootProgram.apply(new RootProgramParams(CALL, rootCodeRevertType)).compile());

    switch (callerType) {
      case DELEGATED -> {
        callerAccount.delegateTo(smcAccount1);
        smcAccount1.setCode(
            actualCallerProgram.apply(calleeAccount, callerCodeRevertType).compile());
      }
      case SMC -> callerAccount.setCode(
          actualCallerProgram.apply(calleeAccount, callerCodeRevertType).compile());
    }

    // Callee code doesn't affect the outcome of an aborted call, but set it up for completeness
    switch (calleeType) {
      case DELEGATED_TO_ROOT -> calleeAccount.delegateTo(rootAccount);
      case DELEGATED_TO_CALLER -> calleeAccount.delegateTo(callerAccount);
      case DELEGATED_TO_SMC -> calleeAccount.delegateTo(smcAccount2); // smcAccount2 has no code
      case SMC -> {} // calleeAccount has no code (empty account — acts as EOA)
    }

    runTest(testInfo);
  }

  /**
   * CALL_EOA delegation test: the caller makes a value-transfer call to the callee. When the
   * callee has no code this reproduces the CALL_EOA_SUCCESS_{WILL,WONT}_REVERT scenarios. When the
   * callee is delegated the test exercises EIP-7702 interactions with value-bearing calls.
   */
  @ParameterizedTest
  @MethodSource("noCalleeRevertDelegationTestSource")
  public void callEoaDelegationTest(
      CallerType callerType,
      CalleeType calleeType,
      RevertType rootCodeRevertType,
      RevertType callerCodeRevertType,
      LoopType loopType,
      TestInfo testInfo) {

    BiFunction<ToyAccount, RevertType, BytecodeCompiler> actualCallerProgram =
        loopType == LoopType.EXIT_EARLY ? conditionalEoaCallerProgram : eoaCallerProgram;

    rootAccount.setCode(
        rootProgram.apply(new RootProgramParams(CALL, rootCodeRevertType)).compile());

    switch (callerType) {
      case DELEGATED -> {
        callerAccount.delegateTo(smcAccount1);
        smcAccount1.setCode(
            actualCallerProgram.apply(calleeAccount, callerCodeRevertType).compile());
      }
      case SMC -> callerAccount.setCode(
          actualCallerProgram.apply(calleeAccount, callerCodeRevertType).compile());
    }

    // For EOA scenario the callee normally has no code; with delegation it has delegation code.
    switch (calleeType) {
      case DELEGATED_TO_ROOT -> calleeAccount.delegateTo(rootAccount);
      case DELEGATED_TO_CALLER -> calleeAccount.delegateTo(callerAccount);
      case DELEGATED_TO_SMC -> calleeAccount.delegateTo(smcAccount2); // smcAccount2 has no code
      case SMC -> {} // calleeAccount has no code (pure EOA)
    }

    runTest(testInfo);
  }

  /**
   * CALL_SMC delegation test: the caller makes a zero-value call to the callee, which must have
   * code. The callee code either stops or reverts based on {@code calleeCodeRevertType}, producing
   * the CALL_SMC_{SUCCESS,FAILURE}_{WILL,WONT}_REVERT scenarios.
   */
  @ParameterizedTest
  @MethodSource("fullDelegationTestSource")
  public void callSmcDelegationTest(
      CallerType callerType,
      CalleeType calleeType,
      RevertType rootCodeRevertType,
      RevertType callerCodeRevertType,
      RevertType calleeCodeRevertType,
      LoopType loopType,
      TestInfo testInfo) {

    BiFunction<ToyAccount, RevertType, BytecodeCompiler> actualCallerProgram =
        loopType == LoopType.EXIT_EARLY ? conditionalSmcCallerProgram : smcCallerProgram;

    rootAccount.setCode(
        rootProgram.apply(new RootProgramParams(CALL, rootCodeRevertType)).compile());

    switch (callerType) {
      case DELEGATED -> {
        callerAccount.delegateTo(smcAccount1);
        smcAccount1.setCode(
            actualCallerProgram.apply(calleeAccount, callerCodeRevertType).compile());
      }
      case SMC -> callerAccount.setCode(
          actualCallerProgram.apply(calleeAccount, callerCodeRevertType).compile());
    }

    // Callee runs simple code (stop or revert); DELEGATED_TO_ROOT/CALLER creates circularity.
    switch (calleeType) {
      case DELEGATED_TO_ROOT -> calleeAccount.delegateTo(rootAccount);
      case DELEGATED_TO_CALLER -> calleeAccount.delegateTo(callerAccount);
      case DELEGATED_TO_SMC -> {
        calleeAccount.delegateTo(smcAccount2);
        smcAccount2.setCode(simpleCalleeCode(calleeCodeRevertType).compile());
      }
      case SMC -> calleeAccount.setCode(simpleCalleeCode(calleeCodeRevertType).compile());
    }

    runTest(testInfo);
  }

  /**
   * CALL_EXCEPTION (staticx) delegation test: the root uses STATICCALL to invoke the caller. The
   * caller then attempts a CALL with a non-zero value, which triggers a static-context violation
   * (STATICX) at the CALL opcode. The hub classifies this as CALL_EXCEPTION.
   *
   * <p>No loopType parameter here: the caller program must not contain SSTORE (which would itself
   * trigger STATICX before reaching the CALL instruction).
   */
  @ParameterizedTest
  @MethodSource("exceptionDelegationTestSource")
  public void callExceptionStaticxDelegationTest(
      CallerType callerType,
      CalleeType calleeType,
      RevertType rootCodeRevertType,
      TestInfo testInfo) {

    rootAccount.setCode(
        rootProgram.apply(new RootProgramParams(STATICCALL, rootCodeRevertType)).compile());

    switch (callerType) {
      case DELEGATED -> {
        callerAccount.delegateTo(smcAccount1);
        smcAccount1.setCode(exceptionStaticxCallerProgram.apply(calleeAccount).compile());
      }
      case SMC -> callerAccount.setCode(
          exceptionStaticxCallerProgram.apply(calleeAccount).compile());
    }

    // Callee doesn't run (exception happens before it is reached), but set up for completeness
    switch (calleeType) {
      case DELEGATED_TO_ROOT -> calleeAccount.delegateTo(rootAccount);
      case DELEGATED_TO_CALLER -> calleeAccount.delegateTo(callerAccount);
      case DELEGATED_TO_SMC -> calleeAccount.delegateTo(smcAccount2);
      case SMC -> {} // calleeAccount has no code
    }

    runTest(testInfo);
  }

  // ── Test-source methods ───────────────────────────────────────────────────

  /** Provides all combinations of parameters including a {@code calleeCodeRevertType}. */
  static Stream<Arguments> fullDelegationTestSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (CallerType callerType : CallerType.values()) {
      for (CalleeType calleeType : CalleeType.values()) {
        for (RevertType rootRevert : RevertType.values()) {
          for (RevertType callerRevert : RevertType.values()) {
            for (RevertType calleeRevert : RevertType.values()) {
              for (LoopType loopType : LoopType.values()) {
                arguments.add(
                    Arguments.of(
                        callerType,
                        calleeType,
                        rootRevert,
                        callerRevert,
                        calleeRevert,
                        loopType));
              }
            }
          }
        }
      }
    }
    return arguments.stream();
  }

  /**
   * Provides combinations of parameters for scenarios where the callee code is not relevant (e.g.
   * CALL_ABORT, CALL_EOA).
   */
  static Stream<Arguments> noCalleeRevertDelegationTestSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (CallerType callerType : CallerType.values()) {
      for (CalleeType calleeType : CalleeType.values()) {
        for (RevertType rootRevert : RevertType.values()) {
          for (RevertType callerRevert : RevertType.values()) {
            for (LoopType loopType : LoopType.values()) {
              arguments.add(
                  Arguments.of(callerType, calleeType, rootRevert, callerRevert, loopType));
            }
          }
        }
      }
    }
    return arguments.stream();
  }

  /**
   * Provides combinations for the CALL_EXCEPTION (staticx) scenario where only root and caller
   * reversal is relevant.
   */
  static Stream<Arguments> exceptionDelegationTestSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (CallerType callerType : CallerType.values()) {
      for (CalleeType calleeType : CalleeType.values()) {
        for (RevertType rootRevert : RevertType.values()) {
          arguments.add(Arguments.of(callerType, calleeType, rootRevert));
        }
      }
    }
    return arguments.stream();
  }
}
