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

package net.consensys.linea.zktracer.module.txndata;

import static net.consensys.linea.zktracer.module.Util.getTxTypeAsInt;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_0;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_1;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_2;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_3;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_4;
import static net.consensys.linea.zktracer.module.txndata.Trace.COMMON_RLP_TXN_PHASE_NUMBER_5;
import static net.consensys.linea.zktracer.module.txndata.Trace.EIP2681_MAX_NONCE;
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
import static net.consensys.linea.zktracer.types.AddressUtils.highPart;
import static net.consensys.linea.zktracer.types.AddressUtils.lowPart;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Conversions.booleanToInt;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import lombok.Getter;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.TransactionType;

public class TxndataOperation extends ModuleOperation {
  private final Wcp wcp;
  private final Euc euc;
  @Getter public final TransactionProcessingMetadata tx;

  private static final Bytes EIP_2681_MAX_NONCE = Bytes.minimalBytes(EIP2681_MAX_NONCE);
  private static final int N_ROWS_TX_MAX =
      Math.max(Math.max(NB_ROWS_TYPE_0, NB_ROWS_TYPE_1), NB_ROWS_TYPE_2);
  private static final int NB_WCP_EUC_ROWS_FRONTIER_ACCESS_LIST = 6;
  private final List<TxnDataComparisonRecord> callsToEucAndWcp = new ArrayList<>(N_ROWS_TX_MAX);
  private final ArrayList<RlptxnOutgoing> valuesToRlptxn = new ArrayList<>(N_ROWS_TX_MAX);
  private final ArrayList<RlptxrcptOutgoing> valuesToRlpTxrcpt = new ArrayList<>(N_ROWS_TX_MAX);

  public TxndataOperation(Wcp wcp, Euc euc, TransactionProcessingMetadata tx) {
    this.wcp = wcp;
    this.euc = euc;
    this.tx = tx;

    this.setCallsToEucAndWcp();
  }

  private void setCallsToEucAndWcp() {
    // i + nonce_row_offset
    final Bytes nonce = Bytes.minimalBytes(tx.getBesuTransaction().getNonce());
    wcp.callLT(nonce, EIP_2681_MAX_NONCE);
    callsToEucAndWcp.add(TxnDataComparisonRecord.callToLt(nonce, EIP_2681_MAX_NONCE, true));

    // i + initial_balance_row_offset
    final Bytes initBalance = bigIntegerToBytes(tx.getInitialBalance());
    final BigInteger value = tx.getBesuTransaction().getValue().getAsBigInteger();
    final Bytes row0arg2 =
        bigIntegerToBytes(
            value.add(
                outgoingLowRow6()
                    .multiply(BigInteger.valueOf(tx.getBesuTransaction().getGasLimit()))));
    wcp.callLT(initBalance, row0arg2);
    callsToEucAndWcp.add(TxnDataComparisonRecord.callToLt(initBalance, row0arg2, false));

    // i + sufficient_gas_row_offset
    final Bytes row1arg1 = Bytes.minimalBytes(tx.getBesuTransaction().getGasLimit());
    final Bytes row1arg2 = Bytes.minimalBytes(tx.getUpfrontGasCost());
    wcp.callLT(row1arg1, row1arg2);
    callsToEucAndWcp.add(TxnDataComparisonRecord.callToLt(row1arg1, row1arg2, false));

    // i + upper_limit_refunds_row_offset
    final Bytes row2arg1 =
        Bytes.minimalBytes(tx.getBesuTransaction().getGasLimit() - tx.getLeftoverGas());
    final Bytes row2arg2 = Bytes.of(MAX_REFUND_QUOTIENT);
    final Bytes refundLimit = euc.callEUC(row2arg1, row2arg2).quotient();
    callsToEucAndWcp.add(TxnDataComparisonRecord.callToEuc(row2arg1, row2arg2, refundLimit));

    // i + effective_refund_row_offset
    final Bytes refundCounterMax = Bytes.minimalBytes(tx.getRefundCounterMax());
    final boolean getFullRefund = wcp.callLT(refundCounterMax, refundLimit);
    callsToEucAndWcp.add(
        TxnDataComparisonRecord.callToLt(refundCounterMax, refundLimit, getFullRefund));

    // i + detecting_empty_call_data_row_offset
    final Bytes row4arg1 = Bytes.minimalBytes(tx.getBesuTransaction().getPayload().size());
    final boolean nonZeroDataSize = wcp.callISZERO(row4arg1);
    callsToEucAndWcp.add(TxnDataComparisonRecord.callToIsZero(row4arg1, nonZeroDataSize));

    switch (tx.getBesuTransaction().getType()) {
      case FRONTIER -> {
        for (int i = NB_WCP_EUC_ROWS_FRONTIER_ACCESS_LIST; i < NB_ROWS_TYPE_0; i++) {
          callsToEucAndWcp.add(TxnDataComparisonRecord.empty());
        }
      }
      case ACCESS_LIST -> {
        for (int i = NB_WCP_EUC_ROWS_FRONTIER_ACCESS_LIST; i < NB_ROWS_TYPE_1; i++) {
          callsToEucAndWcp.add(TxnDataComparisonRecord.empty());
        }
      }
      case EIP1559 -> {
        // i + max_fee_and_basefee_row_offset
        final Bytes maxFee =
            bigIntegerToBytes(tx.getBesuTransaction().getMaxFeePerGas().get().getAsBigInteger());
        final Bytes row5arg2 = Bytes.minimalBytes(tx.getBaseFee());
        wcp.callLT(maxFee, row5arg2);
        callsToEucAndWcp.add(TxnDataComparisonRecord.callToLt(maxFee, row5arg2, false));

        // i + maxfee_and_max_priority_fee_row_offset
        final Bytes row6arg2 =
            bigIntegerToBytes(
                tx.getBesuTransaction().getMaxPriorityFeePerGas().get().getAsBigInteger());
        wcp.callLT(maxFee, row6arg2);
        callsToEucAndWcp.add(TxnDataComparisonRecord.callToLt(maxFee, row6arg2, false));

        // i + computing_effective_gas_price_row_offset
        final Bytes row7arg2 =
            bigIntegerToBytes(
                tx.getBesuTransaction()
                    .getMaxPriorityFeePerGas()
                    .get()
                    .getAsBigInteger()
                    .add(BigInteger.valueOf(tx.getBaseFee())));
        final boolean result = wcp.callLT(maxFee, row7arg2);
        callsToEucAndWcp.add(TxnDataComparisonRecord.callToLt(maxFee, row7arg2, result));
      }
    }
  }

