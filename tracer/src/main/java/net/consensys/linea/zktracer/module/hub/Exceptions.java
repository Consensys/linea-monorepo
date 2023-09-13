/*
 * Copyright ConsenSys AG.
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

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import org.hyperledger.besu.evm.frame.MessageFrame;

public record Exceptions(
    boolean invalidOpcode,
    boolean stackUnderflow,
    boolean stackOverflow,
    boolean outOfMemoryExpansion,
    boolean outOfGas,
    boolean returnDataCopyFault,
    boolean jumpFault,
    boolean staticViolation,
    boolean outOfSStore,
    boolean invalidCodePrefix,
    boolean codeSizeOverflow) {
  /**
   * @return true if no stack exception has been raised
   */
  public boolean noStackException() {
    return !this.stackOverflow() && !this.stackUnderflow();
  }

  public Exceptions snapshot() {
    return new Exceptions(
        invalidOpcode,
        stackUnderflow,
        stackOverflow,
        outOfMemoryExpansion,
        outOfGas,
        returnDataCopyFault,
        jumpFault,
        staticViolation,
        outOfSStore,
        invalidCodePrefix,
        codeSizeOverflow);
  }
  /**
   * @return true if any exception flag has been raised
   */
  public boolean any() {
    return this.invalidOpcode
        || this.stackUnderflow
        || this.stackOverflow
        || this.outOfMemoryExpansion
        || this.outOfGas
        || this.returnDataCopyFault
        || this.jumpFault
        || this.staticViolation
        || this.outOfSStore
        || this.invalidCodePrefix
        || this.codeSizeOverflow;
  }

  /**
   * Compute all the exceptions that may have happened in the current frame and package them in an
   * {@link Exceptions} record.
   *
   * @param frame the context from which to compute the putative exceptions
   * @return all {@link Exceptions} relative to the given frame
   */
  public static Exceptions fromFrame(MessageFrame frame) {
    OpCodeData opCode = OpCode.of(frame.getCurrentOperation().getOpcode()).getData();
    return new Exceptions(
        opCode.mnemonic() == OpCode.INVALID,
        frame.stackSize() < opCode.stackSettings().nbRemoved(),
        frame.stackSize() + opCode.stackSettings().nbAdded() - opCode.stackSettings().nbRemoved()
            > 1024,
        false, // TODO mxp
        false, // TODO OoG
        false, // TODO
        false, // TODO
        frame.isStatic() && !opCode.stackSettings().staticInstruction(),
        false, // TODO
        false, // TODO
        false // TODO
        );
  }
}
