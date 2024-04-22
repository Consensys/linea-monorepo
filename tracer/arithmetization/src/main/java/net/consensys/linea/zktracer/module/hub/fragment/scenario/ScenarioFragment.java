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

package net.consensys.linea.zktracer.module.hub.fragment.scenario;

import java.util.Optional;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.precompiles.PrecompileInvocation;
import net.consensys.linea.zktracer.types.MemorySpan;
import net.consensys.linea.zktracer.types.Precompile;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.worldstate.WorldView;

/** This machine generates lines */
@RequiredArgsConstructor
public class ScenarioFragment implements TraceFragment, PostTransactionDefer {
  private enum CallType {
    /** describes a normal call */
    CALL,
    /** describes the second scenario line required by a call to a precompile */
    PRECOMPILE,
    /** describes a call into initcode */
    CREATE,
    /** describes a RETURN from initcode */
    CODE_DEPOSIT;

    boolean isCall() {
      return this == CALL;
    }

    boolean isPrecompile() {
      return this == PRECOMPILE;
    }

    boolean isCreate() {
      return this == CREATE;
    }

    boolean isDeposit() {
      return this == CODE_DEPOSIT;
    }
  }

  private final Optional<PrecompileInvocation> precompileCall;
  private final CallType type;

  /**
   * Is set if: - this is a CALL to an EOA or a precompile - this is a CREATE with an empty initcode
   */
  private final boolean targetHasCode;

  private final int callerId;
  private final int calleeId;
  private final boolean raisedException;
  private final boolean hasAborted;
  private final boolean hasFailed;
  private final boolean raisedInvalidCodePrefix;

  private boolean callerReverts = false;

  MemorySpan callDataSegment;
  MemorySpan requestedReturnDataSegment;

  /**
   * Is set if: - this is a CALL and the callee reverts - this is a CREATE and the creation failed -
   * this is a PRECOMPILE and the call is invalid (wrong arguments, lack of gas, ...)
   */
  private boolean childContextFails = false;

  public static ScenarioFragment forCall(final Hub hub, boolean targetHasCode) {
    ScenarioFragment r =
        new ScenarioFragment(
            Optional.empty(),
            CallType.CALL,
            targetHasCode,
            hub.currentFrame().id(),
            hub.callStack().futureId(),
            hub.pch().exceptions().any(),
            hub.pch().aborts().any(),
            hub.pch().failures().any(),
            hub.pch().exceptions().invalidCodePrefix());
    hub.defers().postTx(r);
    return r;
  }

  public static ScenarioFragment forCreate(final Hub hub, boolean targetHasCode) {
    return new ScenarioFragment(
        Optional.empty(),
        CallType.CREATE,
        targetHasCode,
        hub.currentFrame().id(),
        hub.callStack().futureId(),
        hub.pch().exceptions().any(),
        hub.pch().aborts().any(),
        hub.pch().failures().any(),
        hub.pch().exceptions().invalidCodePrefix());
  }

  public static ScenarioFragment forSmartContractCallSection(
      final Hub hub, int calledFrameId, int callerFrameId) {
    final ScenarioFragment r =
        new ScenarioFragment(
            Optional.empty(),
            CallType.CALL,
            true,
            callerFrameId,
            calledFrameId,
            false,
            false,
            false,
            false);
    r.callDataSegment = hub.transients().op().callDataSegment();
    r.requestedReturnDataSegment = hub.transients().op().returnDataRequestedSegment();
    return r;
  }

  public static ScenarioFragment forNoCodeCallSection(
      final Hub hub, Optional<PrecompileInvocation> precompileCall, int callerId, int calleeId) {
    ScenarioFragment r;
    if (precompileCall.isPresent()) {
      r =
          new ScenarioFragment(
              precompileCall, CallType.CALL, false, callerId, calleeId, false, false, false, false);
    } else {
      r =
          new ScenarioFragment(
              Optional.empty(),
              CallType.CALL,
              false,
              callerId,
              calleeId,
              false,
              false,
              false,
              false);
    }
    r.callDataSegment = hub.transients().op().callDataSegment();
    r.requestedReturnDataSegment = hub.transients().op().returnDataRequestedSegment();

    r.fillPostCallInformation(hub);
    return r;
  }

