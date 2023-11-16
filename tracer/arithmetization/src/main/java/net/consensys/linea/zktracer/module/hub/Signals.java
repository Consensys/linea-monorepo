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

package net.consensys.linea.zktracer.module.hub;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.evm.frame.MessageFrame;

/**
 * Encodes the signals triggering other components.
 *
 * <p>When a component is requested, also checks that it may actually be triggered in the current
 * context.
 */
@Accessors(fluent = true)
@RequiredArgsConstructor
public class Signals {
  @Getter private boolean mmu;
  @Getter private boolean mxp;
  @Getter private boolean oob;
  @Getter private boolean precompileInfo;
  @Getter private boolean stipend;
  @Getter private boolean exp;
  @Getter private boolean trm;
  @Getter private boolean hashInfo;
  @Getter private boolean romLex;

  private final PlatformController platformController;

  public void reset() {
    this.mmu = false;
    this.mxp = false;
    this.oob = false;
    this.precompileInfo = false;
    this.stipend = false;
    this.exp = false;
    this.trm = false;
    this.hashInfo = false;
    this.romLex = false;
  }

  public Signals snapshot() {
    Signals r = new Signals(null);
    r.mmu = this.mmu;
    r.mxp = this.mxp;
    r.oob = this.oob;
    r.precompileInfo = this.precompileInfo;
    r.stipend = this.stipend;
    r.exp = this.exp;
    r.trm = this.trm;
    r.hashInfo = this.hashInfo;
    r.romLex = this.romLex;

    return r;
  }

  public Signals wantMmu() {
    this.mmu = true;
    return this;
  }

  public Signals wantMxp() {
    this.mxp = true;
    return this;
  }

  public Signals wantStipend() {
    this.stipend = true;
    return this;
  }

