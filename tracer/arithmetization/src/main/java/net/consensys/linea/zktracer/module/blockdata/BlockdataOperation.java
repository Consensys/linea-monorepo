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

import static com.google.common.base.Preconditions.checkArgument;
import static net.consensys.linea.zktracer.module.blockdata.Trace.*;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_GT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_ISZERO;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EVM_INST_LT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.GAS_LIMIT_ADJUSTMENT_FACTOR;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.WCP_INST_GEQ;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.WCP_INST_LEQ;
import static net.consensys.linea.zktracer.types.Conversions.booleanToBytes;

import java.math.BigInteger;
import java.util.Arrays;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.plugin.data.BlockHeader;

@Accessors(fluent = true)
@Getter
public class BlockdataOperation extends ModuleOperation {
  private final Hub hub;
  private final Wcp wcp;
  private final Euc euc;
  private final Bytes chainId;
  private final BlockHeader blockHeader;
  private final BlockHeader prevBlockHeader;
  private final Address coinbaseAddress;
  private final EWord POWER_256_20 = EWord.of(BigInteger.ONE.shiftLeft(20 * 8));
  private final EWord POWER_256_6 = EWord.of(BigInteger.ONE.shiftLeft(6 * 8));

  private final boolean firstBlockInConflation;
  private final int ctMax;
  private final OpCode opCode;
  private final long firstBlockNumber;
  private final int relTxMax;
  private final long relBlock;

  private EWord data;
  private EWord[] arg1;
  private EWord[] arg2;
  private Bytes[] res;
  private final UnsignedByte[] exoInst;
  private final boolean[] wcpFlag;
  private final boolean[] eucFlag;

  public BlockdataOperation(
      Hub hub,
      BlockHeader blockHeader,
      BlockHeader prevBlockHeader,
      int relTxMax,
      Wcp wcp,
      Euc euc,
      Bytes chainId,
      OpCode opCode,
      long firstBlockNumber) {
    // Data from blockHeader
    this.hub = hub;
    this.blockHeader = blockHeader;
    this.prevBlockHeader = prevBlockHeader;
    this.coinbaseAddress = hub.coinbaseAddress;

    this.chainId = chainId;
    this.ctMax = ctMax(opCode);
    this.firstBlockNumber = firstBlockNumber;
    this.relTxMax = relTxMax;
    this.relBlock = blockHeader.getNumber() - firstBlockNumber + 1;
    this.firstBlockInConflation = (blockHeader.getNumber() == firstBlockNumber);
    this.wcp = wcp;
    this.euc = euc;
    this.opCode = opCode;

    // Init non-counter constant columns arrays of size ctMax
    this.wcpFlag = new boolean[ctMax];
    this.eucFlag = new boolean[ctMax];
    this.exoInst = new UnsignedByte[ctMax];
    this.arg1 = new EWord[ctMax];
    this.arg2 = new EWord[ctMax];
    this.res = new Bytes[ctMax];
    Arrays.fill(exoInst, UnsignedByte.ZERO);
    Arrays.fill(arg1, EWord.ZERO);
    Arrays.fill(arg2, EWord.ZERO);
    Arrays.fill(res, EWord.ZERO);

    // Handle opcodes
    switch (opCode) {
      case OpCode.COINBASE -> {
        handleCoinbase();
      }
      case OpCode.TIMESTAMP -> {
        handleTimestamp();
      }
      case OpCode.NUMBER -> {
        handleNumber();
      }
      case OpCode.DIFFICULTY -> {
        handleDifficulty();
      }
      case OpCode.GASLIMIT -> {
        handleGasLimit();
      }
      case OpCode.CHAINID -> {
        handleChainId();
      }
      case OpCode.BASEFEE -> {
        handleBaseFee();
      }
    }
  }

  private void handleCoinbase() {
    data = EWord.ofHexString(hub.coinbaseAddress.toHexString());
    // row i
    wcpCallToLT(0, data, POWER_256_20);
  }

  private void handleTimestamp() {
    data = EWord.of(blockHeader.getTimestamp());
    EWord prevData =
        prevBlockHeader == null ? EWord.ZERO : EWord.of(prevBlockHeader.getTimestamp());

    // row i
    wcpCallToLT(0, data, POWER_256_6);

    // row i + 1
    wcpCallToGT(1, data, prevData);
  }

  private void handleNumber() {
    data = EWord.of(blockHeader.getNumber());

    wcpCallToISZERO(0, EWord.of(firstBlockNumber));

    // row i
    if (firstBlockInConflation) {
      wcpCallToLT(1, data, POWER_256_6);
    }
  }

  private void handleDifficulty() {
    data = EWord.of(blockHeader.getDifficulty().getAsBigInteger());

    // row i
    wcpCallToGEQ(0, data, EWord.ZERO);
  }

  private void handleGasLimit() {
    data = EWord.of(blockHeader.getGasLimit());

    // row i
    // comparison to minimum
    wcpCallToGEQ(0, data, EWord.of(GAS_LIMIT_MINIMUM));

    // row i + 1
    // comparison to maximum
    wcpCallToLEQ(1, data, EWord.of(Bytes.ofUnsignedLong(GAS_LIMIT_MAXIMUM)));

    if (!firstBlockInConflation) {
      EWord prevGasLimit = EWord.of(prevBlockHeader.getGasLimit());
      // row i + 2
      Bytes maxDeviation = eucCall(2, prevGasLimit, EWord.of(GAS_LIMIT_ADJUSTMENT_FACTOR));
      // row i + 3
      BigInteger safeGasLimitUpperBound =
          prevGasLimit.getAsBigInteger().add(maxDeviation.toUnsignedBigInteger());
      wcpCallToLT(3, data, EWord.of(safeGasLimitUpperBound));

      // row i + 4
      wcpCallToGT(
          4,
          data,
          EWord.of(prevGasLimit.toLong() - maxDeviation.toLong())); // TODO: double check this
    }
  }