  public static ScenarioFragment forPrecompileEpilogue(
      final Hub hub, PrecompileInvocation precompile, int callerId, int calleeId) {
    final ScenarioFragment r =
        new ScenarioFragment(
            Optional.of(precompile),
            CallType.PRECOMPILE,
            false,
            callerId,
            calleeId,
            false,
            false,
            false,
            false);
    // This one is already created from a post-tx hook
    r.callDataSegment = precompile.callDataSource();
    r.requestedReturnDataSegment = precompile.requestedReturnDataTarget();
    r.fillPostCallInformation(hub);
    return r;
  }

  private boolean calleeSelfReverts() {
    return this.childContextFails;
  }

  private boolean creationFailed() {
    return this.childContextFails;
  }

  private boolean successfulPrecompileCall() {
    return !this.childContextFails;
  }

  private boolean targetIsPrecompile() {
    return this.precompileCall.isPresent();
  }

  /**
   * Fill the information related to the CALL this fragment stems from. This may be done either
   * through a defer if the fragment is created at runtime, or directly if the fragment is already
   * created within a post-transaction defer.
   *
   * @param hub the execution context
   */
  private void fillPreCallInformation(final Hub hub) {
    this.callDataSegment = hub.callStack().getById(calleeId).callDataInfo().memorySpan();
    this.requestedReturnDataSegment = hub.callStack().getById(calleeId).requestedReturnDataTarget();
  }

  /**
   * Fill the information related to the CALL this fragment stems from. This may be done either
   * through a defer if the fragment is created at runtime, or directly if the fragment is already
   * created within a post-transaction defer.
   *
   * @param hub the execution context
   */
  private void fillPostCallInformation(final Hub hub) {
    // It does not make sense to interrogate a child frame that was never created
    if (!this.hasFailed && !this.hasAborted && !this.raisedException && targetHasCode) {
      // TODO: can a context without code reverts? @Olivier
      this.childContextFails = hub.callStack().getById(calleeId).hasReverted();
    }
    this.callerReverts = hub.callStack().getById(callerId).hasReverted();
  }

