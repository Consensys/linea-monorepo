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

package net.consensys.linea.zktracer.module.mmio;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.runtime.callstack.CallFrameType;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

@Getter
@RequiredArgsConstructor
@Accessors(fluent = true)
public class CallStackReader {
  private final CallStack callStack;

  /**
   * Performs a full copy of the memory based on the identified context.
   *
   * @param contextNumber
   * @param ramIsSource
   * @return returns either the current contents of memory of an execution context or the return
   *     data of a precompile or the calldata of a transaction depending on the type of the context.
   */
  public Bytes fullCopyOfContextMemory(final long contextNumber, final boolean ramIsSource) {
    final CallFrame callFrame = callStack.getByContextNumber(contextNumber);

    if (callFrame.type() == CallFrameType.TRANSACTION_CALL_DATA_HOLDER) {
      return callFrame.callDataInfo().data();
    }

    if (callFrame.type() == CallFrameType.PRECOMPILE_RETURN_DATA) {
      return ramIsSource ? callFrame.outputData() : Bytes.EMPTY;
    }

    final MessageFrame messageFrame = callFrame.frame();

    return messageFrame.shadowReadMemory(0, messageFrame.memoryByteSize());
  }
}
