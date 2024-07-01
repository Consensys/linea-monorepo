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

import static net.consensys.linea.zktracer.module.ecdata.EcData.EC_PRECOMPILES;

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
  @Getter private boolean blockhash;
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
  @Getter private boolean romLex;
  @Getter private boolean rlpAddr;
  @Getter private boolean ecData;

  private final PlatformController platformController;

  public void reset() {
    this.add = false;
    this.blockhash = false;
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
    r.blockhash = this.blockhash;
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
    final short ex = platformController.exceptions();

    // this.gas coincides with CONTEXT_MAY_CHANGE
    this.gas =
        Exceptions.any(ex)
            || this.AUTOMATIC_GAS_MODULE_TRIGGER.contains(hub.opCodeData().instructionFamily());

    if (Exceptions.stackException(ex)) {
      return;
    }

    switch (opCode) {
      case CALLDATACOPY, CODECOPY -> {
        this.mxp =
            Exceptions.outOfMemoryExpansion(ex) || Exceptions.outOfGas(ex) || Exceptions.none(ex);
        this.mmu = Exceptions.none(ex) && !frame.getStackItem(2).isZero();
      }

      case RETURNDATACOPY -> {
        this.oob = Exceptions.none(ex) || Exceptions.returnDataCopyFault(ex);
        this.mxp =
            Exceptions.none(ex) || Exceptions.outOfMemoryExpansion(ex) || Exceptions.outOfGas(ex);
        this.mmu = Exceptions.none(ex) && !frame.getStackItem(2).isZero();
      }

      case EXTCODECOPY -> {
        final boolean nonzeroSize = !frame.getStackItem(3).isZero();
        this.mxp =
            Exceptions.outOfMemoryExpansion(ex) || Exceptions.outOfGas(ex) || Exceptions.none(ex);
        this.trm = Exceptions.outOfGas(ex) || Exceptions.none(ex);
        this.mmu = Exceptions.none(ex) && nonzeroSize;

        final Address address = Words.toAddress(frame.getStackItem(0));
        final boolean targetAddressHasCode =
            Optional.ofNullable(frame.getWorldUpdater().get(address))
                .map(AccountState::hasCode)
                .orElse(false);

        this.romLex = Exceptions.none(ex) && nonzeroSize && targetAddressHasCode;
      }

      case LOG0, LOG1, LOG2, LOG3, LOG4 -> {
        this.mxp =
            Exceptions.outOfMemoryExpansion(ex) || Exceptions.outOfGas(ex) || Exceptions.none(ex);
        this.mmu =
            Exceptions.none(ex)
                && !frame
                    .getStackItem(1)
                    .isZero(); // TODO do not trigger the MMU if the context is going to revert and
        // check the HUB does increment or not the MMU stamp for reverted LOG
        // logInfo and logData are triggered via rlpRcpt at the end of the tx
      }

      case CALL, DELEGATECALL, STATICCALL, CALLCODE -> {
        this.mxp = !Exceptions.staticFault(ex);
        this.stp = Exceptions.outOfGas(ex) || Exceptions.none(ex);
        this.oob = opCode.equals(OpCode.CALL) && Exceptions.staticFault(ex) || Exceptions.none(ex);
        this.trm = Exceptions.outOfGas(ex) || Exceptions.none(ex);

        final boolean triggersAbortingCondition =
            Exceptions.none(ex) && this.platformController.aborts().any();

        final Address target = Words.toAddress(frame.getStackItem(1));
        final boolean targetAddressHasNonEmptyCode =
            Optional.ofNullable(frame.getWorldUpdater().get(target))
                .map(AccountState::hasCode)
                .orElse(false);

        this.romLex =
            Exceptions.none(ex) && !triggersAbortingCondition && targetAddressHasNonEmptyCode;
        this.ecData = Exceptions.none(ex) && EC_PRECOMPILES.contains(target);
        this.exp =
            Exceptions.none(ex)
                && this.platformController.aborts().none()
                && target.equals(Address.MODEXP);
      }

      case CREATE, CREATE2 -> {
        boolean triggersAbortingCondition =
            Exceptions.none(ex) && this.platformController.aborts().any();

        boolean triggersFailureCondition = false;
        if (Exceptions.none(ex) && this.platformController.aborts().none()) {
          triggersFailureCondition = this.platformController.failures().any();
        }

        final boolean nonzeroSize = !frame.getStackItem(2).isZero();
        final boolean isCreate2 = opCode == OpCode.CREATE2;

        this.mxp = !Exceptions.staticFault(ex);
        this.stp = Exceptions.outOfGas(ex) || Exceptions.none(ex);
        this.oob = Exceptions.none(ex);
        this.rlpAddr = Exceptions.none(ex) && !triggersAbortingCondition;
        this.hashInfo =
            Exceptions.none(ex) && !triggersAbortingCondition && nonzeroSize && isCreate2;
        this.romLex =
            Exceptions.none(ex)
                && !triggersAbortingCondition
                && nonzeroSize
                && !triggersFailureCondition;
        this.mmu = this.hashInfo || this.romLex;
      }

      case REVERT -> {
        this.mxp =
            Exceptions.outOfMemoryExpansion(ex) || Exceptions.outOfGas(ex) || Exceptions.none(ex);
        this.mmu =
            Exceptions.none(ex)
                && !frame.getStackItem(1).isZero()
                && !hub.currentFrame().requestedReturnDataTarget().isEmpty();
      }

      case RETURN -> {
        final boolean isDeployment = frame.getType() == MessageFrame.Type.CONTRACT_CREATION;
        final boolean sizeNonZero = !frame.getStackItem(1).isZero();

        // WARN: Static part, other modules may be dynamically requested in the hub
        this.mxp =
            Exceptions.outOfMemoryExpansion(ex)
                || Exceptions.outOfGas(ex)
                || Exceptions.invalidCodePrefix(ex)
                || Exceptions.none(ex);
        this.oob = isDeployment && (Exceptions.codeSizeOverflow(ex) || Exceptions.none(ex));
        this.mmu =
            (isDeployment && Exceptions.invalidCodePrefix(ex))
                || (isDeployment && Exceptions.none(ex) && sizeNonZero)
                || (!isDeployment
                    && Exceptions.none(ex)
                    && sizeNonZero
                    && !hub.currentFrame().requestedReturnDataTarget().isEmpty());
        this.romLex = this.hashInfo = isDeployment && Exceptions.none(ex) && sizeNonZero;
      }

      case EXP -> {
        this.exp = true;
        this.mul = !Exceptions.outOfGas(ex);
      }

        // other opcodes
      case ADD, SUB -> this.add = !Exceptions.outOfGas(ex);
      case MUL -> this.mul = !Exceptions.outOfGas(ex);
      case DIV, SDIV, MOD, SMOD -> this.mod = !Exceptions.outOfGas(ex);
      case ADDMOD, MULMOD -> this.ext = !Exceptions.outOfGas(ex);
      case LT, GT, SLT, SGT, EQ, ISZERO -> this.wcp = !Exceptions.outOfGas(ex);
      case AND, OR, XOR, NOT, SIGNEXTEND, BYTE -> this.bin = !Exceptions.outOfGas(ex);
      case SHL, SHR, SAR -> this.shf = !Exceptions.outOfGas(ex);
      case SHA3 -> {
        this.mxp = true;
        this.hashInfo = Exceptions.none(ex) && !frame.getStackItem(1).isZero();
        this.mmu = this.hashInfo;
      }
      case BALANCE, EXTCODESIZE, EXTCODEHASH, SELFDESTRUCT -> this.trm = true;
      case MLOAD, MSTORE, MSTORE8 -> {
        this.mxp = true;
        this.mmu = Exceptions.none(ex);
      }
      case CALLDATALOAD -> {
        this.oob = true;
        this.mmu =
            Exceptions.none(ex)
                && frame.getInputData().size() > Words.clampedToLong(frame.getStackItem(0));
      }
      case SLOAD -> {}
      case SSTORE, JUMP, JUMPI -> this.oob = true;
      case MSIZE -> this.mxp = Exceptions.none(ex);
      case BLOCKHASH -> this.blockhash = Exceptions.none(ex);
    }
  }
}