  @Override
  public void runPostTx(Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    this.fillPostCallInformation(hub);
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .peekAtScenario(true)
        .pScenarioCallException(type.isCall() && raisedException)
        .pScenarioCallAbort(type.isCall() && hasAborted)
        .pScenarioCallPrcFailure(
            type.isCall()
                && !hasAborted
                && targetIsPrecompile()
                && !callerReverts
                && this.calleeSelfReverts())
        .pScenarioCallPrcSuccessCallerWillRevert(
            type.isCall()
                && !hasAborted
                && targetIsPrecompile()
                && callerReverts
                && !this.calleeSelfReverts())
        .pScenarioCallPrcSuccessCallerWontRevert(
            type.isCall()
                && !hasAborted
                && targetIsPrecompile()
                && !callerReverts
                && !this.calleeSelfReverts())
        .pScenarioCallSmcFailureCallerWillRevert(
            type.isCall()
                && !hasAborted
                && targetHasCode
                && callerReverts
                && this.calleeSelfReverts())
        .pScenarioCallSmcFailureCallerWontRevert(
            type.isCall()
                && !hasAborted
                && targetHasCode
                && !callerReverts
                && this.calleeSelfReverts())
        .pScenarioCallSmcSuccessCallerWontRevert(
            type.isCall()
                && !hasAborted
                && targetHasCode
                && !callerReverts
                && !this.calleeSelfReverts())
        .pScenarioCallSmcSuccessCallerWillRevert(
            type.isCall()
                && !hasAborted
                && targetHasCode
                && callerReverts
                && !this.calleeSelfReverts())
        .pScenarioCallEoaSuccessCallerWontRevert(
            type.isCall()
                && !hasAborted
                && !targetIsPrecompile()
                && !targetHasCode
                && !callerReverts)
        .pScenarioCallEoaSuccessCallerWillRevert(
            type.isCall()
                && !hasAborted
                && !targetIsPrecompile()
                && !targetHasCode
                && callerReverts)
        .pScenarioCreateException(type.isCreate() && raisedException)
        .pScenarioCreateAbort(type.isCreate() && hasAborted)
        .pScenarioCreateFailureConditionWillRevert(type.isCreate() && hasFailed && callerReverts)
        .pScenarioCreateFailureConditionWontRevert(type.isCreate() && hasFailed && !callerReverts)
        .pScenarioCreateEmptyInitCodeWillRevert(type.isCreate() && !targetHasCode && callerReverts)
        .pScenarioCreateEmptyInitCodeWontRevert(type.isCreate() && !targetHasCode && !callerReverts)
        .pScenarioCreateNonemptyInitCodeFailureWillRevert(
            type.isCreate() && targetHasCode && creationFailed() && callerReverts)
        .pScenarioCreateNonemptyInitCodeFailureWontRevert(
            type.isCreate() && targetHasCode && creationFailed() && !callerReverts)
        .pScenarioCreateNonemptyInitCodeSuccessWillRevert(
            type.isCreate() && targetHasCode && !creationFailed() && callerReverts)
        .pScenarioCreateNonemptyInitCodeSuccessWontRevert(
            type.isCreate() && targetHasCode && !creationFailed() && !callerReverts)
        .pScenarioPrcEcrecover(
            precompileCall.map(x -> x.precompile().equals(Precompile.EC_RECOVER)).orElse(false))
        .pScenarioPrcSha2256(
            precompileCall.map(x -> x.precompile().equals(Precompile.SHA2_256)).orElse(false))
        .pScenarioPrcRipemd160(
            precompileCall.map(x -> x.precompile().equals(Precompile.RIPEMD_160)).orElse(false))
        .pScenarioPrcIdentity(
            precompileCall.map(x -> x.precompile().equals(Precompile.IDENTITY)).orElse(false))
        .pScenarioPrcModexp(
            precompileCall.map(x -> x.precompile().equals(Precompile.MODEXP)).orElse(false))
        .pScenarioPrcEcadd(
            precompileCall.map(x -> x.precompile().equals(Precompile.EC_ADD)).orElse(false))
        .pScenarioPrcEcmul(
            precompileCall.map(x -> x.precompile().equals(Precompile.EC_MUL)).orElse(false))
        .pScenarioPrcEcpairing(
            precompileCall.map(x -> x.precompile().equals(Precompile.EC_PAIRING)).orElse(false))
        .pScenarioPrcBlake2F(
            precompileCall.map(x -> x.precompile().equals(Precompile.BLAKE2F)).orElse(false))
        .pScenarioPrcSuccessWillRevert(
            type.isPrecompile() && successfulPrecompileCall() && callerReverts)
        .pScenarioPrcSuccessWontRevert(
            type.isPrecompile() && successfulPrecompileCall() && !callerReverts)
        .pScenarioPrcFailureKnownToHub(
            precompileCall.map(PrecompileInvocation::hubFailure).orElse(false))
        .pScenarioPrcFailureKnownToRam(
            precompileCall.map(PrecompileInvocation::ramFailure).orElse(false))
        .pScenarioPrcCallerGas(precompileCall.map(s -> s.gasAtCall() - s.opCodeGas()).orElse(0L))
        .pScenarioPrcCalleeGas(precompileCall.map(PrecompileInvocation::gasAllowance).orElse(0L))
        .pScenarioPrcReturnGas(
            precompileCall
                .filter(s -> successfulPrecompileCall())
                .map(s -> s.gasAllowance() - s.precompilePrice())
                .orElse(0L))
        .pScenarioPrcCdo(type.isPrecompile() ? callDataSegment.offset() : 0)
        .pScenarioPrcCds(type.isPrecompile() ? callDataSegment.length() : 0)
        .pScenarioPrcRao(type.isPrecompile() ? requestedReturnDataSegment.offset() : 0)
        .pScenarioPrcRac(type.isPrecompile() ? requestedReturnDataSegment.length() : 0)
    //        .pScenarioCodedeposit(type.isDeposit())
    //        .pScenarioCodedepositInvalidCodePrefix(type.isDeposit() && raisedInvalidCodePrefix)
    //        .pScenarioCodedepositValidCodePrefix(false) // TODO: @Olivier
    //        .pScenarioSelfdestruct(false); // TODO: @Olivier
    ;
  }
}
