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

import net.consensys.linea.zktracer.module.runtime.callstack.CallFrame;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

/**
 * Encode the exceptions that can be raised at the end of the execution of a call frame.
 *
 * @param invalidCodePrefix trying to deploy a contract starting with 0xEF
 * @param codeSizeOverflow trying to deploy a contract larger than 24KB
 */
public record ContextExceptions(boolean invalidCodePrefix, boolean codeSizeOverflow) {
  private static final byte EIP_3541_MARKER = (byte) 0xEF;
  private static final int MAX_CODE_SIZE = 24576;

  private static boolean isInvalidCodePrefix(MessageFrame frame) {
    final Bytes deployedCode = frame.getOutputData();
    return !deployedCode.isEmpty() && (deployedCode.get(0) == EIP_3541_MARKER);
  }

  private static boolean isCodeSizeOverflow(MessageFrame frame) {
    final Bytes deployedCode = frame.getOutputData();

    return deployedCode.size() > MAX_CODE_SIZE;
  }

  public static ContextExceptions fromFrame(final CallFrame callFrame, final MessageFrame frame) {
    if (callFrame.isCodeDeploymentStatus()) {
      return new ContextExceptions(isInvalidCodePrefix(frame), isCodeSizeOverflow(frame));
    } else {
      return new ContextExceptions(false, false);
    }
  }

  public boolean any() {
    return this.invalidCodePrefix || this.codeSizeOverflow;
  }
}
