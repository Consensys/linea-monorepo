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

package net.consensys.linea.zktracer.module.rlpaddr;

import static net.consensys.linea.zktracer.module.rlpaddr.Trace.LLARGE;
import static net.consensys.linea.zktracer.module.rlpaddr.Trace.RLP_ADDR_RECIPE_2;
import static net.consensys.linea.zktracer.module.rlpaddr.Trace.RLP_PREFIX_INT_SHORT;
import static net.consensys.linea.zktracer.module.rlpaddr.Trace.RLP_PREFIX_LIST_SHORT;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.byteCounting;
import static net.consensys.linea.zktracer.types.AddressUtils.getCreate2Address;
import static net.consensys.linea.zktracer.types.AddressUtils.getCreateAddress;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Conversions.longToUnsignedBigInteger;
import static net.consensys.linea.zktracer.types.Utils.bitDecomposition;
import static net.consensys.linea.zktracer.types.Utils.leftPadTo;
import static net.consensys.linea.zktracer.types.Utils.rightPadTo;
import static org.hyperledger.besu.crypto.Hash.keccak256;
import static org.hyperledger.besu.evm.internal.Words.clampedToLong;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.rlputils.ByteCountAndPowerOutput;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.BitDecOutput;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;

@RequiredArgsConstructor
public class RlpAddr implements Module {
  private static final Bytes CREATE2_SHIFT = Bytes.minimalBytes(Trace.CREATE2_SHIFT);
  private static final Bytes INT_SHORT = Bytes.ofUnsignedShort(RLP_PREFIX_INT_SHORT);
  private static final UnsignedByte BYTES_LLARGE = UnsignedByte.of(LLARGE);

  private final Hub hub;
  private final StackedList<RlpAddrChunk> chunkList = new StackedList<>();

  @Override
  public String moduleKey() {
    return "RLP_ADDR";
  }

  @Override
  public void enterTransaction() {
    this.chunkList.enter();
  }

  @Override
  public void popTransaction() {
    this.chunkList.pop();
  }

  @Override
  public void traceStartTx(WorldView world, Transaction tx) {
    if (tx.getTo().isEmpty()) {
      final Address senderAddress = tx.getSender();
      final long nonce = tx.getNonce();
      RlpAddrChunk chunk =
          new RlpAddrChunk(
              Address.contractAddress(senderAddress, nonce),
              OpCode.CREATE,
              longToUnsignedBigInteger(nonce),
              senderAddress);
      this.chunkList.add(chunk);
    }
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opcode = this.hub.opCode();
    switch (opcode) {
      case CREATE -> {
        final Address currentAddress = frame.getRecipientAddress();
        RlpAddrChunk chunk =
            new RlpAddrChunk(
                getCreateAddress(frame),
                OpCode.CREATE,
                longToUnsignedBigInteger(frame.getWorldUpdater().get(currentAddress).getNonce()),
                currentAddress);
        this.chunkList.add(chunk);
      }
      case CREATE2 -> {
        final long offset = clampedToLong(frame.getStackItem(1));
        final long length = clampedToLong(frame.getStackItem(2));
        final Bytes initCode = frame.shadowReadMemory(offset, length);
        final Bytes32 salt = Bytes32.leftPad(frame.getStackItem(3));
        final Bytes32 hash = keccak256(initCode);

        RlpAddrChunk chunk =
            new RlpAddrChunk(
                getCreate2Address(frame), OpCode.CREATE2, frame.getRecipientAddress(), salt, hash);
        this.chunkList.add(chunk);
      }
    }
  }

  private void traceCreate2(int stamp, RlpAddrChunk chunk, Trace trace) {
    for (int ct = 0; ct < 6; ct++) {
      trace
          .stamp(stamp)
          .recipe(UnsignedByte.of(RLP_ADDR_RECIPE_2))
          .recipe1(false)
          .recipe2(true)
          .depAddrHi(chunk.depAddress().slice(0, 4))
          .depAddrLo(chunk.depAddress().slice(4, LLARGE))
          .addrHi(chunk.address().slice(0, 4).toLong())
          .addrLo(chunk.address().slice(4, LLARGE))
          .saltHi(chunk.salt().orElseThrow().slice(0, LLARGE))
          .saltLo(chunk.salt().orElseThrow().slice(LLARGE, LLARGE))
          .kecHi(chunk.keccak().orElseThrow().slice(0, LLARGE))
          .kecLo(chunk.keccak().orElseThrow().slice(LLARGE, LLARGE))
          .lc(true)
          .index(UnsignedByte.of(ct))
          .counter(UnsignedByte.of(ct));

      switch (ct) {
        case 0 -> {
          trace.limb(
              rightPadTo(Bytes.concatenate(CREATE2_SHIFT, chunk.address().slice(0, 4)), LLARGE));
          trace.nBytes(UnsignedByte.of(5));
        }
        case 1 -> trace.limb(chunk.address().slice(4, LLARGE)).nBytes(BYTES_LLARGE);
        case 2 -> trace.limb(chunk.salt().orElseThrow().slice(0, LLARGE)).nBytes(BYTES_LLARGE);
        case 3 -> trace.limb(chunk.salt().orElseThrow().slice(LLARGE, LLARGE)).nBytes(BYTES_LLARGE);
        case 4 -> trace.limb(chunk.keccak().orElseThrow().slice(0, LLARGE)).nBytes(BYTES_LLARGE);
        case 5 -> trace
            .limb(chunk.keccak().orElseThrow().slice(LLARGE, LLARGE))
            .nBytes(BYTES_LLARGE);
      }

      // Columns unused for Recipe2
      trace
          .nonce(Bytes.EMPTY)
          .byte1(UnsignedByte.ZERO)
          .acc(Bytes.EMPTY)
          .accBytesize(UnsignedByte.ZERO)
          .power(Bytes.EMPTY)
          .bit1(false)
          .bitAcc(UnsignedByte.ZERO)
          .tinyNonZeroNonce(false);

      trace.validateRow();
    }
  }