  private void handleChainId() {
    data = EWord.of(chainId);

    // row i
    wcpCallToGEQ(0, data, EWord.ZERO);
  }

  private void handleBaseFee() {
    data = EWord.of(blockHeader.getBaseFee().get().getAsBigInteger());

    // row i
    wcpCallToGEQ(0, data, EWord.ZERO);
  }

  @Override
  protected int computeLineCount() {
    return ctMax;
  }

  public void trace(Trace trace) {
    for (short ct = 0; ct < ctMax; ct++) {
      trace
          .iomf(true)
          .ctMax(ctMax - 1)
          .ct(ct)
          .isCoinbase(opCode == OpCode.COINBASE)
          .isTimestamp(opCode == OpCode.TIMESTAMP)
          .isNumber(opCode == OpCode.NUMBER)
          .isDifficulty(opCode == OpCode.DIFFICULTY)
          .isGaslimit(opCode == OpCode.GASLIMIT)
          .isChainid(opCode == OpCode.CHAINID)
          .isBasefee(opCode == OpCode.BASEFEE)
          .inst(UnsignedByte.of(opCode.byteValue()))
          .coinbaseHi(coinbaseAddress.slice(0, 4).toLong())
          .coinbaseLo(coinbaseAddress.slice(4, LLARGE))
          .blockGasLimit(Bytes.ofUnsignedLong(blockHeader.getGasLimit()))
          .basefee(
              Bytes.ofUnsignedLong(blockHeader.getBaseFee().get().getAsBigInteger().longValue()))
          .firstBlockNumber(firstBlockNumber)
          .relBlock((short) relBlock)
          .relTxNumMax((short) relTxMax)
          .dataHi(data.hi())
          .dataLo(data.lo())
          .arg1Hi(arg1[ct].hi())
          .arg1Lo(arg1[ct].lo())
          .arg2Hi(arg2[ct].hi())
          .arg2Lo(arg2[ct].lo())
          .res(res[ct])
          .exoInst(exoInst[ct])
          .wcpFlag(wcpFlag[ct])
          .eucFlag(eucFlag[ct]);

      trace.validateRow();
    }
  }

  // Module call macros
  private boolean wcpCallTo(int w, EWord arg1, EWord arg2, int inst) {
    checkArgument(arg1.bitLength() / 8 <= 32);
    checkArgument(arg2.bitLength() / 8 <= 32);

    this.arg1[w] = arg1;
    this.arg2[w] = arg2;

    final boolean r;
    r =
        switch (inst) {
          case EVM_INST_LT -> wcp.callLT(arg1, arg2);
          case EVM_INST_GT -> wcp.callGT(arg1, arg2);
          case WCP_INST_LEQ -> wcp.callLEQ(arg1, arg2);
          case WCP_INST_GEQ -> wcp.callGEQ(arg1, arg2);
          case EVM_INST_ISZERO -> wcp.callISZERO(arg1);
          default -> throw new IllegalStateException("Unexpected value: " + inst);
        };
    res[w] = booleanToBytes(r);

    exoInst[w] = UnsignedByte.of(inst);

    wcpFlag[w] = true;
    eucFlag[w] = false;

    return r;
  }

  private boolean wcpCallToLT(int w, EWord arg1, EWord arg2) {
    return wcpCallTo(w, arg1, arg2, EVM_INST_LT);
  }

  private boolean wcpCallToGT(int w, EWord arg1, EWord arg2) {
    return wcpCallTo(w, arg1, arg2, EVM_INST_GT);
  }

  private boolean wcpCallToLEQ(int w, EWord arg1, EWord arg2) {
    return wcpCallTo(w, arg1, arg2, WCP_INST_LEQ);
  }

  private boolean wcpCallToGEQ(int w, EWord arg1, EWord arg2) {
    return wcpCallTo(w, arg1, arg2, WCP_INST_GEQ);
  }

  private boolean wcpCallToISZERO(int w, EWord arg1) {
    return wcpCallTo(w, arg1, EWord.ZERO, EVM_INST_ISZERO);
  }

  private Bytes eucCall(int w, EWord arg1, EWord arg2) {
    checkArgument(arg1.bitLength() / 8 <= 16);
    checkArgument(arg2.bitLength() / 8 <= 16);

    this.arg1[w] = arg1;
    this.arg2[w] = arg2;

    res[w] = euc.callEUC(arg1, arg2).quotient();

    wcpFlag[w] = false;
    eucFlag[w] = true;

    return res[w];
  }

  private int ctMax(OpCode opCode) {
    switch (opCode) {
      case OpCode.COINBASE -> {
        return nROWS_CB;
      }
      case OpCode.TIMESTAMP -> {
        return nROWS_TS;
      }
      case OpCode.NUMBER -> {
        return nROWS_NB;
      }
      case OpCode.DIFFICULTY -> {
        return nROWS_DF;
      }
      case OpCode.GASLIMIT -> {
        return nROWS_GL;
      }
      case OpCode.CHAINID -> {
        return nROWS_ID;
      }
      case OpCode.BASEFEE -> {
        return nROWS_BF;
      }
      default -> {
        return nROWS_DEPTH;
      }
    }
  }
}
