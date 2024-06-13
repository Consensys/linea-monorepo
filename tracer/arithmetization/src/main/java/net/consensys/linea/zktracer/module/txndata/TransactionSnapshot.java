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

package net.consensys.linea.zktracer.module.txndata;

import static net.consensys.linea.zktracer.module.Util.getTxTypeAsInt;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_0;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_1;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_2;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_3;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_4;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_5;
import static net.consensys.linea.zktracer.module.txndata.Trace.GAS_CONST_G_ACCESS_LIST_ADRESS;
import static net.consensys.linea.zktracer.module.txndata.Trace.GAS_CONST_G_ACCESS_LIST_STORAGE;
import static net.consensys.linea.zktracer.module.txndata.Trace.GAS_CONST_G_TRANSACTION;
import static net.consensys.linea.zktracer.module.txndata.Trace.GAS_CONST_G_TX_CREATE;
import static net.consensys.linea.zktracer.module.txndata.Trace.GAS_CONST_G_TX_DATA_NONZERO;
import static net.consensys.linea.zktracer.module.txndata.Trace.GAS_CONST_G_TX_DATA_ZERO;
import static net.consensys.linea.zktracer.module.txndata.Trace.LLARGE;
import static net.consensys.linea.zktracer.module.txndata.Trace.MAX_REFUND_QUOTIENT;
import static net.consensys.linea.zktracer.module.txndata.Trace.NB_ROWS_TYPE_0;
import static net.consensys.linea.zktracer.module.txndata.Trace.NB_ROWS_TYPE_1;
import static net.consensys.linea.zktracer.module.txndata.Trace.NB_ROWS_TYPE_2;
import static net.consensys.linea.zktracer.module.txndata.Trace.RLP_RCPT_SUBPHASE_ID_CUMUL_GAS;
import static net.consensys.linea.zktracer.module.txndata.Trace.RLP_RCPT_SUBPHASE_ID_STATUS_CODE;
import static net.consensys.linea.zktracer.module.txndata.Trace.RLP_RCPT_SUBPHASE_ID_TYPE;
import static net.consensys.linea.zktracer.module.txndata.Trace.TYPE_0_RLP_TXN_PHASE_NUMBER_6;
import static net.consensys.linea.zktracer.module.txndata.Trace.TYPE_1_RLP_TXN_PHASE_NUMBER_6;
import static net.consensys.linea.zktracer.module.txndata.Trace.TYPE_1_RLP_TXN_PHASE_NUMBER_7;
import static net.consensys.linea.zktracer.module.txndata.Trace.TYPE_2_RLP_TXN_PHASE_NUMBER_6;
import static net.consensys.linea.zktracer.module.txndata.Trace.TYPE_2_RLP_TXN_PHASE_NUMBER_7;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Conversions.booleanToInt;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Quantity;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.feemarket.TransactionPriceCalculator;
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.worldstate.WorldView;

/** Gathers all the information required to trace a {@link Transaction} in {@link TxnData}. */
@Accessors(fluent = true)
@Getter
public final class TransactionSnapshot extends ModuleOperation {

  private final Wcp wcp;
  private final Euc euc;

  private final Optional<Wei> baseFee;
  private final Bytes blockGasLimit;

  /** Value moved by the transaction */
  private final BigInteger value;

  /** Sender address */
  private final Address from;

  /** Receiver or contract deployment address */
  private final Address to;

  /** Sender nonce */
  private final long nonce;

  /** Number of addresses to pre-warm */
  private final int prewarmedAddressesCount;

  /** Number of storage slots to pre-warm */
  private final int prewarmedStorageKeysCount;

  /** Whether this transaction is a smart contract deployment */
  private final boolean isDeployment;

  /** Whether this transaction triggers the EVM */
  private final boolean requiresEvmExecution;

  /** The transaction {@link TransactionType} */
  private final TransactionType type;

  /** The sender balance when it sent the transaction */
  private final BigInteger initialSenderBalance;