  private void traceCreate(int stamp, RlpAddrChunk chunk, Trace trace) {
    final int RECIPE1_CT_MAX = 8;
    final BigInteger nonce = chunk.nonce().orElseThrow();

    Bytes nonceShifted = leftPadTo(bigIntegerToBytes(nonce), RECIPE1_CT_MAX);
    Boolean tinyNonZeroNonce = true;
    if (nonce.compareTo(BigInteger.ZERO) == 0 || nonce.compareTo(BigInteger.valueOf(128)) >= 0) {
      tinyNonZeroNonce = false;
    }
    // Compute the BYTESIZE and POWER columns
    int nonceByteSize = bigIntegerToBytes(nonce).size();
    if (nonce.equals(BigInteger.ZERO)) {
      nonceByteSize = 0;
    }
    ByteCountAndPowerOutput byteCounting = byteCounting(nonceByteSize, RECIPE1_CT_MAX);

    // Compute the bit decomposition of the last input's byte
    final byte lastByte = nonceShifted.get(RECIPE1_CT_MAX - 1);
    BitDecOutput bitDecomposition = bitDecomposition(0xff & lastByte, RECIPE1_CT_MAX);

    int size_rlp_nonce = nonceByteSize;
    if (!tinyNonZeroNonce) {
      size_rlp_nonce += 1;
    }

    // Bytes RLP(nonce)
    Bytes rlpNonce;
    if (nonce.compareTo(BigInteger.ZERO) == 0) {
      rlpNonce = INT_SHORT;
    } else {
      if (tinyNonZeroNonce) {
        rlpNonce = bigIntegerToBytes(nonce);
      } else {
        rlpNonce =
            Bytes.concatenate(
                bigIntegerToBytes(
                    BigInteger.valueOf(
                        128 + byteCounting.accByteSizeList().get(RECIPE1_CT_MAX - 1))),
                bigIntegerToBytes(nonce));
      }
    }

    for (int ct = 0; ct < 8; ct++) {
      trace
          .stamp(stamp)
          .recipe(UnsignedByte.of(Trace.RLP_ADDR_RECIPE_1))
          .recipe1(true)
          .recipe2(false)
          .addrHi(chunk.address().slice(0, 4).toLong())
          .addrLo(chunk.address().slice(4, LLARGE))
          .depAddrHi(chunk.depAddress().slice(0, 4))
          .depAddrLo(chunk.depAddress().slice(4, LLARGE))
          .nonce(bigIntegerToBytes(nonce))
          .counter(UnsignedByte.of(ct))
          .byte1(UnsignedByte.of(nonceShifted.get(ct)))
          .acc(nonceShifted.slice(0, ct + 1))
          .accBytesize(UnsignedByte.of(byteCounting.accByteSizeList().get(ct)))
          .power(
              bigIntegerToBytes(byteCounting.powerList().get(ct).divide(BigInteger.valueOf(256))))
          .bit1(bitDecomposition.bitDecList().get(ct))
          .bitAcc(UnsignedByte.of(bitDecomposition.bitAccList().get(ct)))
          .tinyNonZeroNonce(tinyNonZeroNonce);

      switch (ct) {
        case 0, 1, 2, 3 -> trace
            .lc(false)
            .limb(Bytes.EMPTY)
            .nBytes(UnsignedByte.ZERO)
            .index(UnsignedByte.ZERO);
        case 4 -> trace
            .lc(true)
            .limb(
                rightPadTo(
                    bigIntegerToBytes(
                        BigInteger.valueOf(RLP_PREFIX_LIST_SHORT)
                            .add(BigInteger.valueOf(21))
                            .add(BigInteger.valueOf(size_rlp_nonce))),
                    LLARGE))
            .nBytes(UnsignedByte.of(1))
            .index(UnsignedByte.ZERO);
        case 5 -> trace
            .lc(true)
            .limb(
                rightPadTo(
                    Bytes.concatenate(
                        bigIntegerToBytes(BigInteger.valueOf(148)), chunk.address().slice(0, 4)),
                    LLARGE))
            .nBytes(UnsignedByte.of(5))
            .index(UnsignedByte.of(1));
        case 6 -> trace
            .lc(true)
            .limb(chunk.address().slice(4, LLARGE))
            .nBytes(UnsignedByte.of(LLARGE))
            .index(UnsignedByte.of(2));
        case 7 -> trace
            .lc(true)
            .limb(rightPadTo(rlpNonce, LLARGE))
            .nBytes(UnsignedByte.of(size_rlp_nonce))
            .index(UnsignedByte.of(3));
      }

      // Column not used fo recipe 1:
      trace
          .saltHi(Bytes.EMPTY)
          .saltLo(Bytes.EMPTY)
          .kecHi(Bytes.EMPTY)
          .kecLo(Bytes.EMPTY)
          .validateRow();
    }
  }

  private void traceChunks(RlpAddrChunk chunk, int stamp, Trace trace) {
    if (chunk.opCode().equals(OpCode.CREATE)) {
      traceCreate(stamp, chunk, trace);
    } else {
      traceCreate2(stamp, chunk, trace);
    }
  }

  @Override
  public int lineCount() {
    return this.chunkList.lineCount();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    for (int i = 0; i < this.chunkList.size(); i++) {
      traceChunks(chunkList.get(i), i + 1, trace);
    }
  }
}
