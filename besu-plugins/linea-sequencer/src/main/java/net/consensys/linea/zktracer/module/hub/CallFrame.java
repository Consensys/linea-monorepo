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

import java.util.List;

import lombok.Getter;
import net.consensys.linea.zktracer.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.Code;

class CallFrame {
  /** the associated context number in the {@link Hub} */
  @Getter int contextNumber;
  /** the associated deployment number in the {@link Hub} */
  @Getter int deploymentNumber;
  /** the position of this {@link CallFrame} in the {@link CallStack} */
  @Getter int id;
  /** the position of this {@link CallFrame} parent in the {@link CallStack} */
  @Getter int parentFrame;
  /** the ID of the latest {@link CallFrame} this frame called */
  @Getter int lastCalled;
  /** all the {@link CallFrame} that have been called by this frame */
  @Getter List<Integer> childFrames;

  /** the {@link Address} of the account executing this {@link CallFrame} */
  @Getter Address address;
  /** the {@link Address} of the code executed in this {@link CallFrame} */
  @Getter Address codeAddress;
  /** the {@link Address} of the parent context of this {@link CallFrame} */
  @Getter Address parentAddress;

  /** the {@link CallFrameType} of this frame */
  @Getter CallFrameType type;
  /** the {@link Code} executing within this frame */
  Code code;

  /** the ether amount given to this frame */
  @Getter Wei value;
  /** the gas given to this frame */
  @Getter long gasEndowment;
  // @Getter RevertReason revertReason;

  /** where does this frame start in the {@link Hub} trace */
  int startLine;
  /** where does this frame end in the {@link Hub} trace */
  int endLine;

  /** the call data given to this frame */
  @Getter Bytes callData;
  /** the data this frame will return */
  @Getter Bytes returnData;

  /** this frame {@link Stack} */
  @Getter final Stack stack = new Stack();

  /** the latched context of this callframe stack */
  StackContext pending;

  /** Create a root call frame */
  CallFrame() {
    this.type = CallFrameType.Root;
  }

  /**
   * Create a normal (non-root) call frame
   *
   * @param contextNumber the CN of this frame in the {@link Hub}
   * @param deploymentNumber the DN of this frame in the {@link Hub}
   * @param id the ID of this frame in the {@link CallStack}
   * @param address the {@link Address} of this frame executor
   * @param type the {@link CallFrameType} of this frame
   * @param caller the ID of this frame caller in the {@link CallStack}
   * @param parentAddress the address of this frame caller
   * @param value how much ether was given to this frame
   * @param gas how much gas was given to this frame
   * @param currentLine where does this frame start in the {@link Hub} trace
   * @param callData {@link Bytes} containing this frame call data
   */
  CallFrame(
      int contextNumber,
      int deploymentNumber,
      int id,
      Address address,
      Code code,
      CallFrameType type,
      int caller,
      Address parentAddress,
      Wei value,
      long gas,
      int currentLine,
      Bytes callData) {
    this.contextNumber = contextNumber;
    this.deploymentNumber = deploymentNumber;
    this.id = id;
    this.address = address;
    this.code = code;
    this.type = type;
    this.parentFrame = caller;
    this.parentAddress = parentAddress;
    this.value = value;
    this.gasEndowment = gas;
    this.startLine = currentLine;
    this.endLine = currentLine;
    this.callData = callData;
  }

  void close(int line) {
    this.endLine = line;
  }

  /**
   * Return the address of this callframe as an {@link EWord}
   *
   * @return the address
   */
  public EWord addressAsEWord() {
    return EWord.of(this.address);
  }
}
