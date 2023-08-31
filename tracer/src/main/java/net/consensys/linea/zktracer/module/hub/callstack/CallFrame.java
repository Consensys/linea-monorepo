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

package net.consensys.linea.zktracer.module.hub.callstack;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.memory.MemorySpan;
import net.consensys.linea.zktracer.module.hub.stack.Stack;
import net.consensys.linea.zktracer.module.hub.stack.StackContext;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.Code;

public class CallFrame {
  /** the position of this {@link CallFrame} in the {@link CallStack}. */
  @Getter private int id;
  /** the associated context number in the {@link Hub}. */
  @Getter private int contextNumber;
  /** */
  @Getter private int accountDeploymentNumber;
  /** */
  @Getter private int codeDeploymentNumber;
  /** */
  @Getter private boolean codeDeploymentStatus;
  /** the position of this {@link CallFrame} parent in the {@link CallStack}. */
  @Getter private int parentFrame;
  /** all the {@link CallFrame} that have been called by this frame. */
  @Getter private List<Integer> childFrames = new ArrayList<>();

  /** the {@link Address} of the account executing this {@link CallFrame}. */
  @Getter private Address address;
  /** the {@link Address} of the code executed in this {@link CallFrame}. */
  @Getter private Address codeAddress;

  /** the {@link CallFrameType} of this frame. */
  @Getter private CallFrameType type;

  /** the {@link Code} executing within this frame. */
  private Code code;

  /** the ether amount given to this frame. */
  @Getter private Wei value;
  /** the gas given to this frame. */
  @Getter private long gasEndowment;
  // @Getter RevertReason revertReason;

  /** where does this frame start in the {@link Hub} trace. */
  private int startLine;
  /** where does this frame end in the {@link Hub} trace. */
  private int endLine;

  /** the call data given to this frame. */
  @Getter private Bytes callData;
  /** the call data span in the parent memory. */
  @Getter private MemorySpan callDataPointer;
  /** the data returned by the latest callee. */
  @Getter @Setter private Bytes returnData;
  /** returnData position within the latest callee memory space. */
  @Getter private MemorySpan returnDataPointer;
  /** where this frame is expected to write its returnData within its parent's memory space. */
  @Getter private MemorySpan returnDataTarget;

  /** this frame {@link Stack}. */
  @Getter private final Stack stack = new Stack();

  /** the latched context of this callframe stack. */
  @Getter @Setter private StackContext pending;

  /** Create a root call frame. */
  CallFrame() {
    this.type = CallFrameType.BEDROCK;
  }

  /**
   * Create a normal (non-root) call frame.
   *
   * @param contextNumber the CN of this frame in the {@link Hub}
   * @param accountDeploymentNumber the DN of this frame in the {@link Hub}
   * @param codeDeploymentNumber the DN of this frame in the {@link Hub}
   * @param codeDeploymentStatus the DN of this frame in the {@link Hub}
   * @param id the ID of this frame in the {@link CallStack}
   * @param address the {@link Address} of this frame executor
   * @param type the {@link CallFrameType} of this frame
   * @param caller the ID of this frame caller in the {@link CallStack}
   * @param value how much ether was given to this frame
   * @param gas how much gas was given to this frame
   * @param currentLine where does this frame start in the {@link Hub} trace
   * @param callData {@link Bytes} containing this frame call data
   */
  CallFrame(
      int contextNumber,
      int accountDeploymentNumber,
      int codeDeploymentNumber,
      boolean codeDeploymentStatus,
      int id,
      Address address,
      Code code,
      CallFrameType type,
      int caller,
      Wei value,
      long gas,
      int currentLine,
      Bytes callData) {
    this.contextNumber = contextNumber;
    this.accountDeploymentNumber = accountDeploymentNumber;
    this.codeDeploymentNumber = codeDeploymentNumber;
    this.codeDeploymentStatus = codeDeploymentStatus;
    this.id = id;
    this.address = address;
    this.code = code;
    this.type = type;
    this.parentFrame = caller;
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
   * Return the address of this callframe as an {@link EWord}.
   *
   * @return the address
   */
  public EWord addressAsEWord() {
    return EWord.of(this.address);
  }

  /**
   * Return the address of the code executed within this callframe as an {@link EWord}.
   *
   * @return the address
   */
  public EWord codeAddressAsEWord() {
    return EWord.of(this.codeAddress);
  }

  /**
   * If any, returns the ID of the latest callee of this frame.
   *
   * @return the ID of the latest callee
   */
  public Optional<Integer> lastCallee() {
    if (this.childFrames.isEmpty()) {
      return Optional.empty();
    }

    return Optional.of(this.childFrames.get(this.childFrames.size() - 1));
  }
}
