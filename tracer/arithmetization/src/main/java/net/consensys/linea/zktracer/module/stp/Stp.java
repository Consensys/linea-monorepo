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

package net.consensys.linea.zktracer.module.stp;

import static java.lang.Long.max;
import static net.consensys.linea.zktracer.types.AddressUtils.getDeploymentAddress;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Conversions.longToBytes32;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.List;

import com.google.common.base.Preconditions;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@RequiredArgsConstructor
public class Stp implements Module {
  private final Hub hub;
  private final Wcp wcp;
  private final Mod mod;

  @Override
  public String moduleKey() {
    return "STP";
  }

  private final StackedSet<StpChunk> chunks = new StackedSet<>();

  @Override
  public void enterTransaction() {
    this.chunks.enter();
  }

  @Override
  public void popTransaction() {
    this.chunks.pop();
  }

  @Override
  public int lineCount() {
    return this.chunks.stream().mapToInt(chunk -> ctMax(chunk) + 1).sum();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void tracePreOpcode(final MessageFrame frame) {
    OpCode opCode = hub.opCode();

    switch (opCode) {
      case CREATE, CREATE2 -> {
        final StpChunk chunk = getCreateData(frame);
        this.chunks.add(chunk);
        this.wcp.callLT(longToBytes32(chunk.gasActual()), Bytes32.ZERO);
        this.wcp.callLT(longToBytes32(chunk.gasActual()), longToBytes32(chunk.gasPrelim()));
        if (!chunk.oogx()) {
          this.mod.callDiv(longToBytes32(getGDiff(chunk)), longToBytes32(64L));
        }
      }
      case CALL, CALLCODE, DELEGATECALL, STATICCALL -> {
        final StpChunk chunk = getCallData(frame);
        this.chunks.add(chunk);
        this.wcp.callLT(longToBytes32(chunk.gasActual()), Bytes32.ZERO);
        if (callCanTransferValue(chunk.opCode())) {
          this.wcp.callISZERO(Bytes32.leftPad(chunk.value()));
        }
        this.wcp.callLT(longToBytes32(chunk.gasActual()), longToBytes32(chunk.gasPrelim()));
        if (!chunk.oogx()) {
          this.mod.callDiv(longToBytes32(getGDiff(chunk)), longToBytes32(64L));
          this.wcp.callLT(chunk.gas().orElseThrow(), longToBytes32(get63of64GDiff(chunk)));
        }
      }
    }
  }

  private StpChunk getCreateData(final MessageFrame frame) {
    final Address to = getDeploymentAddress(frame);
    final Long gasRemaining = frame.getRemainingGas();
    final Long gasMxp = getGasMxpCreate(frame);
    final Long gasPrelim = GasConstants.G_CREATE.cost() + gasMxp;
    return new StpChunk(
        this.hub.opCode(),
        gasRemaining,
        gasPrelim,
        gasRemaining < gasPrelim,
        gasMxp,
        frame.getWorldUpdater().get(frame.getContractAddress()).getBalance(),
        to,
        Bytes32.leftPad(frame.getStackItem(0)));
  }

  private StpChunk getCallData(final MessageFrame frame) {
    final OpCode opcode = this.hub.opCode();
    final long gasActual = frame.getRemainingGas();
    final Bytes32 value =
        callCanTransferValue(opcode) ? Bytes32.leftPad(frame.getStackItem(2)) : Bytes32.ZERO;
    final Address to = Words.toAddress(frame.getStackItem(1));
    final Long gasMxp = getGasMxpCall(frame);
    final boolean toWarm = frame.isAddressWarm(to);
    final boolean toExists =
        opcode == OpCode.CALLCODE
            || (frame.getWorldUpdater().get(to) != null
                && !frame.getWorldUpdater().get(to).isEmpty());

    long gasPrelim = gasMxp;
    if (!value.isZero() && callCanTransferValue(opcode)) {
      gasPrelim += GasConstants.G_CALL_VALUE.cost();
    }
    if (toWarm) {
      gasPrelim += GasConstants.G_WARM_ACCESS.cost();
    } else {
      gasPrelim += GasConstants.G_COLD_ACCOUNT_ACCESS.cost();
    }
    if (!toExists) {
      gasPrelim += GasConstants.G_NEW_ACCOUNT.cost();
    }
    final boolean oogx = gasActual < gasPrelim;
    return new StpChunk(
        opcode,
        gasActual,
        gasPrelim,
        oogx,
        gasMxp,
        frame.getWorldUpdater().get(frame.getContractAddress()).getBalance(),
        to,
        value,
        toExists,
        toWarm,
        Bytes32.leftPad(frame.getStackItem(0)));
  }

  private boolean callCanTransferValue(OpCode opCode) {
    return (opCode == OpCode.CALL) || (opCode == OpCode.CALLCODE);
  }

  // TODO get from Hub.GasProjector
  private Long getGasMxpCreate(final MessageFrame frame) {
    long gasMxp = 0;
    final long offset = Words.clampedToLong(frame.getStackItem(1));
    final long length = Words.clampedToLong(frame.getStackItem(2));
    final long currentMemorySizeInWords = frame.memoryWordSize();
    final long updatedMemorySizeInWords = frame.calculateMemoryExpansion(offset, length);
    if (currentMemorySizeInWords < updatedMemorySizeInWords) {
      // computing the "linear" portion of CREATE2 memory expansion cost
      final long G_mem = GasConstants.G_MEMORY.cost();
      final long squareCurrent = (currentMemorySizeInWords * currentMemorySizeInWords) >> 9;
      final long squareUpdated = (updatedMemorySizeInWords * updatedMemorySizeInWords) >> 9;
      gasMxp +=
          G_mem * (updatedMemorySizeInWords - currentMemorySizeInWords)
              + (squareUpdated - squareCurrent);
    }
    if (OpCode.of(frame.getCurrentOperation().getOpcode()) == OpCode.CREATE2) {
      final long lengthInWords = (length + 31) >> 5; // ⌈ length / 32 ⌉
      gasMxp += lengthInWords * GasConstants.G_KECCAK_256_WORD.cost();
    }
    return gasMxp;
  }

  // TODO get from Hub.GasProjector
  private Long getGasMxpCall(final MessageFrame frame) {
    long gasMxp = 0;

    final int offset =
        callCanTransferValue(OpCode.of(frame.getCurrentOperation().getOpcode())) ? 1 : 0;
    final long cdo = Words.clampedToLong(frame.getStackItem(2 + offset)); // call data offset
    final long cds = Words.clampedToLong(frame.getStackItem(3 + offset)); // call data size
    final long rdo = Words.clampedToLong(frame.getStackItem(4 + offset)); // return data offset
    final long rdl = Words.clampedToLong(frame.getStackItem(5 + offset)); // return data size

    final long memSize = frame.memoryWordSize();
    final long memSizeCallData = frame.calculateMemoryExpansion(cdo, cds);
    final long memSizeReturnData = frame.calculateMemoryExpansion(rdo, rdl);
    final long maybeNewMemSize = max(memSizeReturnData, memSizeCallData);

    if (memSize < maybeNewMemSize) {
      // computing the "linear" portion of CREATE2 memory expansion cost
      final long G_mem = GasConstants.G_MEMORY.cost();
      final long squareCurrent = (memSize * memSize) >> 9;
      final long squareUpdated = (maybeNewMemSize * maybeNewMemSize) >> 9;
      gasMxp += G_mem * (maybeNewMemSize - memSize) + (squareUpdated - squareCurrent);
    }
    return gasMxp;
  }

  private void traceChunks(StpChunk chunk, int stamp, Trace trace) {
    if (chunk.opCode().isCreate()) {
      traceCreate(chunk, stamp, trace);
    } else {
      traceCall(chunk, stamp, trace);
    }
  }

  private long getGDiff(StpChunk chunk) {
    Preconditions.checkArgument(!chunk.oogx());
    return chunk.gasActual() - chunk.gasPrelim();
  }

  private long getGDiffOver64(StpChunk chunk) {
    return getGDiff(chunk) / 64;
  }

  private long get63of64GDiff(StpChunk chunk) {
    return getGDiff(chunk) - getGDiffOver64(chunk);
  }

  private void traceCreate(StpChunk chunk, int stamp, Trace trace) {
    final int ctMax = ctMax(chunk);
    final long gasOopkt = chunk.oogx() ? 0 : get63of64GDiff(chunk);

    for (int ct = 0; ct <= ctMax; ct++) {
      trace
          .stamp(Bytes.ofUnsignedInt(stamp))
          .ct(Bytes.of(ct))
          .ctMax(Bytes.of(ctMax))
          .instruction(UnsignedByte.of(chunk.opCode().byteValue()))
          .isCreate(chunk.opCode() == OpCode.CREATE)
          .isCreate2(chunk.opCode() == OpCode.CREATE2)
          .isCall(false)
          .isCallcode(false)
          .isDelegatecall(false)
          .isStaticcall(false)
          .gasHi(Bytes.EMPTY)
          .gasLo(Bytes.EMPTY)
          .valHi(chunk.value().slice(0, 16))
          .valLo(chunk.value().slice(16, 16))
          .exists(false) // TODO document this
          .warm(false) // TODO document this
          .outOfGasException(chunk.oogx())
          .gasActual(Bytes.ofUnsignedLong(chunk.gasActual()))
          .gasMxp(Bytes.ofUnsignedLong(chunk.gasMxp()))
          .gasUpfront(Bytes.ofUnsignedLong(chunk.gasPrelim()))
          .gasOopkt(Bytes.ofUnsignedLong(gasOopkt))
          .gasStipend(Bytes.EMPTY)
          .arg1Hi(Bytes.EMPTY);

      switch (ct) {
        case 0 -> trace
            .arg1Lo(Bytes.ofUnsignedLong(chunk.gasActual()))
            .arg2Lo(Bytes.EMPTY)
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.EMPTY) // we REQUIRE that the currently available gas is nonnegative
            .wcpFlag(true)
            .modFlag(false)
            .validateRow();
        case 1 -> trace
            .arg1Lo(Bytes.ofUnsignedLong(chunk.gasActual()))
            .arg2Lo(Bytes.ofUnsignedLong(chunk.gasPrelim()))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.of(chunk.oogx() ? 1 : 0))
            .wcpFlag(true)
            .modFlag(false)
            .validateRow();
        case 2 -> trace
            .arg1Lo(Bytes.ofUnsignedLong(getGDiff(chunk)))
            .arg2Lo(Bytes.of(64))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.DIV.byteValue()))
            .resLo(Bytes.ofUnsignedLong(getGDiffOver64(chunk)))
            .wcpFlag(false)
            .modFlag(true)
            .validateRow();
        default -> throw new IllegalArgumentException("counter too big, should be <=" + ctMax);
      }
    }
  }

  private void traceCall(StpChunk chunk, int stamp, Trace trace) {
    final int ctMax = ctMax(chunk);
    final long gasStipend =
        (!chunk.oogx() && callCanTransferValue(chunk.opCode()) && !chunk.value().isZero())
            ? GasConstants.G_CALL_STIPEND.cost()
            : 0;
    final Bytes gasOopkt =
        chunk.oogx()
            ? Bytes.EMPTY
            : bigIntegerToBytes(
                chunk
                    .gas()
                    .orElseThrow()
                    .toUnsignedBigInteger()
                    .min(BigInteger.valueOf(get63of64GDiff(chunk))));

    for (int ct = 0; ct <= ctMax; ct++) {
      trace
          .stamp(Bytes.ofUnsignedInt(stamp))
          .ct(Bytes.of(ct))
          .ctMax(Bytes.of(ctMax))
          .instruction(UnsignedByte.of(chunk.opCode().byteValue()))
          .isCreate(false)
          .isCreate2(false)
          .isCall(chunk.opCode() == OpCode.CALL)
          .isCallcode(chunk.opCode() == OpCode.CALLCODE)
          .isDelegatecall(chunk.opCode() == OpCode.DELEGATECALL)
          .isStaticcall(chunk.opCode() == OpCode.STATICCALL)
          .gasHi(chunk.gas().orElseThrow().slice(0, 16))
          .gasLo(chunk.gas().orElseThrow().slice(16))
          .valHi(chunk.value().slice(0, 16))
          .valLo(chunk.value().slice(16))
          .exists(chunk.toExists().orElseThrow())
          .warm(chunk.toWarm().orElseThrow())
          .outOfGasException(chunk.oogx())
          .gasActual(Bytes.ofUnsignedLong(chunk.gasActual()))
          .gasMxp(Bytes.ofUnsignedLong(chunk.gasMxp()))
          .gasUpfront(Bytes.ofUnsignedLong(chunk.gasPrelim()))
          .gasOopkt(gasOopkt)
          .gasStipend(Bytes.ofUnsignedLong(gasStipend));

      switch (ct) {
        case 0 -> trace
            .arg1Hi(Bytes.EMPTY)
            .arg1Lo(Bytes.ofUnsignedLong(chunk.gasActual()))
            .arg2Lo(Bytes.EMPTY)
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.EMPTY) // we REQUIRE that the currently available gas is nonnegative
            .wcpFlag(true)
            .modFlag(false)
            .validateRow();
        case 1 -> trace
            .arg1Hi(chunk.value().slice(0, 16))
            .arg1Lo(chunk.value().slice(16, 16))
            .arg2Lo(Bytes.EMPTY)
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.ISZERO.byteValue()))
            .resLo(Bytes.of(chunk.value().isZero() ? 1 : 0))
            .wcpFlag(callCanTransferValue(chunk.opCode()))
            .modFlag(false)
            .validateRow();
        case 2 -> trace
            .arg1Hi(Bytes.EMPTY)
            .arg1Lo(Bytes.ofUnsignedLong(chunk.gasActual()))
            .arg2Lo(Bytes.ofUnsignedLong(chunk.gasPrelim()))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.of(chunk.oogx() ? 1 : 0))
            .wcpFlag(true)
            .modFlag(false)
            .validateRow();
          // the following rows are only filled in if no out of gas exception
        case 3 -> trace
            .arg1Hi(Bytes.EMPTY)
            .arg1Lo(Bytes.ofUnsignedLong(getGDiff(chunk)))
            .arg2Lo(Bytes.of(64))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.DIV.byteValue()))
            .resLo(Bytes.ofUnsignedLong(getGDiffOver64(chunk)))
            .wcpFlag(false)
            .modFlag(true)
            .validateRow();
        case 4 -> trace
            .arg1Hi(chunk.gas().orElseThrow().slice(0, 16))
            .arg1Lo(chunk.gas().orElseThrow().slice(16, 16))
            .arg2Lo(Bytes.ofUnsignedLong(getGDiff(chunk) - getGDiffOver64(chunk)))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(
                Bytes.of(
                    chunk
                                .gas()
                                .orElseThrow()
                                .toUnsignedBigInteger()
                                .compareTo(BigInteger.valueOf(get63of64GDiff(chunk)))
                            < 0
                        ? 1
                        : 0))
            .wcpFlag(true)
            .modFlag(false)
            .validateRow();
        default -> throw new IllegalArgumentException("counter too big, should be <=" + ctMax);
      }
    }
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {

    final Trace trace = new Trace(buffers);
    int stamp = 0;
    for (StpChunk chunk : chunks) {
      stamp++;
      traceChunks(chunk, stamp, trace);
    }
  }

  private int ctMax(StpChunk chunk) {
    if (chunk.oogx()) {
      if (chunk.opCode().isCreate()) {
        return 1;
      } else {
        return 2;
      }
    } else {
      if (chunk.opCode().isCreate()) {
        return 2;
      } else {
        return 4;
      }
    }
  }
}
