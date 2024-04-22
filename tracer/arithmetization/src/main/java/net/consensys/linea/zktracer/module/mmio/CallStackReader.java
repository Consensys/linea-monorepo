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

  public Bytes valueFromMemory(final long contextNumber, final boolean ramIsSource) {
    final CallFrame callFrame = callStack.getByContextNumber(contextNumber);

    if (callFrame.type() == CallFrameType.MANTLE || callFrame.type() == CallFrameType.BEDROCK) {
      return callFrame.callDataInfo().data();
    }

    if (callFrame.type() == CallFrameType.PRECOMPILE_RETURN_DATA) {
      if (ramIsSource) {
        return callFrame.returnData();
      } else {
        return Bytes.EMPTY;
      }
    }

    final MessageFrame messageFrame = callFrame.frame();

    return messageFrame.shadowReadMemory(0, messageFrame.memoryByteSize());
  }
}
