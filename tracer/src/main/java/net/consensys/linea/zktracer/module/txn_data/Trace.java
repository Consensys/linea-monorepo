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

package net.consensys.linea.zktracer.module.txn_data;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public record Trace(
    @JsonProperty("ABS_TX_NUM") List<BigInteger> absTxNum,
    @JsonProperty("ABS_TX_NUM_MAX") List<BigInteger> absTxNumMax,
    @JsonProperty("BASEFEE") List<BigInteger> basefee,
    @JsonProperty("BTC_NUM") List<BigInteger> btcNum,
    @JsonProperty("BTC_NUM_MAX") List<BigInteger> btcNumMax,
    @JsonProperty("CALL_DATA_SIZE") List<BigInteger> callDataSize,
    @JsonProperty("CODE_FRAGMENT_INDEX") List<BigInteger> codeFragmentIndex,
    @JsonProperty("COINBASE_HI") List<BigInteger> coinbaseHi,
    @JsonProperty("COINBASE_LO") List<BigInteger> coinbaseLo,
    @JsonProperty("CT") List<BigInteger> ct,
    @JsonProperty("CUMULATIVE_CONSUMED_GAS") List<BigInteger> cumulativeConsumedGas,
    @JsonProperty("FROM_HI") List<BigInteger> fromHi,
    @JsonProperty("FROM_LO") List<BigInteger> fromLo,
    @JsonProperty("GAS_LIMIT") List<BigInteger> gasLimit,
    @JsonProperty("GAS_PRICE") List<BigInteger> gasPrice,
    @JsonProperty("INIT_CODE_SIZE") List<BigInteger> initCodeSize,
    @JsonProperty("INITIAL_BALANCE") List<BigInteger> initialBalance,
    @JsonProperty("INITIAL_GAS") List<BigInteger> initialGas,
    @JsonProperty("IS_DEP") List<Boolean> isDep,
    @JsonProperty("LEFTOVER_GAS") List<BigInteger> leftoverGas,
    @JsonProperty("NONCE") List<BigInteger> nonce,
    @JsonProperty("OUTGOING_HI") List<BigInteger> outgoingHi,
    @JsonProperty("OUTGOING_LO") List<BigInteger> outgoingLo,
    @JsonProperty("OUTGOING_RLP_TXNRCPT") List<BigInteger> outgoingRlpTxnrcpt,
    @JsonProperty("PHASE_RLP_TXN") List<BigInteger> phaseRlpTxn,
    @JsonProperty("PHASE_RLP_TXNRCPT") List<BigInteger> phaseRlpTxnrcpt,
    @JsonProperty("REFUND_AMOUNT") List<BigInteger> refundAmount,
    @JsonProperty("REFUND_COUNTER") List<BigInteger> refundCounter,
    @JsonProperty("REL_TX_NUM") List<BigInteger> relTxNum,
    @JsonProperty("REL_TX_NUM_MAX") List<BigInteger> relTxNumMax,
    @JsonProperty("REQUIRES_EVM_EXECUTION") List<Boolean> requiresEvmExecution,
    @JsonProperty("STATUS_CODE") List<Boolean> statusCode,
    @JsonProperty("TO_HI") List<BigInteger> toHi,
    @JsonProperty("TO_LO") List<BigInteger> toLo,
    @JsonProperty("TYPE0") List<Boolean> type0,
    @JsonProperty("TYPE1") List<Boolean> type1,
    @JsonProperty("TYPE2") List<Boolean> type2,
    @JsonProperty("VALUE") List<BigInteger> value,
    @JsonProperty("WCP_ARG_ONE_LO") List<BigInteger> wcpArgOneLo,
    @JsonProperty("WCP_ARG_TWO_LO") List<BigInteger> wcpArgTwoLo,
    @JsonProperty("WCP_INST") List<BigInteger> wcpInst,
    @JsonProperty("WCP_RES_LO") List<Boolean> wcpResLo) {
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  public int size() {
    return this.absTxNum.size();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ABS_TX_NUM")
    private final List<BigInteger> absTxNum = new ArrayList<>();

    @JsonProperty("ABS_TX_NUM_MAX")
    private final List<BigInteger> absTxNumMax = new ArrayList<>();

    @JsonProperty("BASEFEE")
    private final List<BigInteger> basefee = new ArrayList<>();

    @JsonProperty("BTC_NUM")
    private final List<BigInteger> btcNum = new ArrayList<>();

    @JsonProperty("BTC_NUM_MAX")
    private final List<BigInteger> btcNumMax = new ArrayList<>();

    @JsonProperty("CALL_DATA_SIZE")
    private final List<BigInteger> callDataSize = new ArrayList<>();

    @JsonProperty("CODE_FRAGMENT_INDEX")
    private final List<BigInteger> codeFragmentIndex = new ArrayList<>();

    @JsonProperty("COINBASE_HI")
    private final List<BigInteger> coinbaseHi = new ArrayList<>();

    @JsonProperty("COINBASE_LO")
    private final List<BigInteger> coinbaseLo = new ArrayList<>();

    @JsonProperty("CT")
    private final List<BigInteger> ct = new ArrayList<>();

    @JsonProperty("CUMULATIVE_CONSUMED_GAS")
    private final List<BigInteger> cumulativeConsumedGas = new ArrayList<>();

    @JsonProperty("FROM_HI")
    private final List<BigInteger> fromHi = new ArrayList<>();

    @JsonProperty("FROM_LO")
    private final List<BigInteger> fromLo = new ArrayList<>();

    @JsonProperty("GAS_LIMIT")
    private final List<BigInteger> gasLimit = new ArrayList<>();

    @JsonProperty("GAS_PRICE")
    private final List<BigInteger> gasPrice = new ArrayList<>();

    @JsonProperty("INIT_CODE_SIZE")
    private final List<BigInteger> initCodeSize = new ArrayList<>();

    @JsonProperty("INITIAL_BALANCE")
    private final List<BigInteger> initialBalance = new ArrayList<>();

    @JsonProperty("INITIAL_GAS")
    private final List<BigInteger> initialGas = new ArrayList<>();

    @JsonProperty("IS_DEP")
    private final List<Boolean> isDep = new ArrayList<>();

    @JsonProperty("LEFTOVER_GAS")
    private final List<BigInteger> leftoverGas = new ArrayList<>();

    @JsonProperty("NONCE")
    private final List<BigInteger> nonce = new ArrayList<>();

    @JsonProperty("OUTGOING_HI")
    private final List<BigInteger> outgoingHi = new ArrayList<>();

    @JsonProperty("OUTGOING_LO")
    private final List<BigInteger> outgoingLo = new ArrayList<>();

    @JsonProperty("OUTGOING_RLP_TXNRCPT")
    private final List<BigInteger> outgoingRlpTxnrcpt = new ArrayList<>();

    @JsonProperty("PHASE_RLP_TXN")
    private final List<BigInteger> phaseRlpTxn = new ArrayList<>();

    @JsonProperty("PHASE_RLP_TXNRCPT")
    private final List<BigInteger> phaseRlpTxnrcpt = new ArrayList<>();

    @JsonProperty("REFUND_AMOUNT")
    private final List<BigInteger> refundAmount = new ArrayList<>();

    @JsonProperty("REFUND_COUNTER")
    private final List<BigInteger> refundCounter = new ArrayList<>();

    @JsonProperty("REL_TX_NUM")
    private final List<BigInteger> relTxNum = new ArrayList<>();

    @JsonProperty("REL_TX_NUM_MAX")
    private final List<BigInteger> relTxNumMax = new ArrayList<>();

    @JsonProperty("REQUIRES_EVM_EXECUTION")
    private final List<Boolean> requiresEvmExecution = new ArrayList<>();

    @JsonProperty("STATUS_CODE")
    private final List<Boolean> statusCode = new ArrayList<>();

    @JsonProperty("TO_HI")
    private final List<BigInteger> toHi = new ArrayList<>();

    @JsonProperty("TO_LO")
    private final List<BigInteger> toLo = new ArrayList<>();

    @JsonProperty("TYPE0")
    private final List<Boolean> type0 = new ArrayList<>();

    @JsonProperty("TYPE1")
    private final List<Boolean> type1 = new ArrayList<>();

    @JsonProperty("TYPE2")
    private final List<Boolean> type2 = new ArrayList<>();

    @JsonProperty("VALUE")
    private final List<BigInteger> value = new ArrayList<>();

    @JsonProperty("WCP_ARG_ONE_LO")
    private final List<BigInteger> wcpArgOneLo = new ArrayList<>();

    @JsonProperty("WCP_ARG_TWO_LO")
    private final List<BigInteger> wcpArgTwoLo = new ArrayList<>();

    @JsonProperty("WCP_INST")
    private final List<BigInteger> wcpInst = new ArrayList<>();

    @JsonProperty("WCP_RES_LO")
    private final List<Boolean> wcpResLo = new ArrayList<>();

    TraceBuilder() {}

    public int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.absTxNum.size();
    }

    public TraceBuilder absTxNum(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ABS_TX_NUM already set");
      } else {
        filled.set(0);
      }

      absTxNum.add(b);

      return this;
    }

    public TraceBuilder absTxNumMax(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ABS_TX_NUM_MAX already set");
      } else {
        filled.set(1);
      }

      absTxNumMax.add(b);

      return this;
    }

    public TraceBuilder basefee(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("BASEFEE already set");
      } else {
        filled.set(2);
      }

      basefee.add(b);

      return this;
    }

    public TraceBuilder btcNum(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("BTC_NUM already set");
      } else {
        filled.set(3);
      }

      btcNum.add(b);

      return this;
    }

    public TraceBuilder btcNumMax(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("BTC_NUM_MAX already set");
      } else {
        filled.set(4);
      }

      btcNumMax.add(b);

      return this;
    }

    public TraceBuilder callDataSize(final BigInteger b) {
      if (filled.get(5)) {
        throw new IllegalStateException("CALL_DATA_SIZE already set");
      } else {
        filled.set(5);
      }

      callDataSize.add(b);

      return this;
    }

    public TraceBuilder codeFragmentIndex(final BigInteger b) {
      if (filled.get(6)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX already set");
      } else {
        filled.set(6);
      }

      codeFragmentIndex.add(b);

      return this;
    }

    public TraceBuilder coinbaseHi(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("COINBASE_HI already set");
      } else {
        filled.set(7);
      }

      coinbaseHi.add(b);

      return this;
    }

    public TraceBuilder coinbaseLo(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("COINBASE_LO already set");
      } else {
        filled.set(8);
      }

      coinbaseLo.add(b);

      return this;
    }

    public TraceBuilder ct(final BigInteger b) {
      if (filled.get(9)) {
        throw new IllegalStateException("CT already set");
      } else {
        filled.set(9);
      }

      ct.add(b);

      return this;
    }

    public TraceBuilder cumulativeConsumedGas(final BigInteger b) {
      if (filled.get(10)) {
        throw new IllegalStateException("CUMULATIVE_CONSUMED_GAS already set");
      } else {
        filled.set(10);
      }

      cumulativeConsumedGas.add(b);

      return this;
    }

    public TraceBuilder fromHi(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("FROM_HI already set");
      } else {
        filled.set(11);
      }

      fromHi.add(b);

      return this;
    }

    public TraceBuilder fromLo(final BigInteger b) {
      if (filled.get(12)) {
        throw new IllegalStateException("FROM_LO already set");
      } else {
        filled.set(12);
      }

      fromLo.add(b);

      return this;
    }

    public TraceBuilder gasLimit(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("GAS_LIMIT already set");
      } else {
        filled.set(13);
      }

      gasLimit.add(b);

      return this;
    }

    public TraceBuilder gasPrice(final BigInteger b) {
      if (filled.get(14)) {
        throw new IllegalStateException("GAS_PRICE already set");
      } else {
        filled.set(14);
      }

      gasPrice.add(b);

      return this;
    }

    public TraceBuilder initCodeSize(final BigInteger b) {
      if (filled.get(17)) {
        throw new IllegalStateException("INIT_CODE_SIZE already set");
      } else {
        filled.set(17);
      }

      initCodeSize.add(b);

      return this;
    }

    public TraceBuilder initialBalance(final BigInteger b) {
      if (filled.get(15)) {
        throw new IllegalStateException("INITIAL_BALANCE already set");
      } else {
        filled.set(15);
      }

      initialBalance.add(b);

      return this;
    }

    public TraceBuilder initialGas(final BigInteger b) {
      if (filled.get(16)) {
        throw new IllegalStateException("INITIAL_GAS already set");
      } else {
        filled.set(16);
      }

      initialGas.add(b);

      return this;
    }

    public TraceBuilder isDep(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("IS_DEP already set");
      } else {
        filled.set(18);
      }

      isDep.add(b);

      return this;
    }

    public TraceBuilder leftoverGas(final BigInteger b) {
      if (filled.get(19)) {
        throw new IllegalStateException("LEFTOVER_GAS already set");
      } else {
        filled.set(19);
      }

      leftoverGas.add(b);

      return this;
    }

    public TraceBuilder nonce(final BigInteger b) {
      if (filled.get(20)) {
        throw new IllegalStateException("NONCE already set");
      } else {
        filled.set(20);
      }

      nonce.add(b);

      return this;
    }

    public TraceBuilder outgoingHi(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("OUTGOING_HI already set");
      } else {
        filled.set(21);
      }

      outgoingHi.add(b);

      return this;
    }

    public TraceBuilder outgoingLo(final BigInteger b) {
      if (filled.get(22)) {
        throw new IllegalStateException("OUTGOING_LO already set");
      } else {
        filled.set(22);
      }

      outgoingLo.add(b);

      return this;
    }

    public TraceBuilder outgoingRlpTxnrcpt(final BigInteger b) {
      if (filled.get(23)) {
        throw new IllegalStateException("OUTGOING_RLP_TXNRCPT already set");
      } else {
        filled.set(23);
      }

      outgoingRlpTxnrcpt.add(b);

      return this;
    }

    public TraceBuilder phaseRlpTxn(final BigInteger b) {
      if (filled.get(24)) {
        throw new IllegalStateException("PHASE_RLP_TXN already set");
      } else {
        filled.set(24);
      }

      phaseRlpTxn.add(b);

      return this;
    }

    public TraceBuilder phaseRlpTxnrcpt(final BigInteger b) {
      if (filled.get(25)) {
        throw new IllegalStateException("PHASE_RLP_TXNRCPT already set");
      } else {
        filled.set(25);
      }

      phaseRlpTxnrcpt.add(b);

      return this;
    }

    public TraceBuilder refundAmount(final BigInteger b) {
      if (filled.get(26)) {
        throw new IllegalStateException("REFUND_AMOUNT already set");
      } else {
        filled.set(26);
      }

      refundAmount.add(b);

      return this;
    }

    public TraceBuilder refundCounter(final BigInteger b) {
      if (filled.get(27)) {
        throw new IllegalStateException("REFUND_COUNTER already set");
      } else {
        filled.set(27);
      }

      refundCounter.add(b);

      return this;
    }

    public TraceBuilder relTxNum(final BigInteger b) {
      if (filled.get(28)) {
        throw new IllegalStateException("REL_TX_NUM already set");
      } else {
        filled.set(28);
      }

      relTxNum.add(b);

      return this;
    }

    public TraceBuilder relTxNumMax(final BigInteger b) {
      if (filled.get(29)) {
        throw new IllegalStateException("REL_TX_NUM_MAX already set");
      } else {
        filled.set(29);
      }

      relTxNumMax.add(b);

      return this;
    }

    public TraceBuilder requiresEvmExecution(final Boolean b) {
      if (filled.get(30)) {
        throw new IllegalStateException("REQUIRES_EVM_EXECUTION already set");
      } else {
        filled.set(30);
      }

      requiresEvmExecution.add(b);

      return this;
    }

    public TraceBuilder statusCode(final Boolean b) {
      if (filled.get(31)) {
        throw new IllegalStateException("STATUS_CODE already set");
      } else {
        filled.set(31);
      }

      statusCode.add(b);

      return this;
    }

    public TraceBuilder toHi(final BigInteger b) {
      if (filled.get(32)) {
        throw new IllegalStateException("TO_HI already set");
      } else {
        filled.set(32);
      }

      toHi.add(b);

      return this;
    }

    public TraceBuilder toLo(final BigInteger b) {
      if (filled.get(33)) {
        throw new IllegalStateException("TO_LO already set");
      } else {
        filled.set(33);
      }

      toLo.add(b);

      return this;
    }

    public TraceBuilder type0(final Boolean b) {
      if (filled.get(34)) {
        throw new IllegalStateException("TYPE0 already set");
      } else {
        filled.set(34);
      }

      type0.add(b);

      return this;
    }

    public TraceBuilder type1(final Boolean b) {
      if (filled.get(35)) {
        throw new IllegalStateException("TYPE1 already set");
      } else {
        filled.set(35);
      }

      type1.add(b);

      return this;
    }

    public TraceBuilder type2(final Boolean b) {
      if (filled.get(36)) {
        throw new IllegalStateException("TYPE2 already set");
      } else {
        filled.set(36);
      }

      type2.add(b);

      return this;
    }

    public TraceBuilder value(final BigInteger b) {
      if (filled.get(37)) {
        throw new IllegalStateException("VALUE already set");
      } else {
        filled.set(37);
      }

      value.add(b);

      return this;
    }

    public TraceBuilder wcpArgOneLo(final BigInteger b) {
      if (filled.get(38)) {
        throw new IllegalStateException("WCP_ARG_ONE_LO already set");
      } else {
        filled.set(38);
      }

      wcpArgOneLo.add(b);

      return this;
    }

    public TraceBuilder wcpArgTwoLo(final BigInteger b) {
      if (filled.get(39)) {
        throw new IllegalStateException("WCP_ARG_TWO_LO already set");
      } else {
        filled.set(39);
      }

      wcpArgTwoLo.add(b);

      return this;
    }

    public TraceBuilder wcpInst(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("WCP_INST already set");
      } else {
        filled.set(40);
      }

      wcpInst.add(b);

      return this;
    }

    public TraceBuilder wcpResLo(final Boolean b) {
      if (filled.get(41)) {
        throw new IllegalStateException("WCP_RES_LO already set");
      } else {
        filled.set(41);
      }

      wcpResLo.add(b);

      return this;
    }

    public TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("ABS_TX_NUM has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ABS_TX_NUM_MAX has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("BASEFEE has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("BTC_NUM has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("BTC_NUM_MAX has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("CALL_DATA_SIZE has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("COINBASE_HI has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("COINBASE_LO has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("CT has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("CUMULATIVE_CONSUMED_GAS has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("FROM_HI has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("FROM_LO has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("GAS_LIMIT has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("GAS_PRICE has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("INIT_CODE_SIZE has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("INITIAL_BALANCE has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("INITIAL_GAS has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("IS_DEP has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("LEFTOVER_GAS has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("NONCE has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("OUTGOING_HI has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("OUTGOING_LO has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("OUTGOING_RLP_TXNRCPT has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("PHASE_RLP_TXN has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("PHASE_RLP_TXNRCPT has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("REFUND_AMOUNT has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("REFUND_COUNTER has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("REL_TX_NUM has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("REL_TX_NUM_MAX has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("REQUIRES_EVM_EXECUTION has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("STATUS_CODE has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("TO_HI has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("TO_LO has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("TYPE0 has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("TYPE1 has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("TYPE2 has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("VALUE has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("WCP_ARG_ONE_LO has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("WCP_ARG_TWO_LO has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("WCP_INST has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("WCP_RES_LO has not been filled");
      }

      filled.clear();

      return this;
    }

    public TraceBuilder fillAndValidateRow() {
      if (!filled.get(0)) {
        absTxNum.add(BigInteger.ZERO);
        this.filled.set(0);
      }
      if (!filled.get(1)) {
        absTxNumMax.add(BigInteger.ZERO);
        this.filled.set(1);
      }
      if (!filled.get(2)) {
        basefee.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(3)) {
        btcNum.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(4)) {
        btcNumMax.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(5)) {
        callDataSize.add(BigInteger.ZERO);
        this.filled.set(5);
      }
      if (!filled.get(6)) {
        codeFragmentIndex.add(BigInteger.ZERO);
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        coinbaseHi.add(BigInteger.ZERO);
        this.filled.set(7);
      }
      if (!filled.get(8)) {
        coinbaseLo.add(BigInteger.ZERO);
        this.filled.set(8);
      }
      if (!filled.get(9)) {
        ct.add(BigInteger.ZERO);
        this.filled.set(9);
      }
      if (!filled.get(10)) {
        cumulativeConsumedGas.add(BigInteger.ZERO);
        this.filled.set(10);
      }
      if (!filled.get(11)) {
        fromHi.add(BigInteger.ZERO);
        this.filled.set(11);
      }
      if (!filled.get(12)) {
        fromLo.add(BigInteger.ZERO);
        this.filled.set(12);
      }
      if (!filled.get(13)) {
        gasLimit.add(BigInteger.ZERO);
        this.filled.set(13);
      }
      if (!filled.get(14)) {
        gasPrice.add(BigInteger.ZERO);
        this.filled.set(14);
      }
      if (!filled.get(17)) {
        initCodeSize.add(BigInteger.ZERO);
        this.filled.set(17);
      }
      if (!filled.get(15)) {
        initialBalance.add(BigInteger.ZERO);
        this.filled.set(15);
      }
      if (!filled.get(16)) {
        initialGas.add(BigInteger.ZERO);
        this.filled.set(16);
      }
      if (!filled.get(18)) {
        isDep.add(false);
        this.filled.set(18);
      }
      if (!filled.get(19)) {
        leftoverGas.add(BigInteger.ZERO);
        this.filled.set(19);
      }
      if (!filled.get(20)) {
        nonce.add(BigInteger.ZERO);
        this.filled.set(20);
      }
      if (!filled.get(21)) {
        outgoingHi.add(BigInteger.ZERO);
        this.filled.set(21);
      }
      if (!filled.get(22)) {
        outgoingLo.add(BigInteger.ZERO);
        this.filled.set(22);
      }
      if (!filled.get(23)) {
        outgoingRlpTxnrcpt.add(BigInteger.ZERO);
        this.filled.set(23);
      }
      if (!filled.get(24)) {
        phaseRlpTxn.add(BigInteger.ZERO);
        this.filled.set(24);
      }
      if (!filled.get(25)) {
        phaseRlpTxnrcpt.add(BigInteger.ZERO);
        this.filled.set(25);
      }
      if (!filled.get(26)) {
        refundAmount.add(BigInteger.ZERO);
        this.filled.set(26);
      }
      if (!filled.get(27)) {
        refundCounter.add(BigInteger.ZERO);
        this.filled.set(27);
      }
      if (!filled.get(28)) {
        relTxNum.add(BigInteger.ZERO);
        this.filled.set(28);
      }
      if (!filled.get(29)) {
        relTxNumMax.add(BigInteger.ZERO);
        this.filled.set(29);
      }
      if (!filled.get(30)) {
        requiresEvmExecution.add(false);
        this.filled.set(30);
      }
      if (!filled.get(31)) {
        statusCode.add(false);
        this.filled.set(31);
      }
      if (!filled.get(32)) {
        toHi.add(BigInteger.ZERO);
        this.filled.set(32);
      }
      if (!filled.get(33)) {
        toLo.add(BigInteger.ZERO);
        this.filled.set(33);
      }
      if (!filled.get(34)) {
        type0.add(false);
        this.filled.set(34);
      }
      if (!filled.get(35)) {
        type1.add(false);
        this.filled.set(35);
      }
      if (!filled.get(36)) {
        type2.add(false);
        this.filled.set(36);
      }
      if (!filled.get(37)) {
        value.add(BigInteger.ZERO);
        this.filled.set(37);
      }
      if (!filled.get(38)) {
        wcpArgOneLo.add(BigInteger.ZERO);
        this.filled.set(38);
      }
      if (!filled.get(39)) {
        wcpArgTwoLo.add(BigInteger.ZERO);
        this.filled.set(39);
      }
      if (!filled.get(40)) {
        wcpInst.add(BigInteger.ZERO);
        this.filled.set(40);
      }
      if (!filled.get(41)) {
        wcpResLo.add(false);
        this.filled.set(41);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          absTxNum,
          absTxNumMax,
          basefee,
          btcNum,
          btcNumMax,
          callDataSize,
          codeFragmentIndex,
          coinbaseHi,
          coinbaseLo,
          ct,
          cumulativeConsumedGas,
          fromHi,
          fromLo,
          gasLimit,
          gasPrice,
          initCodeSize,
          initialBalance,
          initialGas,
          isDep,
          leftoverGas,
          nonce,
          outgoingHi,
          outgoingLo,
          outgoingRlpTxnrcpt,
          phaseRlpTxn,
          phaseRlpTxnrcpt,
          refundAmount,
          refundCounter,
          relTxNum,
          relTxNumMax,
          requiresEvmExecution,
          statusCode,
          toHi,
          toLo,
          type0,
          type1,
          type2,
          value,
          wcpArgOneLo,
          wcpArgTwoLo,
          wcpInst,
          wcpResLo);
    }
  }
}