  @Override
  protected int computeLineCount() {
    // Count the number of rows of each tx, only depending on the type of the transaction
    return switch (tx.getBesuTransaction().getType()) {
      case FRONTIER -> NB_ROWS_TYPE_0;
      case ACCESS_LIST -> NB_ROWS_TYPE_1;
      case EIP1559 -> NB_ROWS_TYPE_2;
      default -> throw new RuntimeException(
          "Transaction type not supported:" + tx.getBesuTransaction().getType());
    };
  }

  private void setRlptxnValues() {
    // i+0
    valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_0,
            Bytes.EMPTY,
            Bytes.ofUnsignedInt(getTxTypeAsInt(tx.getBesuTransaction().getType()))));
    // i+1
    valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_1,
            tx.isDeployment() ? Bytes.EMPTY : tx.getEffectiveRecipient().slice(0, 4),
            tx.isDeployment() ? Bytes.EMPTY : lowPart(tx.getEffectiveRecipient())));

    // i+2
    valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_2,
            Bytes.EMPTY,
            Bytes.ofUnsignedLong(tx.getBesuTransaction().getNonce())));

    // i+3
    valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_3,
            tx.isDeployment() ? Bytes.of(1) : Bytes.EMPTY,
            bigIntegerToBytes(tx.getBesuTransaction().getValue().getAsBigInteger())));

    // i+4
    valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_4,
            Bytes.ofUnsignedLong(tx.getDataCost()),
            Bytes.ofUnsignedLong(tx.getBesuTransaction().getPayload().size())));

    // i+5
    valuesToRlptxn.add(
        RlptxnOutgoing.set(
            (short) COMMON_RLP_TXN_PHASE_NUMBER_5,
            Bytes.EMPTY,
            Bytes.ofUnsignedLong(tx.getBesuTransaction().getGasLimit())));

    switch (tx.getBesuTransaction().getType()) {
      case FRONTIER -> {
        // i+6
        valuesToRlptxn.add(
            RlptxnOutgoing.set(
                (short) TYPE_0_RLP_TXN_PHASE_NUMBER_6,
                Bytes.EMPTY,
                Bytes.minimalBytes(tx.getEffectiveGasPrice())));
        for (int i = 7; i < NB_ROWS_TYPE_0 + 1; i++) {
          valuesToRlptxn.add(RlptxnOutgoing.empty());
        }
      }
      case ACCESS_LIST -> {
        // i+6
        valuesToRlptxn.add(
            RlptxnOutgoing.set(
                (short) TYPE_1_RLP_TXN_PHASE_NUMBER_6,
                Bytes.EMPTY,
                Bytes.minimalBytes(tx.getEffectiveGasPrice())));

        // i+7
        valuesToRlptxn.add(
            RlptxnOutgoing.set(
                (short) TYPE_1_RLP_TXN_PHASE_NUMBER_7,
                Bytes.ofUnsignedInt(tx.numberWarmedKey()),
                Bytes.ofUnsignedInt(tx.numberWarmedAddress())));

        for (int i = 8; i < NB_ROWS_TYPE_1 + 1; i++) {
          valuesToRlptxn.add(RlptxnOutgoing.empty());
        }
      }

      case EIP1559 -> {
        // i+6
        valuesToRlptxn.add(
            RlptxnOutgoing.set(
                (short) TYPE_2_RLP_TXN_PHASE_NUMBER_6,
                bigIntegerToBytes(
                    tx.getBesuTransaction().getMaxPriorityFeePerGas().get().getAsBigInteger()),
                bigIntegerToBytes(outgoingLowRow6())));

        // i+7
        valuesToRlptxn.add(
            RlptxnOutgoing.set(
                (short) TYPE_2_RLP_TXN_PHASE_NUMBER_7,
                Bytes.ofUnsignedInt(tx.numberWarmedKey()),
                Bytes.ofUnsignedInt(tx.numberWarmedAddress())));

        for (int i = 8; i < NB_ROWS_TYPE_2 + 1; i++) {
          valuesToRlptxn.add(RlptxnOutgoing.empty());
        }
      }
    }
  }

  public void setRlptxrcptValues() {
    // i+0
    this.valuesToRlpTxrcpt.add(
        RlptxrcptOutgoing.set(
            (short) RLP_RCPT_SUBPHASE_ID_TYPE, getTxTypeAsInt(tx.getBesuTransaction().getType())));
    // i+1
    this.valuesToRlpTxrcpt.add(
        RlptxrcptOutgoing.set(
            (short) RLP_RCPT_SUBPHASE_ID_STATUS_CODE, booleanToInt(tx.statusCode())));
    // i+2
    this.valuesToRlpTxrcpt.add(
        RlptxrcptOutgoing.set(
            (short) RLP_RCPT_SUBPHASE_ID_CUMUL_GAS, tx.getAccumulatedGasUsedInBlock()));
    // i+3 to i+MAX_NB_ROWS
    for (int ct = 3; ct < N_ROWS_TX_MAX; ct++) {
      this.valuesToRlpTxrcpt.add(RlptxrcptOutgoing.emptyValue());
    }
  }

  public void setCallWcpLastTxOfBlock(final Bytes blockGasLimit) {
    final Bytes arg1 = Bytes.minimalBytes(tx.getAccumulatedGasUsedInBlock());
    wcp.callLEQ(arg1, blockGasLimit);
    callsToEucAndWcp.add(TxnDataComparisonRecord.callToLeq(arg1, blockGasLimit, true));
  }

  private BigInteger outgoingLowRow6() {
    return switch (tx.getBesuTransaction().getType()) {
      case FRONTIER, ACCESS_LIST -> tx.getBesuTransaction().getGasPrice().get().getAsBigInteger();
      case EIP1559 -> tx.getBesuTransaction().getMaxFeePerGas().get().getAsBigInteger();
      default -> throw new RuntimeException(
          "Transaction type not supported:" + tx.getBesuTransaction().getType());
    };
  }

  public void traceTx(Trace trace, BlockSnapshot block, int absTxNumMax) {

    this.setRlptxnValues();
    this.setRlptxrcptValues();

    final boolean isLastTxOfTheBlock =
        tx.getRelativeTransactionNumber() == block.getNbOfTxsInBlock();
    if (isLastTxOfTheBlock) {
      valuesToRlptxn.add(RlptxnOutgoing.empty());
      valuesToRlpTxrcpt.add(RlptxrcptOutgoing.emptyValue());
    }

    final long fromHi = highPart(tx.getSender());
    final Bytes fromLo = lowPart(tx.getSender());
    final Bytes nonce = Bytes.ofUnsignedLong(tx.getBesuTransaction().getNonce());
    final Bytes initialBalance = bigIntegerToBytes(tx.getInitialBalance());
    final Bytes value = bigIntegerToBytes(tx.getBesuTransaction().getValue().getAsBigInteger());
    final long toHi = highPart(tx.getEffectiveRecipient());
    final Bytes toLo = lowPart(tx.getEffectiveRecipient());
    final Bytes gasLimit = Bytes.minimalBytes(tx.getBesuTransaction().getGasLimit());
    final Bytes gasInitiallyAvailable = Bytes.minimalBytes(tx.getInitiallyAvailableGas());
    final Bytes gasPrice = Bytes.minimalBytes(tx.getEffectiveGasPrice());
    final Bytes priorityFeePerGas = Bytes.minimalBytes(tx.feeRateForCoinbase());
    final Bytes baseFee = block.getBaseFee().get().toMinimalBytes();
    final long coinbaseHi = highPart(block.getCoinbaseAddress());
    final Bytes coinbaseLo = lowPart(block.getCoinbaseAddress());
    final int callDataSize = tx.isDeployment() ? 0 : tx.getBesuTransaction().getPayload().size();
    final int initCodeSize = tx.isDeployment() ? tx.getBesuTransaction().getPayload().size() : 0;
    final Bytes gasLeftOver = Bytes.minimalBytes(tx.getLeftoverGas());
    final Bytes refundCounter = Bytes.minimalBytes(tx.getRefundCounterMax());
    final Bytes refundEffective = Bytes.minimalBytes(tx.getGasRefunded());
    final Bytes cumulativeGas = Bytes.minimalBytes(tx.getAccumulatedGasUsedInBlock());

    final int nbLInes = isLastTxOfTheBlock ? this.lineCount() + 1 : this.lineCount();

    for (int ct = 0; ct < nbLInes; ct++) {
      trace
          .absTxNumMax(absTxNumMax)
          .absTxNum(tx.getAbsoluteTransactionNumber())
          .relBlock(tx.getRelativeBlockNumber())
          .relTxNumMax(block.getNbOfTxsInBlock())
          .relTxNum(tx.getRelativeTransactionNumber())
          .isLastTxOfBlock(isLastTxOfTheBlock)
          .ct(UnsignedByte.of(ct))
          .fromHi(fromHi)
          .fromLo(fromLo)
          .nonce(nonce)
          .initialBalance(initialBalance)
          .value(value)
          .toHi(toHi)
          .toLo(toLo)
          .isDep(tx.isDeployment())
          .gasLimit(gasLimit)
          .gasInitiallyAvailable(gasInitiallyAvailable)
          .gasPrice(gasPrice)
          .priorityFeePerGas(priorityFeePerGas)
          .basefee(baseFee)
          .coinbaseHi(coinbaseHi)
          .coinbaseLo(coinbaseLo)
          .blockGasLimit(block.getBlockGasLimit())
          .callDataSize(callDataSize)
          .initCodeSize(initCodeSize)
          .type0(tx.getBesuTransaction().getType() == TransactionType.FRONTIER)
          .type1(tx.getBesuTransaction().getType() == TransactionType.ACCESS_LIST)
          .type2(tx.getBesuTransaction().getType() == TransactionType.EIP1559)
          .requiresEvmExecution(tx.requiresEvmExecution())
          .copyTxcd(tx.copyTransactionCallData())
          .gasLeftover(gasLeftOver)
          .refundCounter(refundCounter)
          .refundEffective(refundEffective)
          .gasCumulative(cumulativeGas)
          .statusCode(tx.statusCode())
          .codeFragmentIndex(tx.getCodeFragmentIndex())
          .phaseRlpTxn(UnsignedByte.of(valuesToRlptxn.get(ct).phase()))
          .outgoingHi(valuesToRlptxn.get(ct).outGoingHi())
          .outgoingLo(valuesToRlptxn.get(ct).outGoingLo())
          .eucFlag(callsToEucAndWcp.get(ct).eucFlag())
          .wcpFlag(callsToEucAndWcp.get(ct).wcpFlag())
          .inst(UnsignedByte.of(callsToEucAndWcp.get(ct).instruction()))
          .argOneLo(callsToEucAndWcp.get(ct).arg1())
          .argTwoLo(callsToEucAndWcp.get(ct).arg2())
          .res(callsToEucAndWcp.get(ct).result())
          .phaseRlpTxnrcpt(UnsignedByte.of(valuesToRlpTxrcpt.get(ct).phase()))
          .outgoingRlpTxnrcpt(Bytes.ofUnsignedLong(valuesToRlpTxrcpt.get(ct).outgoing()))
          .validateRow();
    }
  }
}
