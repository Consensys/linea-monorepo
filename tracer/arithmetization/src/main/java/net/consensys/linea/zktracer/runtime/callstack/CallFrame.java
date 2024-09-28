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

import static com.google.common.base.Preconditions.*;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.module.romlex.ContractMetadata;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import net.consensys.linea.zktracer.runtime.stack.Stack;
import net.consensys.linea.zktracer.runtime.stack.StackContext;
import net.consensys.linea.zktracer.types.Bytecode;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.frame.MessageFrame;

@Accessors(fluent = true)
public class CallFrame {
  public static final CallFrame EMPTY = new CallFrame();

  @Setter public int universalParentReturnDataContextNumber;

  /** the position of this {@link CallFrame} in the {@link CallStack}. */
  @Getter private int id;

  /** the context number of the frame, i.e. the hub stamp at its creation */
  @Getter private final int contextNumber;

  /** the depth of this CallFrame within its call hierarchy. */
  @Getter private int depth;

  /** true iff the current context was spawned by a deployment transaction or a CREATE(2) opcode */
  @Getter private boolean isDeployment;

  public boolean isMessageCall() {
    return !isDeployment;
  }

  /** the ID of this {@link CallFrame} parent in the {@link CallStack}. */
  @Getter private int callerId;

  /** all the {@link CallFrame} that have been called by this frame. */
  @Getter private final List<Integer> childFramesId = new ArrayList<>();

  /** the {@link Address} of the account executing this {@link CallFrame}. */
  @Getter private final Address accountAddress;

  @Getter private int accountDeploymentNumber;

  /** A memoized {@link EWord} conversion of `address` */
  private EWord eAddress = null;

  /** the {@link Address} of the code executed in this {@link CallFrame}. */
  @Getter private Address byteCodeAddress = Address.ZERO;

  @Getter private int byteCodeDeploymentNumber;

  /** the {@link Bytecode} executing within this frame. */
  @Getter private Bytecode code = Bytecode.EMPTY;

  /** the CFI of this frame bytecode if applicable */
  @Getter private int codeFragmentIndex = -1;

  /** A memoized {@link EWord} conversion of `codeAddress` */
  private EWord eCodeAddress = null;

  @Getter private Address callerAddress = Address.ZERO;

  /** the {@link CallFrameType} of this frame. */
  @Getter private final CallFrameType type;

  public int getCodeFragmentIndex(Hub hub) {
    return this == CallFrame.EMPTY || type == CallFrameType.TRANSACTION_CALL_DATA_HOLDER
        ? 0
        : hub.getCfiByMetaData(byteCodeAddress, byteCodeDeploymentNumber, isDeployment);
  }

  @Getter @Setter private int pc;
  @Getter @Setter private OpCode opCode = OpCode.STOP;
  @Getter @Setter private OpCodeData opCodeData = OpCodes.of(OpCode.STOP);
  @Getter private MessageFrame frame; // TODO: can we make this final ?

  /** the ether amount given to this frame. */
  @Getter private Wei value = Wei.fromHexString("0xBadF00d"); // Marker for debugging

  /** the gas given to this frame. */
  @Getter private long gasStipend;

  /** the call data given to this frame. */
  @Getter CallDataInfo callDataInfo;

  /** the latest child context to have been called from this frame */
  @Getter @Setter private int returnDataContextNumber = 0;

  /** the data returned by the latest callee. */
  @Getter @Setter private Bytes returnData = Bytes.EMPTY;

  /** returnData position within the latest callee memory space. */
  @Getter @Setter private MemorySpan returnDataSpan = new MemorySpan(0, 0);

  /** the return data provided by this frame */
  @Getter @Setter private Bytes outputData = Bytes.EMPTY;

  /** where this frame store its return data in its own RAM */
  @Getter @Setter private MemorySpan outputDataSpan;

