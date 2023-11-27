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

package net.consensys.linea.zktracer.runtime.callstack;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Bytecode;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.memory.MemorySpan;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import net.consensys.linea.zktracer.runtime.stack.Stack;
import net.consensys.linea.zktracer.runtime.stack.StackContext;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.MemoryRange;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.frame.MessageFrame;

@Accessors(fluent = true)
public class CallFrame {
  public static final CallFrame EMPTY = new CallFrame(Address.ZERO);
  /** the position of this {@link CallFrame} in the {@link CallStack}. */
  @Getter private int id;
  /** the context number of the frame, i.e. the hub stamp at its creation */
  @Getter private final int contextNumber;
  /** the depth of this CallFrame within its call hierarchy. */
  @Getter private int depth;
  /** */
  @Getter private int accountDeploymentNumber;
  /** */
  @Getter private int codeDeploymentNumber;
  /** */
  @Getter private boolean underDeployment;

  @Getter @Setter private TraceSection needsUnlatchingAtReEntry = null;

  /** the position of this {@link CallFrame} parent in the {@link CallStack}. */
  @Getter private int parentFrame;
  /** all the {@link CallFrame} that have been called by this frame. */
  @Getter private final List<Integer> childFrames = new ArrayList<>();

  /** the {@link Address} of the account executing this {@link CallFrame}. */
  @Getter private final Address address;
  /** the {@link Address} of the code executed in this {@link CallFrame}. */
  @Getter private Address codeAddress = Address.ZERO;

  /** the {@link CallFrameType} of this frame. */
  @Getter private final CallFrameType type;

  /** the {@link Bytecode} executing within this frame. */
  @Getter private Bytecode code = Bytecode.EMPTY;

  @Getter @Setter private int pc;
  @Getter @Setter private OpCode opCode = OpCode.STOP;
  @Getter @Setter private OpCodeData opCodeData = OpCodes.of(OpCode.STOP);
  @Getter private MessageFrame frame;

  /** the ether amount given to this frame. */
  @Getter private Wei value = Wei.fromHexString("0xbadf00d"); // Marker for debugging
  /** the gas given to this frame. */
  @Getter private long gasEndowment;

  /** the call data given to this frame. */
  @Getter private Bytes callData = Bytes.EMPTY;
  /** the call data span in the parent memory. */
  @Getter private final MemorySpan callDataPointer;
  /** the data returned by the latest callee. */
  @Getter @Setter private Bytes returnData = Bytes.EMPTY;
  /** returnData position within the latest callee memory space. */
  @Getter @Setter private MemorySpan returnDataPointer = new MemorySpan(0, 0);
  /** where this frame is expected to write its returnData within its parent's memory space. */
  @Getter private final MemorySpan returnDataTarget;

  // where I should put my RETURNDATARange in my caller's RAM
  @Getter private MemoryRange returnTarget;

  // where my CALLDATA is in my caller's RAM
  @Getter private MemoryRange callDataRange;

  // position of the returner's RETURNDATARange in its RAM
  @Getter private MemoryRange returnDataRange;

  // last called context
  @Getter private int returner;

  @Getter @Setter private int selfRevertsAt = 0;
  @Getter @Setter private int getsRevertedAt = 0;

  /** this frame {@link Stack}. */
  @Getter private final Stack stack = new Stack();

  /** the latched context of this callframe stack. */
  @Getter @Setter private StackContext pending;

  /** Create a root call frame. */
  CallFrame(Address address) {
    this.type = CallFrameType.BEDROCK;
    this.contextNumber = 0;
    this.address = address;
    this.callDataPointer = new MemorySpan(0, 0);
    this.returnDataTarget = new MemorySpan(0, 0);
  }

  /**
   * Create a normal (non-root) call frame.
   *
   * @param accountDeploymentNumber the DN of this frame in the {@link Hub}
   * @param codeDeploymentNumber the DN of this frame in the {@link Hub}
   * @param isDeployment whether the executing code is initcode
   * @param id the ID of this frame in the {@link CallStack}
   * @param hubStamp the hub stamp at the frame creation
   * @param address the {@link Address} of this frame executor
   * @param type the {@link CallFrameType} of this frame
   * @param caller the ID of this frame caller in the {@link CallStack}
   * @param value how much ether was given to this frame
   * @param gas how much gas was given to this frame
   * @param callData {@link Bytes} containing this frame call data
   */
  CallFrame(
      int accountDeploymentNumber,
      int codeDeploymentNumber,
      boolean isDeployment,
      int id,
      int hubStamp,
      Address address,
      Address codeAddress,
      Bytecode code,
      CallFrameType type,
      int caller,
      Wei value,
      long gas,
      Bytes callData,
      int depth) {
    this.accountDeploymentNumber = accountDeploymentNumber;
    this.codeDeploymentNumber = codeDeploymentNumber;
    this.underDeployment = isDeployment;
    this.id = id;
    this.contextNumber = hubStamp + 1;
    this.address = address;
    this.codeAddress = codeAddress;
    this.code = code;
    this.type = type;
    this.parentFrame = caller;
    this.value = value;
    this.gasEndowment = gas;
    this.callData = callData;
    this.callDataPointer = new MemorySpan(0, callData.size());
    this.depth = depth;
    this.returnDataPointer = new MemorySpan(0, 0);
    this.returnDataTarget = new MemorySpan(0, 0); // TODO: fix me Franklin
  }

  /**
   * Return the address of this CallFrame as an {@link EWord}.
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

  private void revertChildren(CallStack callStack, int stamp) {
    if (this.getsRevertedAt == 0) {
      this.getsRevertedAt = stamp;
      this.childFrames.stream()
          .map(callStack::get)
          .forEach(frame -> frame.revertChildren(callStack, stamp));
    }
  }

  public void revert(CallStack callStack, int stamp) {
    if (this.selfRevertsAt == 0) {
      this.selfRevertsAt = stamp;
      this.revertChildren(callStack, stamp);
    } else if (stamp != this.selfRevertsAt) {
      throw new IllegalStateException("a context can not self-reverse twice");
    }
  }

  public boolean hasReverted() {
    return (this.selfRevertsAt > 0) || (this.getsRevertedAt > 0);
  }

  public void frame(MessageFrame frame) {
    this.frame = frame;
    this.opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    this.opCodeData = OpCodes.of(this.opCode);
    this.pc = frame.getPC();
  }
}
