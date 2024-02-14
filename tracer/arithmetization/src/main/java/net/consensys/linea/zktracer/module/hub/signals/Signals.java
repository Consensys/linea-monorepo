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

package net.consensys.linea.zktracer.module.hub.signals;

import static net.consensys.linea.zktracer.module.ec_data.EcData.EC_PRECOMPILES;

import java.util.Optional;
import java.util.Set;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.InstructionFamily;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

/**
 * Encodes the signals triggering other components.
 *
 * <p>When a component is requested, also checks that it may actually be triggered in the current
 * context.
 */
@Accessors(fluent = true)
@RequiredArgsConstructor
public class Signals {
  private final Set<InstructionFamily> AUTOMATIC_GAS_MODULE_TRIGGER =
      Set.of(
          InstructionFamily.CREATE,
          InstructionFamily.CALL,
          InstructionFamily.HALT,
          InstructionFamily.INVALID);

  @Getter private boolean add;
  @Getter private boolean bin;
  @Getter private boolean mul;
  @Getter private boolean ext;
  @Getter private boolean mod;
  @Getter private boolean wcp;
  @Getter private boolean shf;

  @Getter private boolean gas;
  @Getter private boolean mmu;
  @Getter private boolean mxp;
  @Getter private boolean oob;
  @Getter private boolean stp;
  @Getter private boolean exp;
  @Getter private boolean trm;
  @Getter private boolean hashInfo;
  @Getter private boolean logInfo;
  @Getter private boolean romLex;
  @Getter private boolean rlpAddr;
  @Getter private boolean ecData;

  private final PlatformController platformController;

  public void reset() {
    this.add = false;
    this.bin = false;
    this.mul = false;
    this.ext = false;
    this.mod = false;
    this.wcp = false;
    this.shf = false;

    this.gas = false;
    this.mmu = false;
    this.mxp = false;
    this.oob = false;
    this.stp = false;
    this.exp = false;
    this.trm = false;
    this.hashInfo = false;
    this.romLex = false;
    this.rlpAddr = false;
    this.ecData = false;
  }

  public Signals snapshot() {
    Signals r = new Signals(null);
    r.add = this.add;
    r.bin = this.bin;
    r.mul = this.mul;
    r.ext = this.ext;
    r.mod = this.mod;
    r.wcp = this.wcp;
    r.shf = this.shf;

    r.gas = this.gas;
    r.mmu = this.mmu;
    r.mxp = this.mxp;
    r.oob = this.oob;
    r.stp = this.stp;
    r.exp = this.exp;
    r.trm = this.trm;
    r.hashInfo = this.hashInfo;
    r.romLex = this.romLex;
    r.rlpAddr = this.rlpAddr;
    r.ecData = this.ecData;

    return r;
  }