  /** where this frame is expected to write its outputData within its parent's memory space. */
  @Getter private MemorySpan returnDataTargetInCaller = MemorySpan.empty();

  @Getter @Setter private boolean selfReverts = false;
  @Getter @Setter private boolean getsReverted = false;

  /** the hub stamp at which this frame reverts (0 means it does not revert) */
  @Getter @Setter private int revertStamp = 0;

  /** this frame {@link Stack}. */
  @Getter private final Stack stack = new Stack();

  /** the latched context of this callframe stack. */
  @Getter @Setter private StackContext pending;

  /**
   * the section responsible for the creation of a child context, either a CALL or a CREATE
   * instruction
   */
  @Getter @Setter private TraceSection childSpanningSection;

  public static void updateParentContextReturnData(
      Hub hub, Bytes outputData, MemorySpan returnDataSource) {
    CallFrame parent = hub.callStack().parent();
    parent.returnDataContextNumber = hub.currentFrame().contextNumber;
    parent.returnData = outputData;
    parent.outputDataSpan(returnDataSource);
  }

  /** Create a MANTLE call frame. */
  CallFrame(final Address origin, final Bytes callData, final int contextNumber) {
    type = CallFrameType.TRANSACTION_CALL_DATA_HOLDER;
    this.contextNumber = contextNumber;
    accountAddress = origin;
    callDataInfo = new CallDataInfo(callData, 0, callData.size(), contextNumber);
  }

  // TODO: should die ?
  /** Create a PRECOMPILE_RETURN_DATA callFrame */
  CallFrame(
      final int contextNumber,
      final Bytes precompileResult,
      final int returnDataOffset,
      final Address precompileAddress) {
    checkArgument(
        returnDataOffset == 0 || precompileAddress == Address.MODEXP,
        "ReturnDataOffset is 0 for all precompile except Modexp");
    type = CallFrameType.PRECOMPILE_RETURN_DATA;
    this.contextNumber = contextNumber;
    outputData = precompileResult;
    outputDataSpan = new MemorySpan(returnDataOffset, precompileResult.size());
    accountAddress = precompileAddress;
  }

  /** Create an empty call frame. */
  CallFrame() {
    type = CallFrameType.EMPTY;
    contextNumber = 0;
    accountAddress = Address.ZERO;
    callerId = -1;
    callDataInfo = new CallDataInfo(Bytes.EMPTY, 0, 0, 0);
  }

  /**
   * Create a non-root call frame. Below we abbreviate Context Number to CN
   *
   * @param type the {@link CallFrameType} of this frame
   * @param id ID of this frame in the {@link CallStack}
   * @param contextNumber of this call frame
   * @param depth call stack depth of the current execution context
   * @param isDeployment whether the executing byteCode is initcode
   * @param value how much ether was given to this frame
   * @param gasStipend how much gasStipend was given to this frame
   * @param accountAddress {@link Address} of this frame executor
   * @param accountDeploymentNumber DN of the account address
   * @param byteCodeAddress address whose byteCode executes in the present frame
   * @param byteCodeDeploymentNumber DN of this call frame in the {@link Hub}
   * @param byteCode byteCode that executes in the present context
   * @param callerAddress either account address of the caller/creator context
   * @param callDataContextNumber CN of the RAM segment wherein the call data lives
   * @param callerId ID of the caller frame in the {@link CallStack}
   * @param callData {@link Bytes} containing this frame's call data
   * @param callDataOffset offset of call data in the caller's RAM (if applicable)
   * @param callDataSize size (in bytes) of the call data
   */
  CallFrame(
      CallFrameType type,
      int id,
      int contextNumber,
      int depth,
      boolean isDeployment,
      Wei value,
      long gasStipend,
      Address accountAddress,
      int accountDeploymentNumber,
      Address byteCodeAddress,
      int byteCodeDeploymentNumber,
      Bytecode byteCode,
      Address callerAddress,
      long callDataContextNumber,
      int callerId,
      Bytes callData,
      long callDataOffset,
      long callDataSize) {
    this.type = type;
    this.id = id;
    this.contextNumber = contextNumber;
    this.isDeployment = isDeployment;
    this.value = value;
    this.gasStipend = gasStipend;
    this.depth = depth;
    this.accountAddress = accountAddress;
    this.accountDeploymentNumber = accountDeploymentNumber;
    this.byteCodeAddress = byteCodeAddress;
    this.byteCodeDeploymentNumber = byteCodeDeploymentNumber;
    this.code = byteCode;
    this.callerAddress = callerAddress;
    this.callerId = callerId;
    this.callDataInfo =
        new CallDataInfo(callData, callDataOffset, callDataSize, callDataContextNumber);
    this.outputDataSpan = MemorySpan.empty();
    this.returnDataSpan = MemorySpan.empty();
    this.returnDataTargetInCaller = MemorySpan.empty();
  }

