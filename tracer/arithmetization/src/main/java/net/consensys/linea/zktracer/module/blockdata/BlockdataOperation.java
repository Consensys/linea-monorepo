/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.module.blockdata;

import static net.consensys.linea.zktracer.module.blockdata.Trace.MAX_CT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_BASEFEE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_CHAINID;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_COINBASE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_DIFFICULTY;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_GASLIMIT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_NUMBER;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_TIMESTAMP;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LINEA_BASE_FEE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LINEA_BLOCK_GAS_LIMIT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.math.BigInteger;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;

@Accessors(fluent = true)
@Getter
@RequiredArgsConstructor
public class BlockdataOperation extends ModuleOperation {
  private final Address coinbase;
  private final long timestamp;
  private final long absoluteBlockNumber;
  private final BigInteger difficulty;
  private final int relTxMax;

  @Override
  protected int computeLineCount() {
    return MAX_CT + 1;
  }

  public void trace(
      Trace trace, final int relBlock, final long firstBlockNumber, final BigInteger chainId) {
    for (short ct = 0; ct <= MAX_CT; ct++) {
      traceBlockConstant(trace, relBlock, firstBlockNumber);
      traceRowDependant(trace, ct, relBlock, chainId);
      trace.validateRow();
    }
  }

  private void traceRowDependant(
      Trace trace, final short ct, final int relBlock, final BigInteger chainId) {
    trace.ct(ct);

    Bytes32 data;
    switch (ct) {
      case 0 -> {
        data = Bytes32.leftPad(this.coinbase);
        trace.inst(UnsignedByte.of(EVM_INST_COINBASE)).wcpFlag(false);
      }
      case 1 -> {
        data = Bytes32.leftPad(Bytes.ofUnsignedLong(this.timestamp));
        trace.inst(UnsignedByte.of(EVM_INST_TIMESTAMP)).wcpFlag(relBlock != 1);
      }
      case 2 -> {
        data = Bytes32.leftPad(Bytes.ofUnsignedLong(this.absoluteBlockNumber));
        trace.inst(UnsignedByte.of(EVM_INST_NUMBER)).wcpFlag(false);
      }
      case 3 -> {
        data = Bytes32.leftPad(bigIntegerToBytes(this.difficulty));
        trace.inst(UnsignedByte.of(EVM_INST_DIFFICULTY)).wcpFlag(false);
      }
      case 4 -> {
        data = Bytes32.leftPad(Bytes.ofUnsignedLong(LINEA_BLOCK_GAS_LIMIT));
        trace.inst(UnsignedByte.of(EVM_INST_GASLIMIT)).wcpFlag(false);
      }
      case 5 -> {
        data = Bytes32.leftPad(bigIntegerToBytes(chainId));
        trace.inst(UnsignedByte.of(EVM_INST_CHAINID)).wcpFlag(false);
      }
      case 6 -> {
        data = Bytes32.leftPad(Bytes.ofUnsignedLong(LINEA_BASE_FEE));
        trace.inst(UnsignedByte.of(EVM_INST_BASEFEE)).wcpFlag(false);
      }
      default -> throw new IllegalArgumentException(
          String.format("Blockdata max CT is %s, can't write %s", MAX_CT, ct));
    }

    trace
        .dataHi(data.slice(0, LLARGE))
        .dataLo(data.slice(LLARGE, LLARGE))
        .byteHi0(UnsignedByte.of(data.get(0)))
        .byteHi1(UnsignedByte.of(data.get(1)))
        .byteHi2(UnsignedByte.of(data.get(2)))
        .byteHi3(UnsignedByte.of(data.get(3)))
        .byteHi4(UnsignedByte.of(data.get(4)))
        .byteHi5(UnsignedByte.of(data.get(5)))
        .byteHi6(UnsignedByte.of(data.get(6)))
        .byteHi7(UnsignedByte.of(data.get(7)))
        .byteHi8(UnsignedByte.of(data.get(8)))
        .byteHi9(UnsignedByte.of(data.get(9)))
        .byteHi10(UnsignedByte.of(data.get(10)))
        .byteHi11(UnsignedByte.of(data.get(11)))
        .byteHi12(UnsignedByte.of(data.get(12)))
        .byteHi13(UnsignedByte.of(data.get(13)))
        .byteHi14(UnsignedByte.of(data.get(14)))
        .byteHi15(UnsignedByte.of(data.get(15)))
        .byteLo0(UnsignedByte.of(data.get(LLARGE + 0)))
        .byteLo1(UnsignedByte.of(data.get(LLARGE + 1)))
        .byteLo2(UnsignedByte.of(data.get(LLARGE + 2)))
        .byteLo3(UnsignedByte.of(data.get(LLARGE + 3)))
        .byteLo4(UnsignedByte.of(data.get(LLARGE + 4)))
        .byteLo5(UnsignedByte.of(data.get(LLARGE + 5)))
        .byteLo6(UnsignedByte.of(data.get(LLARGE + 6)))
        .byteLo7(UnsignedByte.of(data.get(LLARGE + 7)))
        .byteLo8(UnsignedByte.of(data.get(LLARGE + 8)))
        .byteLo9(UnsignedByte.of(data.get(LLARGE + 9)))
        .byteLo10(UnsignedByte.of(data.get(LLARGE + 10)))
        .byteLo11(UnsignedByte.of(data.get(LLARGE + 11)))
        .byteLo12(UnsignedByte.of(data.get(LLARGE + 12)))
        .byteLo13(UnsignedByte.of(data.get(LLARGE + 13)))
        .byteLo14(UnsignedByte.of(data.get(LLARGE + 14)))
        .byteLo15(UnsignedByte.of(data.get(LLARGE + 15)));
  }

  private void traceBlockConstant(Trace trace, final int relBlock, final long firstBlockNumber) {
    trace
        .firstBlockNumber(firstBlockNumber)
        .relBlock((short) relBlock)
        .relTxNumMax((short) this.relTxMax)
        .coinbaseHi(this.coinbase.slice(0, 4).toLong())
        .coinbaseLo(this.coinbase.slice(4, LLARGE))
        .blockGasLimit(LINEA_BLOCK_GAS_LIMIT)
        .basefee(LINEA_BASE_FEE);
  }
}
