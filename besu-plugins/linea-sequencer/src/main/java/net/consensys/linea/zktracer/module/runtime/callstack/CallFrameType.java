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

package net.consensys.linea.zktracer.module.runtime.callstack;

import net.consensys.linea.zktracer.opcode.OpCode;

public enum CallFrameType {
  /** Executing deployment code. */
  INIT_CODE,
  /** Executing standard contract. */
  STANDARD,
  /** Within a delegate call. */
  DELEGATE,
  /** Within a static call. */
  STATIC,
  /** Within a call code. */
  CALL_CODE,
  /** The bedrock context. */
  BEDROCK;

  /**
   * Returns the kind of {@link CallFrameType} context that an opcode will create; throws if the
   * opcode does not create a new context.
   *
   * @param opCode a context-changing {@link OpCode}
   * @return the associated {@link CallFrameType}
   */
  public CallFrameType ofOpCode(OpCode opCode) {
    if (this.isStatic()) {
      return STATIC;
    } else {
      return switch (opCode) {
        case CREATE, CREATE2 -> INIT_CODE;
        case DELEGATECALL -> DELEGATE;
        case CALLCODE -> CALL_CODE;
        case STATICCALL -> STATIC;
        default -> throw new IllegalStateException(String.valueOf(opCode));
      };
    }
  }

  public boolean isStatic() {
    return this == STATIC;
  }
}