  /** The payload of the transaction, calldata or initcode */
  private final Bytes payload;

  private final int callDataSize;

  private final long gasLimit;
  private final BigInteger effectiveGasPrice;
  private final Optional<? extends Quantity> maxFeePerGas;
  private final Optional<? extends Quantity> maxPriorityFeePerGas;
  @Setter private boolean status;
  @Setter private long refundCounter;
  @Setter private long leftoverGas;
  @Setter private long effectiveGasRefund;
  @Setter private long cumulativeGasConsumption;

  @Setter private boolean getFullTip;

  private final List<TxnDataComparisonRecord> callsToEucAndWcp;
  private final List<RlptxnOutgoing> valuesToRlptxn;
  private final List<RlptxrcptOutgoing> valuesToRlpTxrcpt;

  // plus one because the last tx of the block has one more row
  private final int MAX_NB_ROWS =
      Math.max(Math.max(NB_ROWS_TYPE_1, NB_ROWS_TYPE_2), NB_ROWS_TYPE_0) + 1;

  public TransactionSnapshot(
      Wcp wcp,
      Euc euc,
      BigInteger value,
      Address from,
      Address to,
      long nonce,
      int prewarmedAddressesCount,
      int prewarmedStorageKeysCount,
      boolean isDeployment,
      boolean requiresEvmExecution,
      TransactionType type,
      BigInteger initialSenderBalance,
      Bytes payload,
      long gasLimit,
      BigInteger effectiveGasPrice,
      Optional<? extends Quantity> maxFeePerGas,
      Optional<? extends Quantity> maxPriorityFeePerGas,
      Bytes blockGasLimit,
      Optional<Wei> baseFee) {

    this.wcp = wcp;
    this.euc = euc;

    this.value = value;
    this.from = from;
    this.to = to;
    this.nonce = nonce;
    this.prewarmedAddressesCount = prewarmedAddressesCount;
    this.prewarmedStorageKeysCount = prewarmedStorageKeysCount;
    this.isDeployment = isDeployment;
    this.requiresEvmExecution = requiresEvmExecution;
    this.type = type;
    this.initialSenderBalance = initialSenderBalance;
    this.payload = payload;
    this.gasLimit = gasLimit;
    this.effectiveGasPrice = effectiveGasPrice;
    this.maxFeePerGas = maxFeePerGas;
    this.maxPriorityFeePerGas = maxPriorityFeePerGas;
    this.callDataSize = this.isDeployment ? 0 : this.payload.size();
    this.callsToEucAndWcp = new ArrayList<>();
    this.valuesToRlptxn = new ArrayList<>();
    this.valuesToRlpTxrcpt = new ArrayList<>();
    this.blockGasLimit = blockGasLimit;
    this.baseFee = baseFee;
  }

  public static TransactionSnapshot fromTransaction(
      Wcp wcp,
      Euc euc,
      Transaction tx,
      WorldView world,
      Optional<Wei> baseFee,
      Bytes blockGasLimit) {

    return new TransactionSnapshot(
        wcp,
        euc,
        tx.getValue().getAsBigInteger(),
        tx.getSender(),
        tx.getTo()
            .map(x -> (Address) x)
            .orElse(Address.contractAddress(tx.getSender(), tx.getNonce())),
        tx.getNonce(),
        tx.getAccessList().map(List::size).orElse(0),
        tx.getAccessList()
            .map(
                accessSet ->
                    accessSet.stream()
                        .mapToInt(accessSetItem -> accessSetItem.storageKeys().size())
                        .sum())
            .orElse(0),
        tx.getTo().isEmpty(),
        tx.getTo().map(world::get).map(AccountState::hasCode).orElse(!tx.getPayload().isEmpty()),
        tx.getType(),
        Optional.ofNullable(tx.getSender())
            .map(world::get)
            .map(x -> x.getBalance().getAsBigInteger())
            .orElse(BigInteger.ZERO),
        tx.getPayload().copy(),
        tx.getGasLimit(),
        computeEffectiveGasPrice(baseFee, tx),
        tx.getMaxFeePerGas(),
        tx.getMaxPriorityFeePerGas(),
        blockGasLimit,
        baseFee);
  }