  public Signals wantOob() {
    this.oob = true;
    return this;
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
    OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    final Exceptions ex = platformController.exceptions();

    switch (opCode) {
      case CALLDATACOPY, CODECOPY -> {
        this.mxp = ex.outOfMemoryExpansion() || ex.outOfGas() || ex.none();
        this.mmu = ex.none() && !frame.getStackItem(1).isZero();
      }

      case RETURNDATACOPY -> {
        this.oob = ex.returnDataCopyFault() || ex.none();
        this.mxp = ex.outOfMemoryExpansion() || ex.outOfGas() || ex.none();
        this.mmu = ex.none() && !frame.getStackItem(1).isZero();
      }

      case EXTCODECOPY -> {
        this.mxp = ex.outOfMemoryExpansion() || ex.outOfGas() || ex.none();
        this.mmu = ex.none() && !frame.getStackItem(2).isZero();
        this.trm = ex.outOfGas() || ex.none();
      }

      case LOG0, LOG1, LOG2, LOG3, LOG4 -> {
        this.mxp = ex.outOfMemoryExpansion() || ex.outOfGas() || ex.none();
        this.mmu = ex.none() && !frame.getStackItem(1).isZero(); // TODO: retcon to false if REVERT
      }

      case CALL, DELEGATECALL, STATICCALL, CALLCODE -> {
        // WARN: nothing to see here, dynamically requested
      }

      case CREATE, CREATE2 -> {
        // WARN: nothing to see here, cf scenarios – vous qui entrez ici, abandonnez tout espoir
      }

      case REVERT -> {
        this.mxp = ex.outOfMemoryExpansion() || ex.outOfGas() || ex.none();
        this.mmu =
            ex.none()
                && !frame.getStackItem(1).isZero()
                && hub.currentFrame().returnDataTarget().length() > 0;
      }

      case RETURN -> {
        final boolean isDeployment = frame.getType() == MessageFrame.Type.CONTRACT_CREATION;
        final boolean sizeNonZero =
            Hub.maybeStackItem(frame, 1).map(size -> !size.isZero()).orElse(false);

        // WARN: Static part, other modules may be dynamically requested in the hub
        this.mxp =
            ex.outOfMemoryExpansion() || ex.outOfGas() || ex.invalidCodePrefix() || ex.none();
        this.oob = isDeployment && (ex.codeSizeOverflow() || ex.none());
        this.mmu =
            (isDeployment && ex.invalidCodePrefix())
                || (!isDeployment
                    && ex.none()
                    && sizeNonZero
                    && hub.currentFrame().returnDataTarget().length() > 0)
                || (isDeployment && ex.none() && sizeNonZero);
        this.romLex = this.hashInfo = isDeployment && ex.none() && sizeNonZero;
      }

        // TODO: these opcodes
      case ADD -> {}
      case MUL -> {}
      case SUB -> {}
      case DIV -> {}
      case SDIV -> {}
      case MOD -> {}
      case SMOD -> {}
      case ADDMOD -> {}
      case MULMOD -> {}
      case EXP -> {}
      case SIGNEXTEND -> {}
      case LT -> {}
      case GT -> {}
      case SLT -> {}
      case SGT -> {}
      case EQ -> {}
      case ISZERO -> {}
      case AND -> {}
      case OR -> {}
      case XOR -> {}
      case NOT -> {}
      case BYTE -> {}
      case SHL -> {}
      case SHR -> {}
      case SAR -> {}
      case SHA3 -> {}
      case ADDRESS -> {}
      case BALANCE -> {}
      case ORIGIN -> {}
      case CALLER -> {}
      case CALLVALUE -> {}
      case CALLDATALOAD -> {}
      case CALLDATASIZE -> {}
      case CODESIZE -> {}
      case GASPRICE -> {}
      case EXTCODESIZE -> {}
      case RETURNDATASIZE -> {}
      case EXTCODEHASH -> {}
      case BLOCKHASH -> {}
      case COINBASE -> {}
      case TIMESTAMP -> {}
      case NUMBER -> {}
      case DIFFICULTY -> {}
      case GASLIMIT -> {}
      case CHAINID -> {}
      case SELFBALANCE -> {}
      case BASEFEE -> {}
      case POP -> {}
      case MLOAD -> {}
      case MSTORE -> {}
      case MSTORE8 -> {}
      case SLOAD -> {}
      case SSTORE -> {}
      case JUMP -> {}
      case JUMPI -> {}
      case PC -> {}
      case MSIZE -> {}
      case GAS -> {}
      case JUMPDEST -> {}
      case PUSH1 -> {}
      case PUSH2 -> {}
      case PUSH3 -> {}
      case PUSH4 -> {}
      case PUSH5 -> {}
      case PUSH6 -> {}
      case PUSH7 -> {}
      case PUSH8 -> {}
      case PUSH9 -> {}
      case PUSH10 -> {}
      case PUSH11 -> {}
      case PUSH12 -> {}
      case PUSH13 -> {}
      case PUSH14 -> {}
      case PUSH15 -> {}
      case PUSH16 -> {}
      case PUSH17 -> {}
      case PUSH18 -> {}
      case PUSH19 -> {}
      case PUSH20 -> {}
      case PUSH21 -> {}
      case PUSH22 -> {}
      case PUSH23 -> {}
      case PUSH24 -> {}
      case PUSH25 -> {}
      case PUSH26 -> {}
      case PUSH27 -> {}
      case PUSH28 -> {}
      case PUSH29 -> {}
      case PUSH30 -> {}
      case PUSH31 -> {}
      case PUSH32 -> {}
      case DUP1 -> {}
      case DUP2 -> {}
      case DUP3 -> {}
      case DUP4 -> {}
      case DUP5 -> {}
      case DUP6 -> {}
      case DUP7 -> {}
      case DUP8 -> {}
      case DUP9 -> {}
      case DUP10 -> {}
      case DUP11 -> {}
      case DUP12 -> {}
      case DUP13 -> {}
      case DUP14 -> {}
      case DUP15 -> {}
      case DUP16 -> {}
      case SWAP1 -> {}
      case SWAP2 -> {}
      case SWAP3 -> {}
      case SWAP4 -> {}
      case SWAP5 -> {}
      case SWAP6 -> {}
      case SWAP7 -> {}
      case SWAP8 -> {}
      case SWAP9 -> {}
      case SWAP10 -> {}
      case SWAP11 -> {}
      case SWAP12 -> {}
      case SWAP13 -> {}
      case SWAP14 -> {}
      case SWAP15 -> {}
      case SWAP16 -> {}
      case INVALID -> {}
      case SELFDESTRUCT -> {}
    }
  }
}
