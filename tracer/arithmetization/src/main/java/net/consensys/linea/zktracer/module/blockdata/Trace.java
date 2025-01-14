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

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

/**
 * WARNING: This code is generated automatically.
 *
 * <p>Any modifications to this code may be overwritten and could lead to unexpected behavior.
 * Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public class Trace {
  public static final int nROWS_BF = 0x1;
  public static final int nROWS_CB = 0x1;
  public static final int nROWS_DEPTH = 0xd;
  public static final int nROWS_DF = 0x1;
  public static final int nROWS_GL = 0x5;
  public static final int nROWS_ID = 0x1;
  public static final int nROWS_NB = 0x2;
  public static final int nROWS_TS = 0x2;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer arg1Hi;
  private final MappedByteBuffer arg1Lo;
  private final MappedByteBuffer arg2Hi;
  private final MappedByteBuffer arg2Lo;
  private final MappedByteBuffer basefee;
  private final MappedByteBuffer blockGasLimit;
  private final MappedByteBuffer coinbaseHi;
  private final MappedByteBuffer coinbaseLo;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer ctMax;
  private final MappedByteBuffer dataHi;
  private final MappedByteBuffer dataLo;
  private final MappedByteBuffer eucFlag;
  private final MappedByteBuffer exoInst;
  private final MappedByteBuffer firstBlockNumber;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer iomf;
  private final MappedByteBuffer isBasefee;
  private final MappedByteBuffer isChainid;
  private final MappedByteBuffer isCoinbase;
  private final MappedByteBuffer isDifficulty;
  private final MappedByteBuffer isGaslimit;
  private final MappedByteBuffer isNumber;
  private final MappedByteBuffer isTimestamp;
  private final MappedByteBuffer relBlock;
  private final MappedByteBuffer relTxNumMax;
  private final MappedByteBuffer res;
  private final MappedByteBuffer wcpFlag;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("blockdata.ARG_1_HI", 16, length));
      headers.add(new ColumnHeader("blockdata.ARG_1_LO", 16, length));
      headers.add(new ColumnHeader("blockdata.ARG_2_HI", 16, length));
      headers.add(new ColumnHeader("blockdata.ARG_2_LO", 16, length));
      headers.add(new ColumnHeader("blockdata.BASEFEE", 8, length));
      headers.add(new ColumnHeader("blockdata.BLOCK_GAS_LIMIT", 8, length));
      headers.add(new ColumnHeader("blockdata.COINBASE_HI", 4, length));
      headers.add(new ColumnHeader("blockdata.COINBASE_LO", 16, length));
      headers.add(new ColumnHeader("blockdata.CT", 1, length));
      headers.add(new ColumnHeader("blockdata.CT_MAX", 1, length));
      headers.add(new ColumnHeader("blockdata.DATA_HI", 16, length));
      headers.add(new ColumnHeader("blockdata.DATA_LO", 16, length));
      headers.add(new ColumnHeader("blockdata.EUC_FLAG", 1, length));
      headers.add(new ColumnHeader("blockdata.EXO_INST", 1, length));
      headers.add(new ColumnHeader("blockdata.FIRST_BLOCK_NUMBER", 6, length));
      headers.add(new ColumnHeader("blockdata.INST", 1, length));
      headers.add(new ColumnHeader("blockdata.IOMF", 1, length));
      headers.add(new ColumnHeader("blockdata.IS_BASEFEE", 1, length));
      headers.add(new ColumnHeader("blockdata.IS_CHAINID", 1, length));
      headers.add(new ColumnHeader("blockdata.IS_COINBASE", 1, length));
      headers.add(new ColumnHeader("blockdata.IS_DIFFICULTY", 1, length));
      headers.add(new ColumnHeader("blockdata.IS_GASLIMIT", 1, length));
      headers.add(new ColumnHeader("blockdata.IS_NUMBER", 1, length));
      headers.add(new ColumnHeader("blockdata.IS_TIMESTAMP", 1, length));
      headers.add(new ColumnHeader("blockdata.REL_BLOCK", 2, length));
      headers.add(new ColumnHeader("blockdata.REL_TX_NUM_MAX", 2, length));
      headers.add(new ColumnHeader("blockdata.RES", 16, length));
      headers.add(new ColumnHeader("blockdata.WCP_FLAG", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.arg1Hi = buffers.get(0);
    this.arg1Lo = buffers.get(1);
    this.arg2Hi = buffers.get(2);
    this.arg2Lo = buffers.get(3);
    this.basefee = buffers.get(4);
    this.blockGasLimit = buffers.get(5);
    this.coinbaseHi = buffers.get(6);
    this.coinbaseLo = buffers.get(7);
    this.ct = buffers.get(8);
    this.ctMax = buffers.get(9);
    this.dataHi = buffers.get(10);
    this.dataLo = buffers.get(11);
    this.eucFlag = buffers.get(12);
    this.exoInst = buffers.get(13);
    this.firstBlockNumber = buffers.get(14);
    this.inst = buffers.get(15);
    this.iomf = buffers.get(16);
    this.isBasefee = buffers.get(17);
    this.isChainid = buffers.get(18);
    this.isCoinbase = buffers.get(19);
    this.isDifficulty = buffers.get(20);
    this.isGaslimit = buffers.get(21);
    this.isNumber = buffers.get(22);
    this.isTimestamp = buffers.get(23);
    this.relBlock = buffers.get(24);
    this.relTxNumMax = buffers.get(25);
    this.res = buffers.get(26);
    this.wcpFlag = buffers.get(27);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace arg1Hi(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("blockdata.ARG_1_HI already set");
    } else {
      filled.set(0);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockdata.ARG_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace arg1Lo(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("blockdata.ARG_1_LO already set");
    } else {
      filled.set(1);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockdata.ARG_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace arg2Hi(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("blockdata.ARG_2_HI already set");
    } else {
      filled.set(2);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockdata.ARG_2_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace arg2Lo(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("blockdata.ARG_2_LO already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockdata.ARG_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace basefee(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("blockdata.BASEFEE already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("blockdata.BASEFEE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { basefee.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { basefee.put(bs.get(j)); }

    return this;
  }

  public Trace blockGasLimit(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("blockdata.BLOCK_GAS_LIMIT already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("blockdata.BLOCK_GAS_LIMIT has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { blockGasLimit.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockGasLimit.put(bs.get(j)); }

    return this;
  }

  public Trace coinbaseHi(final long b) {
    if (filled.get(6)) {
      throw new IllegalStateException("blockdata.COINBASE_HI already set");
    } else {
      filled.set(6);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("blockdata.COINBASE_HI has invalid value (" + b + ")"); }
    coinbaseHi.put((byte) (b >> 24));
    coinbaseHi.put((byte) (b >> 16));
    coinbaseHi.put((byte) (b >> 8));
    coinbaseHi.put((byte) b);


    return this;
  }

  public Trace coinbaseLo(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("blockdata.COINBASE_LO already set");
    } else {
      filled.set(7);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockdata.COINBASE_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { coinbaseLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { coinbaseLo.put(bs.get(j)); }

    return this;
  }

  public Trace ct(final long b) {
    if (filled.get(8)) {
      throw new IllegalStateException("blockdata.CT already set");
    } else {
      filled.set(8);
    }

    if(b >= 8L) { throw new IllegalArgumentException("blockdata.CT has invalid value (" + b + ")"); }
    ct.put((byte) b);


    return this;
  }

  public Trace ctMax(final long b) {
    if (filled.get(9)) {
      throw new IllegalStateException("blockdata.CT_MAX already set");
    } else {
      filled.set(9);
    }

    if(b >= 8L) { throw new IllegalArgumentException("blockdata.CT_MAX has invalid value (" + b + ")"); }
    ctMax.put((byte) b);


    return this;
  }

  public Trace dataHi(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("blockdata.DATA_HI already set");
    } else {
      filled.set(10);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockdata.DATA_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { dataHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { dataHi.put(bs.get(j)); }

    return this;
  }

  public Trace dataLo(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("blockdata.DATA_LO already set");
    } else {
      filled.set(11);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockdata.DATA_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { dataLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { dataLo.put(bs.get(j)); }

    return this;
  }

  public Trace eucFlag(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("blockdata.EUC_FLAG already set");
    } else {
      filled.set(12);
    }

    eucFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoInst(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("blockdata.EXO_INST already set");
    } else {
      filled.set(13);
    }

    exoInst.put(b.toByte());

    return this;
  }

  public Trace firstBlockNumber(final long b) {
    if (filled.get(14)) {
      throw new IllegalStateException("blockdata.FIRST_BLOCK_NUMBER already set");
    } else {
      filled.set(14);
    }

    if(b >= 281474976710656L) { throw new IllegalArgumentException("blockdata.FIRST_BLOCK_NUMBER has invalid value (" + b + ")"); }
    firstBlockNumber.put((byte) (b >> 40));
    firstBlockNumber.put((byte) (b >> 32));
    firstBlockNumber.put((byte) (b >> 24));
    firstBlockNumber.put((byte) (b >> 16));
    firstBlockNumber.put((byte) (b >> 8));
    firstBlockNumber.put((byte) b);


    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("blockdata.INST already set");
    } else {
      filled.set(15);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace iomf(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("blockdata.IOMF already set");
    } else {
      filled.set(16);
    }

    iomf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isBasefee(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("blockdata.IS_BASEFEE already set");
    } else {
      filled.set(17);
    }

    isBasefee.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isChainid(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("blockdata.IS_CHAINID already set");
    } else {
      filled.set(18);
    }

    isChainid.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCoinbase(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("blockdata.IS_COINBASE already set");
    } else {
      filled.set(19);
    }

    isCoinbase.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isDifficulty(final Boolean b) {
    if (filled.get(20)) {
      throw new IllegalStateException("blockdata.IS_DIFFICULTY already set");
    } else {
      filled.set(20);
    }

    isDifficulty.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isGaslimit(final Boolean b) {
    if (filled.get(21)) {
      throw new IllegalStateException("blockdata.IS_GASLIMIT already set");
    } else {
      filled.set(21);
    }

    isGaslimit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isNumber(final Boolean b) {
    if (filled.get(22)) {
      throw new IllegalStateException("blockdata.IS_NUMBER already set");
    } else {
      filled.set(22);
    }

    isNumber.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isTimestamp(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("blockdata.IS_TIMESTAMP already set");
    } else {
      filled.set(23);
    }

    isTimestamp.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace relBlock(final long b) {
    if (filled.get(24)) {
      throw new IllegalStateException("blockdata.REL_BLOCK already set");
    } else {
      filled.set(24);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("blockdata.REL_BLOCK has invalid value (" + b + ")"); }
    relBlock.put((byte) (b >> 8));
    relBlock.put((byte) b);


    return this;
  }

  public Trace relTxNumMax(final long b) {
    if (filled.get(25)) {
      throw new IllegalStateException("blockdata.REL_TX_NUM_MAX already set");
    } else {
      filled.set(25);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("blockdata.REL_TX_NUM_MAX has invalid value (" + b + ")"); }
    relTxNumMax.put((byte) (b >> 8));
    relTxNumMax.put((byte) b);


    return this;
  }

  public Trace res(final Bytes b) {
    if (filled.get(26)) {
      throw new IllegalStateException("blockdata.RES already set");
    } else {
      filled.set(26);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockdata.RES has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { res.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { res.put(bs.get(j)); }

    return this;
  }

  public Trace wcpFlag(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("blockdata.WCP_FLAG already set");
    } else {
      filled.set(27);
    }

    wcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("blockdata.ARG_1_HI has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("blockdata.ARG_1_LO has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("blockdata.ARG_2_HI has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("blockdata.ARG_2_LO has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("blockdata.BASEFEE has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("blockdata.BLOCK_GAS_LIMIT has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("blockdata.COINBASE_HI has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("blockdata.COINBASE_LO has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("blockdata.CT has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("blockdata.CT_MAX has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("blockdata.DATA_HI has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("blockdata.DATA_LO has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("blockdata.EUC_FLAG has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("blockdata.EXO_INST has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("blockdata.FIRST_BLOCK_NUMBER has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("blockdata.INST has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("blockdata.IOMF has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("blockdata.IS_BASEFEE has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("blockdata.IS_CHAINID has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("blockdata.IS_COINBASE has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("blockdata.IS_DIFFICULTY has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("blockdata.IS_GASLIMIT has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("blockdata.IS_NUMBER has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("blockdata.IS_TIMESTAMP has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("blockdata.REL_BLOCK has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("blockdata.REL_TX_NUM_MAX has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("blockdata.RES has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("blockdata.WCP_FLAG has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      arg1Hi.position(arg1Hi.position() + 16);
    }

    if (!filled.get(1)) {
      arg1Lo.position(arg1Lo.position() + 16);
    }

    if (!filled.get(2)) {
      arg2Hi.position(arg2Hi.position() + 16);
    }

    if (!filled.get(3)) {
      arg2Lo.position(arg2Lo.position() + 16);
    }

    if (!filled.get(4)) {
      basefee.position(basefee.position() + 8);
    }

    if (!filled.get(5)) {
      blockGasLimit.position(blockGasLimit.position() + 8);
    }

    if (!filled.get(6)) {
      coinbaseHi.position(coinbaseHi.position() + 4);
    }

    if (!filled.get(7)) {
      coinbaseLo.position(coinbaseLo.position() + 16);
    }

    if (!filled.get(8)) {
      ct.position(ct.position() + 1);
    }

    if (!filled.get(9)) {
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(10)) {
      dataHi.position(dataHi.position() + 16);
    }

    if (!filled.get(11)) {
      dataLo.position(dataLo.position() + 16);
    }

    if (!filled.get(12)) {
      eucFlag.position(eucFlag.position() + 1);
    }

    if (!filled.get(13)) {
      exoInst.position(exoInst.position() + 1);
    }

    if (!filled.get(14)) {
      firstBlockNumber.position(firstBlockNumber.position() + 6);
    }

    if (!filled.get(15)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(16)) {
      iomf.position(iomf.position() + 1);
    }

    if (!filled.get(17)) {
      isBasefee.position(isBasefee.position() + 1);
    }

    if (!filled.get(18)) {
      isChainid.position(isChainid.position() + 1);
    }

    if (!filled.get(19)) {
      isCoinbase.position(isCoinbase.position() + 1);
    }

    if (!filled.get(20)) {
      isDifficulty.position(isDifficulty.position() + 1);
    }

    if (!filled.get(21)) {
      isGaslimit.position(isGaslimit.position() + 1);
    }

    if (!filled.get(22)) {
      isNumber.position(isNumber.position() + 1);
    }

    if (!filled.get(23)) {
      isTimestamp.position(isTimestamp.position() + 1);
    }

    if (!filled.get(24)) {
      relBlock.position(relBlock.position() + 2);
    }

    if (!filled.get(25)) {
      relTxNumMax.position(relTxNumMax.position() + 2);
    }

    if (!filled.get(26)) {
      res.position(res.position() + 16);
    }

    if (!filled.get(27)) {
      wcpFlag.position(wcpFlag.position() + 1);
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public void build() {
    if (!filled.isEmpty()) {
      throw new IllegalStateException("Cannot build trace with a non-validated row.");
    }
  }
}