  // dataCost returns the gas cost of the call data / init code
  public long dataCost() {
    Bytes payload = this.payload();
    long dataCost = 0;
    for (int i = 0; i < payload.size(); i++) {
      dataCost += payload.get(i) == 0 ? GAS_CONST_G_TX_DATA_ZERO : GAS_CONST_G_TX_DATA_NONZERO;
    }
    return dataCost;
  }

  private static BigInteger computeEffectiveGasPrice(Optional<Wei> baseFee, Transaction tx) {
    return switch (tx.getType()) {
      case FRONTIER, ACCESS_LIST -> tx.getGasPrice().get().getAsBigInteger();
      case EIP1559 -> TransactionPriceCalculator.eip1559()
          .price((org.hyperledger.besu.ethereum.core.Transaction) tx, baseFee)
          .getAsBigInteger();
      default -> throw new RuntimeException("transaction type not supported");
    };
  }

  // getData should return either:
  // - call data (message call)
  // - init code (contract creation)
  int maxCounter() {
    return switch (this.type()) {
      case FRONTIER -> NB_ROWS_TYPE_0 - 1;
      case ACCESS_LIST -> NB_ROWS_TYPE_1 - 1;
      case EIP1559 -> NB_ROWS_TYPE_2 - 1;
      default -> throw new RuntimeException("transaction type not supported");
    };
  }

  long getUpfrontGasCost() {
    long initialCost = this.dataCost();

    if (this.isDeployment()) {
      initialCost += GAS_CONST_G_TX_CREATE;
    }

    initialCost += GAS_CONST_G_TRANSACTION;

    if (this.type() != TransactionType.FRONTIER) {
      initialCost += (long) this.prewarmedAddressesCount() * GAS_CONST_G_ACCESS_LIST_ADRESS;
      initialCost += (long) this.prewarmedStorageKeysCount() * GAS_CONST_G_ACCESS_LIST_STORAGE;
    }

    return initialCost;
  }

