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

package net.consensys.linea.zktracer.module.txn_data;

import java.nio.MappedByteBuffer;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import org.apache.tuweni.bytes.Bytes;

/**
 * WARNING: This code is generated automatically.
 *
 * <p>Any modifications to this code may be overwritten and could lead to unexpected behavior.
 * Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public class Trace {
  static final int CREATE2_SHIFT = 255;
  static final int G_TXDATA_NONZERO = 16;
  static final int G_TXDATA_ZERO = 4;
  static final int INT_LONG = 183;
  static final int INT_SHORT = 128;
  static final int LIST_LONG = 247;
  static final int LIST_SHORT = 192;
  static final int LLARGE = 16;
  static final int LLARGEMO = 15;
  static final int RLPADDR_CONST_RECIPE_1 = 1;
  static final int RLPADDR_CONST_RECIPE_2 = 2;
  static final int RLPRECEIPT_SUBPHASE_ID_ADDR = 53;
  static final int RLPRECEIPT_SUBPHASE_ID_CUMUL_GAS = 3;
  static final int RLPRECEIPT_SUBPHASE_ID_DATA_LIMB = 77;
  static final int RLPRECEIPT_SUBPHASE_ID_DATA_SIZE = 83;
  static final int RLPRECEIPT_SUBPHASE_ID_NO_LOG_ENTRY = 11;
  static final int RLPRECEIPT_SUBPHASE_ID_STATUS_CODE = 2;
  static final int RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE = 65;
  static final int RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA = 96;
  static final int RLPRECEIPT_SUBPHASE_ID_TYPE = 7;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer absTxNum;
  private final MappedByteBuffer absTxNumMax;
  private final MappedByteBuffer basefee;
  private final MappedByteBuffer btcNum;
  private final MappedByteBuffer btcNumMax;
  private final MappedByteBuffer callDataSize;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer coinbaseHi;
  private final MappedByteBuffer coinbaseLo;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer cumulativeConsumedGas;
  private final MappedByteBuffer fromHi;
  private final MappedByteBuffer fromLo;
  private final MappedByteBuffer gasLimit;
  private final MappedByteBuffer gasPrice;
  private final MappedByteBuffer initCodeSize;
  private final MappedByteBuffer initialBalance;
  private final MappedByteBuffer initialGas;
  private final MappedByteBuffer isDep;
  private final MappedByteBuffer leftoverGas;
  private final MappedByteBuffer nonce;
  private final MappedByteBuffer outgoingHi;
  private final MappedByteBuffer outgoingLo;
  private final MappedByteBuffer outgoingRlpTxnrcpt;
  private final MappedByteBuffer phaseRlpTxn;
  private final MappedByteBuffer phaseRlpTxnrcpt;
  private final MappedByteBuffer refundAmount;
  private final MappedByteBuffer refundCounter;
  private final MappedByteBuffer relTxNum;
  private final MappedByteBuffer relTxNumMax;
  private final MappedByteBuffer requiresEvmExecution;
  private final MappedByteBuffer statusCode;
  private final MappedByteBuffer toHi;
  private final MappedByteBuffer toLo;
  private final MappedByteBuffer type0;
  private final MappedByteBuffer type1;
  private final MappedByteBuffer type2;
  private final MappedByteBuffer value;
  private final MappedByteBuffer wcpArgOneLo;
  private final MappedByteBuffer wcpArgTwoLo;
  private final MappedByteBuffer wcpInst;
  private final MappedByteBuffer wcpResLo;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("txnData.ABS_TX_NUM", 32, length),
        new ColumnHeader("txnData.ABS_TX_NUM_MAX", 32, length),
        new ColumnHeader("txnData.BASEFEE", 32, length),
        new ColumnHeader("txnData.BTC_NUM", 32, length),
        new ColumnHeader("txnData.BTC_NUM_MAX", 32, length),
        new ColumnHeader("txnData.CALL_DATA_SIZE", 32, length),
        new ColumnHeader("txnData.CODE_FRAGMENT_INDEX", 32, length),
        new ColumnHeader("txnData.COINBASE_HI", 32, length),
        new ColumnHeader("txnData.COINBASE_LO", 32, length),
        new ColumnHeader("txnData.CT", 32, length),
        new ColumnHeader("txnData.CUMULATIVE_CONSUMED_GAS", 32, length),
        new ColumnHeader("txnData.FROM_HI", 32, length),
        new ColumnHeader("txnData.FROM_LO", 32, length),
        new ColumnHeader("txnData.GAS_LIMIT", 32, length),
        new ColumnHeader("txnData.GAS_PRICE", 32, length),
        new ColumnHeader("txnData.INIT_CODE_SIZE", 32, length),
        new ColumnHeader("txnData.INITIAL_BALANCE", 32, length),
        new ColumnHeader("txnData.INITIAL_GAS", 32, length),
        new ColumnHeader("txnData.IS_DEP", 1, length),
        new ColumnHeader("txnData.LEFTOVER_GAS", 32, length),
        new ColumnHeader("txnData.NONCE", 32, length),
        new ColumnHeader("txnData.OUTGOING_HI", 32, length),
        new ColumnHeader("txnData.OUTGOING_LO", 32, length),
        new ColumnHeader("txnData.OUTGOING_RLP_TXNRCPT", 32, length),
        new ColumnHeader("txnData.PHASE_RLP_TXN", 32, length),
        new ColumnHeader("txnData.PHASE_RLP_TXNRCPT", 32, length),
        new ColumnHeader("txnData.REFUND_AMOUNT", 32, length),
        new ColumnHeader("txnData.REFUND_COUNTER", 32, length),
        new ColumnHeader("txnData.REL_TX_NUM", 32, length),
        new ColumnHeader("txnData.REL_TX_NUM_MAX", 32, length),
        new ColumnHeader("txnData.REQUIRES_EVM_EXECUTION", 1, length),
        new ColumnHeader("txnData.STATUS_CODE", 1, length),
        new ColumnHeader("txnData.TO_HI", 32, length),
        new ColumnHeader("txnData.TO_LO", 32, length),
        new ColumnHeader("txnData.TYPE0", 1, length),
        new ColumnHeader("txnData.TYPE1", 1, length),
        new ColumnHeader("txnData.TYPE2", 1, length),
        new ColumnHeader("txnData.VALUE", 32, length),
        new ColumnHeader("txnData.WCP_ARG_ONE_LO", 32, length),
        new ColumnHeader("txnData.WCP_ARG_TWO_LO", 32, length),
        new ColumnHeader("txnData.WCP_INST", 32, length),
        new ColumnHeader("txnData.WCP_RES_LO", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.absTxNum = buffers.get(0);
    this.absTxNumMax = buffers.get(1);
    this.basefee = buffers.get(2);
    this.btcNum = buffers.get(3);
    this.btcNumMax = buffers.get(4);
    this.callDataSize = buffers.get(5);
    this.codeFragmentIndex = buffers.get(6);
    this.coinbaseHi = buffers.get(7);
    this.coinbaseLo = buffers.get(8);
    this.ct = buffers.get(9);
    this.cumulativeConsumedGas = buffers.get(10);
    this.fromHi = buffers.get(11);
    this.fromLo = buffers.get(12);
    this.gasLimit = buffers.get(13);
    this.gasPrice = buffers.get(14);
    this.initCodeSize = buffers.get(15);
    this.initialBalance = buffers.get(16);
    this.initialGas = buffers.get(17);
    this.isDep = buffers.get(18);
    this.leftoverGas = buffers.get(19);
    this.nonce = buffers.get(20);
    this.outgoingHi = buffers.get(21);
    this.outgoingLo = buffers.get(22);
    this.outgoingRlpTxnrcpt = buffers.get(23);
    this.phaseRlpTxn = buffers.get(24);
    this.phaseRlpTxnrcpt = buffers.get(25);
    this.refundAmount = buffers.get(26);
    this.refundCounter = buffers.get(27);
    this.relTxNum = buffers.get(28);
    this.relTxNumMax = buffers.get(29);
    this.requiresEvmExecution = buffers.get(30);
    this.statusCode = buffers.get(31);
    this.toHi = buffers.get(32);
    this.toLo = buffers.get(33);
    this.type0 = buffers.get(34);
    this.type1 = buffers.get(35);
    this.type2 = buffers.get(36);
    this.value = buffers.get(37);
    this.wcpArgOneLo = buffers.get(38);
    this.wcpArgTwoLo = buffers.get(39);
    this.wcpInst = buffers.get(40);
    this.wcpResLo = buffers.get(41);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absTxNum(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("txnData.ABS_TX_NUM already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      absTxNum.put((byte) 0);
    }
    absTxNum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace absTxNumMax(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("txnData.ABS_TX_NUM_MAX already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      absTxNumMax.put((byte) 0);
    }
    absTxNumMax.put(b.toArrayUnsafe());

    return this;
  }

  public Trace basefee(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("txnData.BASEFEE already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      basefee.put((byte) 0);
    }
    basefee.put(b.toArrayUnsafe());

    return this;
  }

  public Trace btcNum(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("txnData.BTC_NUM already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      btcNum.put((byte) 0);
    }
    btcNum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace btcNumMax(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("txnData.BTC_NUM_MAX already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      btcNumMax.put((byte) 0);
    }
    btcNumMax.put(b.toArrayUnsafe());

    return this;
  }

  public Trace callDataSize(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("txnData.CALL_DATA_SIZE already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      callDataSize.put((byte) 0);
    }
    callDataSize.put(b.toArrayUnsafe());

    return this;
  }

  public Trace codeFragmentIndex(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("txnData.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeFragmentIndex.put((byte) 0);
    }
    codeFragmentIndex.put(b.toArrayUnsafe());

    return this;
  }

  public Trace coinbaseHi(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("txnData.COINBASE_HI already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      coinbaseHi.put((byte) 0);
    }
    coinbaseHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace coinbaseLo(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("txnData.COINBASE_LO already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      coinbaseLo.put((byte) 0);
    }
    coinbaseLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace ct(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("txnData.CT already set");
    } else {
      filled.set(9);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      ct.put((byte) 0);
    }
    ct.put(b.toArrayUnsafe());

    return this;
  }

  public Trace cumulativeConsumedGas(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("txnData.CUMULATIVE_CONSUMED_GAS already set");
    } else {
      filled.set(10);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      cumulativeConsumedGas.put((byte) 0);
    }
    cumulativeConsumedGas.put(b.toArrayUnsafe());

    return this;
  }

  public Trace fromHi(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("txnData.FROM_HI already set");
    } else {
      filled.set(11);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      fromHi.put((byte) 0);
    }
    fromHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace fromLo(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("txnData.FROM_LO already set");
    } else {
      filled.set(12);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      fromLo.put((byte) 0);
    }
    fromLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gasLimit(final Bytes b) {
    if (filled.get(13)) {
      throw new IllegalStateException("txnData.GAS_LIMIT already set");
    } else {
      filled.set(13);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasLimit.put((byte) 0);
    }
    gasLimit.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gasPrice(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("txnData.GAS_PRICE already set");
    } else {
      filled.set(14);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasPrice.put((byte) 0);
    }
    gasPrice.put(b.toArrayUnsafe());

    return this;
  }

  public Trace initCodeSize(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("txnData.INIT_CODE_SIZE already set");
    } else {
      filled.set(17);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      initCodeSize.put((byte) 0);
    }
    initCodeSize.put(b.toArrayUnsafe());

    return this;
  }

  public Trace initialBalance(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("txnData.INITIAL_BALANCE already set");
    } else {
      filled.set(15);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      initialBalance.put((byte) 0);
    }
    initialBalance.put(b.toArrayUnsafe());

    return this;
  }

  public Trace initialGas(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("txnData.INITIAL_GAS already set");
    } else {
      filled.set(16);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      initialGas.put((byte) 0);
    }
    initialGas.put(b.toArrayUnsafe());

    return this;
  }

  public Trace isDep(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("txnData.IS_DEP already set");
    } else {
      filled.set(18);
    }

    isDep.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace leftoverGas(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("txnData.LEFTOVER_GAS already set");
    } else {
      filled.set(19);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      leftoverGas.put((byte) 0);
    }
    leftoverGas.put(b.toArrayUnsafe());

    return this;
  }

  public Trace nonce(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("txnData.NONCE already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonce.put((byte) 0);
    }
    nonce.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingHi(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("txnData.OUTGOING_HI already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingHi.put((byte) 0);
    }
    outgoingHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingLo(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("txnData.OUTGOING_LO already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingLo.put((byte) 0);
    }
    outgoingLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingRlpTxnrcpt(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("txnData.OUTGOING_RLP_TXNRCPT already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingRlpTxnrcpt.put((byte) 0);
    }
    outgoingRlpTxnrcpt.put(b.toArrayUnsafe());

    return this;
  }

  public Trace phaseRlpTxn(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("txnData.PHASE_RLP_TXN already set");
    } else {
      filled.set(24);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      phaseRlpTxn.put((byte) 0);
    }
    phaseRlpTxn.put(b.toArrayUnsafe());

    return this;
  }

  public Trace phaseRlpTxnrcpt(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("txnData.PHASE_RLP_TXNRCPT already set");
    } else {
      filled.set(25);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      phaseRlpTxnrcpt.put((byte) 0);
    }
    phaseRlpTxnrcpt.put(b.toArrayUnsafe());

    return this;
  }

  public Trace refundAmount(final Bytes b) {
    if (filled.get(26)) {
      throw new IllegalStateException("txnData.REFUND_AMOUNT already set");
    } else {
      filled.set(26);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      refundAmount.put((byte) 0);
    }
    refundAmount.put(b.toArrayUnsafe());

    return this;
  }

  public Trace refundCounter(final Bytes b) {
    if (filled.get(27)) {
      throw new IllegalStateException("txnData.REFUND_COUNTER already set");
    } else {
      filled.set(27);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      refundCounter.put((byte) 0);
    }
    refundCounter.put(b.toArrayUnsafe());

    return this;
  }

  public Trace relTxNum(final Bytes b) {
    if (filled.get(28)) {
      throw new IllegalStateException("txnData.REL_TX_NUM already set");
    } else {
      filled.set(28);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      relTxNum.put((byte) 0);
    }
    relTxNum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace relTxNumMax(final Bytes b) {
    if (filled.get(29)) {
      throw new IllegalStateException("txnData.REL_TX_NUM_MAX already set");
    } else {
      filled.set(29);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      relTxNumMax.put((byte) 0);
    }
    relTxNumMax.put(b.toArrayUnsafe());

    return this;
  }

  public Trace requiresEvmExecution(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("txnData.REQUIRES_EVM_EXECUTION already set");
    } else {
      filled.set(30);
    }

    requiresEvmExecution.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace statusCode(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("txnData.STATUS_CODE already set");
    } else {
      filled.set(31);
    }

    statusCode.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace toHi(final Bytes b) {
    if (filled.get(32)) {
      throw new IllegalStateException("txnData.TO_HI already set");
    } else {
      filled.set(32);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      toHi.put((byte) 0);
    }
    toHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace toLo(final Bytes b) {
    if (filled.get(33)) {
      throw new IllegalStateException("txnData.TO_LO already set");
    } else {
      filled.set(33);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      toLo.put((byte) 0);
    }
    toLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace type0(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("txnData.TYPE0 already set");
    } else {
      filled.set(34);
    }

    type0.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace type1(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("txnData.TYPE1 already set");
    } else {
      filled.set(35);
    }

    type1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace type2(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("txnData.TYPE2 already set");
    } else {
      filled.set(36);
    }

    type2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace value(final Bytes b) {
    if (filled.get(37)) {
      throw new IllegalStateException("txnData.VALUE already set");
    } else {
      filled.set(37);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      value.put((byte) 0);
    }
    value.put(b.toArrayUnsafe());

    return this;
  }

  public Trace wcpArgOneLo(final Bytes b) {
    if (filled.get(38)) {
      throw new IllegalStateException("txnData.WCP_ARG_ONE_LO already set");
    } else {
      filled.set(38);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      wcpArgOneLo.put((byte) 0);
    }
    wcpArgOneLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace wcpArgTwoLo(final Bytes b) {
    if (filled.get(39)) {
      throw new IllegalStateException("txnData.WCP_ARG_TWO_LO already set");
    } else {
      filled.set(39);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      wcpArgTwoLo.put((byte) 0);
    }
    wcpArgTwoLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace wcpInst(final Bytes b) {
    if (filled.get(40)) {
      throw new IllegalStateException("txnData.WCP_INST already set");
    } else {
      filled.set(40);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      wcpInst.put((byte) 0);
    }
    wcpInst.put(b.toArrayUnsafe());

    return this;
  }

  public Trace wcpResLo(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("txnData.WCP_RES_LO already set");
    } else {
      filled.set(41);
    }

    wcpResLo.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("txnData.ABS_TX_NUM has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("txnData.ABS_TX_NUM_MAX has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("txnData.BASEFEE has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("txnData.BTC_NUM has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("txnData.BTC_NUM_MAX has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("txnData.CALL_DATA_SIZE has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("txnData.CODE_FRAGMENT_INDEX has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("txnData.COINBASE_HI has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("txnData.COINBASE_LO has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("txnData.CT has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("txnData.CUMULATIVE_CONSUMED_GAS has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("txnData.FROM_HI has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("txnData.FROM_LO has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("txnData.GAS_LIMIT has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("txnData.GAS_PRICE has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("txnData.INIT_CODE_SIZE has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("txnData.INITIAL_BALANCE has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("txnData.INITIAL_GAS has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("txnData.IS_DEP has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("txnData.LEFTOVER_GAS has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("txnData.NONCE has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("txnData.OUTGOING_HI has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("txnData.OUTGOING_LO has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("txnData.OUTGOING_RLP_TXNRCPT has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("txnData.PHASE_RLP_TXN has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("txnData.PHASE_RLP_TXNRCPT has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("txnData.REFUND_AMOUNT has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("txnData.REFUND_COUNTER has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("txnData.REL_TX_NUM has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("txnData.REL_TX_NUM_MAX has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("txnData.REQUIRES_EVM_EXECUTION has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("txnData.STATUS_CODE has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("txnData.TO_HI has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("txnData.TO_LO has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("txnData.TYPE0 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("txnData.TYPE1 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("txnData.TYPE2 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("txnData.VALUE has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("txnData.WCP_ARG_ONE_LO has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("txnData.WCP_ARG_TWO_LO has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("txnData.WCP_INST has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("txnData.WCP_RES_LO has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      absTxNum.position(absTxNum.position() + 32);
    }

    if (!filled.get(1)) {
      absTxNumMax.position(absTxNumMax.position() + 32);
    }

    if (!filled.get(2)) {
      basefee.position(basefee.position() + 32);
    }

    if (!filled.get(3)) {
      btcNum.position(btcNum.position() + 32);
    }

    if (!filled.get(4)) {
      btcNumMax.position(btcNumMax.position() + 32);
    }

    if (!filled.get(5)) {
      callDataSize.position(callDataSize.position() + 32);
    }

    if (!filled.get(6)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 32);
    }

    if (!filled.get(7)) {
      coinbaseHi.position(coinbaseHi.position() + 32);
    }

    if (!filled.get(8)) {
      coinbaseLo.position(coinbaseLo.position() + 32);
    }

    if (!filled.get(9)) {
      ct.position(ct.position() + 32);
    }

    if (!filled.get(10)) {
      cumulativeConsumedGas.position(cumulativeConsumedGas.position() + 32);
    }

    if (!filled.get(11)) {
      fromHi.position(fromHi.position() + 32);
    }

    if (!filled.get(12)) {
      fromLo.position(fromLo.position() + 32);
    }

    if (!filled.get(13)) {
      gasLimit.position(gasLimit.position() + 32);
    }

    if (!filled.get(14)) {
      gasPrice.position(gasPrice.position() + 32);
    }

    if (!filled.get(17)) {
      initCodeSize.position(initCodeSize.position() + 32);
    }

    if (!filled.get(15)) {
      initialBalance.position(initialBalance.position() + 32);
    }

    if (!filled.get(16)) {
      initialGas.position(initialGas.position() + 32);
    }

    if (!filled.get(18)) {
      isDep.position(isDep.position() + 1);
    }

    if (!filled.get(19)) {
      leftoverGas.position(leftoverGas.position() + 32);
    }

    if (!filled.get(20)) {
      nonce.position(nonce.position() + 32);
    }

    if (!filled.get(21)) {
      outgoingHi.position(outgoingHi.position() + 32);
    }

    if (!filled.get(22)) {
      outgoingLo.position(outgoingLo.position() + 32);
    }

    if (!filled.get(23)) {
      outgoingRlpTxnrcpt.position(outgoingRlpTxnrcpt.position() + 32);
    }

    if (!filled.get(24)) {
      phaseRlpTxn.position(phaseRlpTxn.position() + 32);
    }

    if (!filled.get(25)) {
      phaseRlpTxnrcpt.position(phaseRlpTxnrcpt.position() + 32);
    }

    if (!filled.get(26)) {
      refundAmount.position(refundAmount.position() + 32);
    }

    if (!filled.get(27)) {
      refundCounter.position(refundCounter.position() + 32);
    }

    if (!filled.get(28)) {
      relTxNum.position(relTxNum.position() + 32);
    }

    if (!filled.get(29)) {
      relTxNumMax.position(relTxNumMax.position() + 32);
    }

    if (!filled.get(30)) {
      requiresEvmExecution.position(requiresEvmExecution.position() + 1);
    }

    if (!filled.get(31)) {
      statusCode.position(statusCode.position() + 1);
    }

    if (!filled.get(32)) {
      toHi.position(toHi.position() + 32);
    }

    if (!filled.get(33)) {
      toLo.position(toLo.position() + 32);
    }

    if (!filled.get(34)) {
      type0.position(type0.position() + 1);
    }

    if (!filled.get(35)) {
      type1.position(type1.position() + 1);
    }

    if (!filled.get(36)) {
      type2.position(type2.position() + 1);
    }

    if (!filled.get(37)) {
      value.position(value.position() + 32);
    }

    if (!filled.get(38)) {
      wcpArgOneLo.position(wcpArgOneLo.position() + 32);
    }

    if (!filled.get(39)) {
      wcpArgTwoLo.position(wcpArgTwoLo.position() + 32);
    }

    if (!filled.get(40)) {
      wcpInst.position(wcpInst.position() + 32);
    }

    if (!filled.get(41)) {
      wcpResLo.position(wcpResLo.position() + 1);
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