  /**
   * Setup all the signalling required to trigger modules for the execution of the current
   * operation.
   *
   * @param frame the currently executing frame
   * @param platformController the parent controller
   * @param hub the execution context
   */
  public void prepare(MessageFrame frame, PlatformController platformController, Hub hub) {
    final OpCode opCode = hub.opCode();
    final Exceptions ex = platformController.exceptions();

    // this.gas coincides with CONTEXT_MAY_CHANGE
    this.gas =
        ex.any()
            || this.AUTOMATIC_GAS_MODULE_TRIGGER.contains(hub.opCodeData().instructionFamily());

    if (ex.stackException()) {
      return;
    }

    switch (opCode) {
      case CALLDATACOPY, CODECOPY -> {
        this.mxp = ex.outOfMemoryExpansion() || ex.outOfGas() || ex.none();
        this.mmu = ex.none() && !frame.getStackItem(1).isZero();
      }

      case RETURNDATACOPY -> {
        this.oob = ex.none() || ex.returnDataCopyFault();
        this.mxp = ex.none() || ex.outOfMemoryExpansion() || ex.outOfGas();
        this.mmu = ex.none() && !frame.getStackItem(1).isZero();
      }

      case EXTCODECOPY -> {
        boolean nonzeroSize = !frame.getStackItem(2).isZero();
        this.mxp = ex.outOfMemoryExpansion() || ex.outOfGas() || ex.none();
        this.trm = ex.outOfGas() || ex.none();
        this.mmu = ex.none() && nonzeroSize;

        Address address = Words.toAddress(frame.getStackItem(0));
        final boolean targetAddressHasCode =
            Optional.ofNullable(frame.getWorldUpdater().get(address))
                .map(AccountState::hasCode)
                .orElse(false);

        this.romLex = ex.none() && nonzeroSize && targetAddressHasCode;
      }

      case LOG0, LOG1, LOG2, LOG3, LOG4 -> {
        this.mxp = ex.outOfMemoryExpansion() || ex.outOfGas() || ex.none();
        this.mmu = ex.none() && !frame.getStackItem(1).isZero();
        // logInfo and logData are triggered via rlpRcpt at the end of the tx
      }

      case CALL, DELEGATECALL, STATICCALL, CALLCODE -> {
        this.mxp = !ex.staticFault();
        this.stp = ex.outOfGas() || ex.none();
        this.oob = opCode.equals(OpCode.CALL) && ex.staticFault() || ex.none();
        this.trm = ex.outOfGas() || ex.none();

        final boolean triggersAbortingCondition =
            ex.none() && this.platformController.aborts().any();

        final Address target = Words.toAddress(frame.getStackItem(1));
        final boolean targetAddressHasNonEmptyCode =
            Optional.ofNullable(frame.getWorldUpdater().get(target))
                .map(AccountState::hasCode)
                .orElse(false);

        this.romLex = ex.none() && !triggersAbortingCondition && targetAddressHasNonEmptyCode;

        this.ecData = ex.none() && EC_PRECOMPILES.contains(target);
      }

      case CREATE, CREATE2 -> {
        boolean triggersAbortingCondition = ex.none() && this.platformController.aborts().any();

        boolean triggersFailureCondition = false;
        if (ex.none() && this.platformController.aborts().none()) {
          triggersFailureCondition = this.platformController.failures().any();
        }

        final boolean nonzeroSize = !frame.getStackItem(2).isZero();
        final boolean isCreate2 = opCode == OpCode.CREATE2;

        this.mxp = !ex.staticFault();
        this.stp = ex.outOfGas() || ex.none();
        this.oob = ex.none();
        this.rlpAddr = ex.none() && !triggersAbortingCondition;
        this.hashInfo = ex.none() && !triggersAbortingCondition && nonzeroSize && isCreate2;
        this.romLex =
            ex.none() && !triggersAbortingCondition && nonzeroSize && !triggersFailureCondition;
        this.mmu = this.hashInfo || this.romLex;
      }

      case REVERT -> {
        this.mxp = ex.outOfMemoryExpansion() || ex.outOfGas() || ex.none();
        this.mmu =
            ex.none()
                && !frame.getStackItem(1).isZero()
                && hub.currentFrame().requestedReturnDataTarget().length() > 0;
      }

      case RETURN -> {
        final boolean isDeployment = frame.getType() == MessageFrame.Type.CONTRACT_CREATION;
        final boolean sizeNonZero = !frame.getStackItem(1).isZero();

        // WARN: Static part, other modules may be dynamically requested in the hub
        this.mxp =
            ex.outOfMemoryExpansion() || ex.outOfGas() || ex.invalidCodePrefix() || ex.none();
        this.oob = isDeployment && (ex.codeSizeOverflow() || ex.none());
        this.mmu =
            (isDeployment && ex.invalidCodePrefix())
                || (isDeployment && ex.none() && sizeNonZero)
                || (!isDeployment
                    && ex.none()
                    && sizeNonZero
                    && hub.currentFrame().requestedReturnDataTarget().length() > 0);
        this.romLex = this.hashInfo = isDeployment && ex.none() && sizeNonZero;
      }

      case EXP -> {
        this.exp = true;
        this.mul = !ex.outOfGas();
      }

        // other opcodes
      case ADD, SUB -> this.add = !ex.outOfGas();
      case MUL -> this.mul = !ex.outOfGas();
      case DIV, SDIV, MOD, SMOD -> this.mod = !ex.outOfGas();
      case ADDMOD, MULMOD -> this.ext = !ex.outOfGas();
      case LT, GT, SLT, SGT, EQ, ISZERO -> this.wcp = !ex.outOfGas();
      case AND, OR, XOR, NOT, SIGNEXTEND, BYTE -> this.bin = !ex.outOfGas();
      case SHL, SHR, SAR -> this.shf = !ex.outOfGas();
      case SHA3 -> {
        this.mxp = true;
        this.hashInfo = ex.none() && !frame.getStackItem(0).isZero();
        this.mmu = this.hashInfo;
      }
      case BALANCE, EXTCODESIZE, EXTCODEHASH, SELFDESTRUCT -> this.trm = true;
      case MLOAD, MSTORE, MSTORE8 -> {
        this.mxp = true;
        this.mmu = !ex.any();
      }
      case CALLDATALOAD -> {
        this.oob = true;
        this.mmu = frame.getInputData().size() > Words.clampedToLong(frame.getStackItem(0));
      }
      case SLOAD -> {}
      case SSTORE, JUMP, JUMPI -> this.oob = true;
      case MSIZE -> this.mxp = ex.none();
    }
  }
}