  public boolean isRoot() {
    return this.depth == 0;
  }

  /**
   * Return the address of this CallFrame as an {@link EWord}.
   *
   * @return the address
   */
  public EWord addressAsEWord() {
    if (this.eAddress == null) {
      this.eAddress = EWord.of(this.accountAddress);
    }
    return this.eAddress;
  }

  /**
   * Return the address of the code executed within this callframe as an {@link EWord}.
   *
   * @return the address
   */
  public EWord codeAddressAsEWord() {
    if (this.eCodeAddress == null) {
      this.eCodeAddress = EWord.of(this.byteCodeAddress);
    }
    return this.eCodeAddress;
  }

  /**
   * If any, returns the ID of the latest callee of this frame.
   *
   * @return the ID of the latest callee
   */
  public Optional<Integer> lastCallee() {
    if (this.childFramesId.isEmpty()) {
      return Optional.empty();
    }

    return Optional.of(this.childFramesId.get(this.childFramesId.size() - 1));
  }

  /**
   * Returns a {@link ContractMetadata} instance representing the executed contract.
   *
   * @return the executed contract metadata
   */
  public ContractMetadata metadata() {
    return ContractMetadata.make(byteCodeAddress, byteCodeDeploymentNumber, isDeployment);
  }

  private void revertChildren(CallStack callStack, int parentRevertStamp) {
    childFramesId.stream()
        .map(callStack::getById)
        .forEach(
            frame -> {
              frame.getsReverted = true;
              if (!frame.selfReverts) {
                frame.revertStamp = parentRevertStamp;
              }
              frame.revertChildren(callStack, parentRevertStamp);
            });
  }

  public void setRevertStamps(CallStack callStack, int currentStamp) {
    if (selfReverts) {
      throw new IllegalStateException(
          String.format(
              "a context can not self-revert twice, it already reverts at %s, can't revert again at %s",
              revertStamp, currentStamp));
    }
    selfReverts = true;
    revertStamp = currentStamp;
    this.revertChildren(callStack, revertStamp);
  }

  public boolean willRevert() {
    return selfReverts() || getsReverted();
  }

  public void initializeFrame(final MessageFrame frame) {
    this.frame = frame;
  }

  public void frame(MessageFrame frame) {
    this.frame = frame;
    opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    opCodeData = OpCodes.of(opCode);
    pc = frame.getPC();
  }

  public static Bytes extractContiguousLimbsFromMemory(
      final MessageFrame frame, final MemorySpan memorySpan) {
    // TODO: optimize me please. Need a review of the MMU operation handling.
    return memorySpan.isEmpty() ? Bytes.EMPTY : frame.shadowReadMemory(0, frame.memoryByteSize());
  }

  public OpCode getOpCode() {
    return getOpCode(frame);
  }

  public static OpCode getOpCode(MessageFrame frame) {
    return OpCode.of(0xFF & frame.getCurrentOperation().getOpcode());
  }
}
