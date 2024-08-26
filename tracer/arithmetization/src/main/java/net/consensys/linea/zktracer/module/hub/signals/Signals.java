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
  @Getter private boolean mxp;
  @Getter private boolean oob;
  @Getter private boolean stp;
  @Getter private boolean exp;
  @Getter private boolean trm;
  @Getter private boolean hashInfo;
  @Getter private boolean rlpAddr;

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
    this.mxp = false;
    this.oob = false;
    this.stp = false;
    this.exp = false;
    this.trm = false;
    this.hashInfo = false;
    this.rlpAddr = false;
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
    r.mxp = this.mxp;
    r.oob = this.oob;
    r.stp = this.stp;
    r.exp = this.exp;
    r.trm = this.trm;
    r.hashInfo = this.hashInfo;
    r.rlpAddr = this.rlpAddr;

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
      case CALL, DELEGATECALL, STATICCALL, CALLCODE -> {
        this.mxp = !Exceptions.staticFault(ex);
        this.stp = Exceptions.outOfGasException(ex) || Exceptions.none(ex);
        this.oob = opCode.equals(OpCode.CALL) && Exceptions.staticFault(ex) || Exceptions.none(ex);
        this.trm = Exceptions.outOfGasException(ex) || Exceptions.none(ex);

        final boolean triggersAbortingCondition =
            Exceptions.none(ex) && this.platformController.abortingConditions().any();

        final Address target = Words.toAddress(frame.getStackItem(1));
        final boolean targetAddressHasNonEmptyCode =
            Optional.ofNullable(frame.getWorldUpdater().get(target))
                .map(AccountState::hasCode)
                .orElse(false);

        this.exp =
            Exceptions.none(ex)
                && this.platformController.abortingConditions().none()
                && target.equals(Address.MODEXP);
      }

      case REVERT -> this.mxp =
          Exceptions.memoryExpansionException(ex)
              || Exceptions.outOfGasException(ex)
              || Exceptions.none(ex);

      case EXP -> {
        this.exp = true; // TODO: use expCall instead
        this.mul = !Exceptions.outOfGasException(ex);
      }

        // other opcodes
      case ADD, SUB -> this.add = !Exceptions.outOfGasException(ex);
      case MUL -> this.mul = !Exceptions.outOfGasException(ex);
      case DIV, SDIV, MOD, SMOD -> this.mod = !Exceptions.outOfGasException(ex);
      case ADDMOD, MULMOD -> this.ext = !Exceptions.outOfGasException(ex);
      case LT, GT, SLT, SGT, EQ, ISZERO -> this.wcp = !Exceptions.outOfGasException(ex);
      case AND, OR, XOR, NOT, SIGNEXTEND, BYTE -> this.bin = !Exceptions.outOfGasException(ex);
      case SHL, SHR, SAR -> this.shf = !Exceptions.outOfGasException(ex);
      case SHA3 -> {
        this.mxp = true;
        this.hashInfo = Exceptions.none(ex) && !frame.getStackItem(1).isZero();
      }
      case BALANCE, EXTCODESIZE, EXTCODEHASH, SELFDESTRUCT -> this.trm = true;
      case MLOAD, MSTORE, MSTORE8 -> {
        this.mxp = true;
      }
      case CALLDATALOAD -> this.oob = true;
      case SLOAD -> {}
      case SSTORE, JUMP, JUMPI -> this.oob = true;
      case MSIZE -> this.mxp = Exceptions.none(ex);
      case BLOCKHASH -> this.blockhash = Exceptions.none(ex);
    }
  }
}
