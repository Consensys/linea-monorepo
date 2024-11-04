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

import java.util.Set;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.InstructionFamily;
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
  private static final Set<InstructionFamily> AUTOMATIC_GAS_MODULE_TRIGGER =
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

  private final PlatformController platformController;

  public void reset() {
    add = false;
    blockhash = false;
    bin = false;
    mul = false;
    ext = false;
    mod = false;
    wcp = false;
    shf = false;
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

    if (Exceptions.stackException(ex)) {
      return;
    }

    switch (opCode) {
      case EXP, MUL -> this.mul = !Exceptions.outOfGasException(ex);
      case ADD, SUB -> this.add = !Exceptions.outOfGasException(ex);
      case DIV, SDIV, MOD, SMOD -> this.mod = !Exceptions.outOfGasException(ex);
      case ADDMOD, MULMOD -> this.ext = !Exceptions.outOfGasException(ex);
      case LT, GT, SLT, SGT, EQ, ISZERO -> this.wcp = !Exceptions.outOfGasException(ex);
      case AND, OR, XOR, NOT, SIGNEXTEND, BYTE -> this.bin = !Exceptions.outOfGasException(ex);
      case SHL, SHR, SAR -> this.shf = !Exceptions.outOfGasException(ex);
      case BLOCKHASH -> this.blockhash = Exceptions.none(ex);
    }
  }
}