  public void setRlptxnValues() {
    // i+0
    this.valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_0,
            Bytes.EMPTY,
            Bytes.ofUnsignedInt(getTxTypeAsInt(this.type))));
    // i+1
    this.valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_1,
            isDeployment ? Bytes.EMPTY : this.to.slice(0, 4),
            isDeployment ? Bytes.EMPTY : this.to.slice(4, LLARGE)));

    // i+2
    this.valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_2, Bytes.EMPTY, Bytes.ofUnsignedLong(this.nonce)));

    // i+3
    this.valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_3,
            isDeployment ? Bytes.of(1) : Bytes.EMPTY,
            bigIntegerToBytes(this.value)));

    // i+4
    this.valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_4,
            Bytes.ofUnsignedLong(this.dataCost()),
            Bytes.ofUnsignedLong(this.payload.size())));

    // i+5
    this.valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_5,
            Bytes.EMPTY,
            Bytes.ofUnsignedLong(this.gasLimit)));

    switch (this.type) {
      case FRONTIER -> {
        // i+6
        this.valuesToRlptxn.add(
            RlptxnOutgoing.set(
                (short) TYPE_0_RLP_TXN_PHASE_NUMBER_6,
                Bytes.EMPTY,
                bigIntegerToBytes(this.effectiveGasPrice)));
        for (int i = 7; i < NB_ROWS_TYPE_0 + 1; i++) {
          this.valuesToRlptxn.add(RlptxnOutgoing.empty());
        }
      }
      case ACCESS_LIST -> {
        // i+6
        this.valuesToRlptxn.add(
            RlptxnOutgoing.set(
                (short) TYPE_1_RLP_TXN_PHASE_NUMBER_6,
                Bytes.EMPTY,
                bigIntegerToBytes(this.effectiveGasPrice)));

        // i+7
        this.valuesToRlptxn.add(
            RlptxnOutgoing.set(
                (short) TYPE_1_RLP_TXN_PHASE_NUMBER_7,
                Bytes.ofUnsignedInt(this.prewarmedStorageKeysCount),
                Bytes.ofUnsignedInt(this.prewarmedAddressesCount)));

        for (int i = 8; i < NB_ROWS_TYPE_1 + 1; i++) {
          this.valuesToRlptxn.add(RlptxnOutgoing.empty());
        }
      }

      case EIP1559 -> {
        // i+6
        this.valuesToRlptxn.add(
            RlptxnOutgoing.set(
                (short) TYPE_2_RLP_TXN_PHASE_NUMBER_6,
                bigIntegerToBytes(this.maxPriorityFeePerGas.get().getAsBigInteger()),
                bigIntegerToBytes(this.maxFeePerGas.get().getAsBigInteger())));

        // i+7
        this.valuesToRlptxn.add(
            RlptxnOutgoing.set(
                (short) TYPE_2_RLP_TXN_PHASE_NUMBER_7,
                Bytes.ofUnsignedInt(this.prewarmedStorageKeysCount),
                Bytes.ofUnsignedInt(this.prewarmedAddressesCount)));

        for (int i = 8; i < NB_ROWS_TYPE_2 + 1; i++) {
          this.valuesToRlptxn.add(RlptxnOutgoing.empty());
        }
      }
    }
  }

  public void setRlptxrcptValues() {
    // i+0
    this.valuesToRlpTxrcpt.add(
        RlptxrcptOutgoing.set((short) RLP_RCPT_SUBPHASE_ID_TYPE, getTxTypeAsInt(this.type())));
    // i+1
    this.valuesToRlpTxrcpt.add(
        RlptxrcptOutgoing.set((short) RLP_RCPT_SUBPHASE_ID_STATUS_CODE, booleanToInt(this.status)));
    // i+2
    this.valuesToRlpTxrcpt.add(
        RlptxrcptOutgoing.set(
            (short) RLP_RCPT_SUBPHASE_ID_CUMUL_GAS, this.cumulativeGasConsumption));
    // i+3 to i+MAX_NB_ROWS
    for (int ct = 3; ct < MAX_NB_ROWS; ct++) {
      this.valuesToRlpTxrcpt.add(RlptxrcptOutgoing.emptyValue());
    }
  }

  public void setCallsToEucAndWcp() {
    // i+0
    final Bytes row0arg1 = bigIntegerToBytes(this.initialSenderBalance);
    final BigInteger value = this.value;
    final BigInteger maxFeeShortHand = setOutgoingLoRowPlus6();
    final BigInteger gasLimit = BigInteger.valueOf(this.gasLimit);
    final Bytes row0arg2 = bigIntegerToBytes(value.add(maxFeeShortHand.multiply(gasLimit)));
    wcp.callLT(row0arg1, row0arg2);
    this.callsToEucAndWcp.add(TxnDataComparisonRecord.callToLt(row0arg1, row0arg2, false));

    // i+1
    final Bytes row1arg1 = Bytes.minimalBytes(this.gasLimit);
    final Bytes row1arg2 = Bytes.minimalBytes(this.getUpfrontGasCost());
    wcp.callLT(row1arg1, row1arg2);
    this.callsToEucAndWcp.add(TxnDataComparisonRecord.callToLt(row1arg1, row1arg2, false));

    // i+2
    final Bytes row2arg1 = Bytes.minimalBytes(this.gasLimit - this.leftoverGas);
    final Bytes row2arg2 = Bytes.of(MAX_REFUND_QUOTIENT);
    final Bytes refundLimit = euc.callEUC(row2arg1, row2arg2).quotient();
    this.callsToEucAndWcp.add(TxnDataComparisonRecord.callToEuc(row2arg1, row2arg2, refundLimit));

    // i+3
    final Bytes refundCounterMax = Bytes.minimalBytes(this.refundCounter);
    final boolean getFullRefund = wcp.callLT(refundCounterMax, refundLimit);
    this.callsToEucAndWcp.add(
        TxnDataComparisonRecord.callToLt(refundCounterMax, refundLimit, getFullRefund));

    this.effectiveGasRefund(
        getFullRefund ? leftoverGas + this.refundCounter : leftoverGas + refundLimit.toInt());

    // i+4
    final Bytes row4arg1 = Bytes.minimalBytes(this.payload.size());
    final boolean nonZeroDataSize = wcp.callISZERO(row4arg1);
    this.callsToEucAndWcp.add(TxnDataComparisonRecord.callToIsZero(row4arg1, nonZeroDataSize));

    switch (this.type) {
      case FRONTIER -> {
        for (int i = 5; i < NB_ROWS_TYPE_0; i++) {
          this.callsToEucAndWcp.add(TxnDataComparisonRecord.empty());
        }
      }
      case ACCESS_LIST -> {
        for (int i = 5; i < NB_ROWS_TYPE_1; i++) {
          this.callsToEucAndWcp.add(TxnDataComparisonRecord.empty());
        }
      }
      case EIP1559 -> {
        // i+5
        final Bytes maxFee = bigIntegerToBytes(this.maxFeePerGas.get().getAsBigInteger());
        final Bytes row5arg2 = Bytes.minimalBytes(this.baseFee.get().intValue());
        wcp.callLT(maxFee, row5arg2);
        this.callsToEucAndWcp.add(TxnDataComparisonRecord.callToLt(maxFee, row5arg2, false));

        // i+6
        final Bytes row6arg2 = bigIntegerToBytes(this.maxPriorityFeePerGas.get().getAsBigInteger());
        wcp.callLT(maxFee, row6arg2);
        this.callsToEucAndWcp.add(TxnDataComparisonRecord.callToLt(maxFee, row6arg2, false));

        // i+7
        final Bytes row7arg2 =
            bigIntegerToBytes(
                this.maxPriorityFeePerGas
                    .get()
                    .getAsBigInteger()
                    .add(this.baseFee.get().getAsBigInteger()));
        final boolean result = wcp.callLT(maxFee, row7arg2);
        getFullTip = !result;
        this.callsToEucAndWcp.add(TxnDataComparisonRecord.callToLt(maxFee, row7arg2, result));
      }
    }
  }

  public void setCallWcpLastTxOfBlock(final Bytes blockGasLimit) {
    final Bytes arg1 = Bytes.minimalBytes(this.cumulativeGasConsumption);
    this.wcp.callLEQ(arg1, blockGasLimit);
    this.callsToEucAndWcp.add(TxnDataComparisonRecord.callToLeq(arg1, blockGasLimit, true));
  }

  public Bytes computeGasPriceColumn() {
    switch (this.type) {
      case FRONTIER, ACCESS_LIST -> {
        return bigIntegerToBytes(effectiveGasPrice);
      }
      case EIP1559 -> {
        return getFullTip
            ? bigIntegerToBytes(
                this.baseFee
                    .get()
                    .getAsBigInteger()
                    .add(this.maxPriorityFeePerGas.get().getAsBigInteger()))
            : bigIntegerToBytes(this.maxFeePerGas.get().getAsBigInteger());
      }
      default -> throw new IllegalArgumentException("Transaction type not supported");
    }
  }

  private BigInteger setOutgoingLoRowPlus6() {
    switch (this.type) {
      case FRONTIER, ACCESS_LIST -> {
        return this.effectiveGasPrice;
      }
      case EIP1559 -> {
        return this.maxFeePerGas.get().getAsBigInteger();
      }
      default -> throw new IllegalArgumentException("Transaction type not supported");
    }
  }

  @Override
  protected int computeLineCount() {
    throw new IllegalStateException("should never be called");
  }
}
