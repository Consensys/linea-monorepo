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

package net.consensys.linea.zktracer.module.hub.fragment;

import static net.consensys.linea.zktracer.types.AddressUtils.highPart;
import static net.consensys.linea.zktracer.types.AddressUtils.lowPart;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;

public final class TransactionFragment implements TraceFragment {
  private final Hub hub;
  private final Address coinbaseAddress;
  private final TransactionProcessingMetadata transactionProcessingMetadata;
  @Setter private TraceSection parentSection;

  private TransactionFragment(
      Hub hub, TransactionProcessingMetadata transactionProcessingMetadata) {
    this.hub = hub;
    this.coinbaseAddress = Address.wrap(hub.coinbaseAddress.copy());
    this.transactionProcessingMetadata = transactionProcessingMetadata;
  }

  public static TransactionFragment prepare(
      Hub hub, TransactionProcessingMetadata transactionProcessingMetadata) {
    return new TransactionFragment(hub, transactionProcessingMetadata);
  }

  @Override
  public Trace trace(Trace trace) {
    final Transaction tx = transactionProcessingMetadata.getBesuTransaction();
    final Address to = transactionProcessingMetadata.getEffectiveRecipient();
    final Address from = transactionProcessingMetadata.getSender();

    return trace
        .peekAtTransaction(true)
        .pTransactionFromAddressHi(highPart(from))
        .pTransactionFromAddressLo(lowPart(from))
        .pTransactionNonce(Bytes.ofUnsignedLong(tx.getNonce()))
        .pTransactionInitialBalance(
            bigIntegerToBytes(transactionProcessingMetadata.getInitialBalance()))
        .pTransactionValue(bigIntegerToBytes(tx.getValue().getAsBigInteger()))
        .pTransactionToAddressHi(highPart(to))
        .pTransactionToAddressLo(lowPart(to))
        .pTransactionRequiresEvmExecution(transactionProcessingMetadata.requiresEvmExecution())
        .pTransactionCopyTxcd(transactionProcessingMetadata.copyTransactionCallData())
        .pTransactionIsDeployment(tx.getTo().isEmpty())
        .pTransactionIsType2(tx.getType() == TransactionType.EIP1559)
        .pTransactionGasLimit(Bytes.minimalBytes(tx.getGasLimit()))
        .pTransactionGasInitiallyAvailable(
            Bytes.minimalBytes(transactionProcessingMetadata.getInitiallyAvailableGas()))
        .pTransactionGasPrice(
            Bytes.minimalBytes(transactionProcessingMetadata.getEffectiveGasPrice()))
        .pTransactionPriorityFeePerGas(
            Bytes.minimalBytes(transactionProcessingMetadata.feeRateForCoinbase()))
        .pTransactionBasefee(Bytes.minimalBytes(transactionProcessingMetadata.getBaseFee()))
        .pTransactionCallDataSize(tx.getData().map(Bytes::size).orElse(0))
        .pTransactionInitCodeSize(tx.getInit().map(Bytes::size).orElse(0))
        .pTransactionStatusCode(transactionProcessingMetadata.statusCode())
        .pTransactionGasLeftover(Bytes.minimalBytes(transactionProcessingMetadata.getLeftoverGas()))
        .pTransactionRefundCounterInfinity(
            Bytes.minimalBytes(transactionProcessingMetadata.getRefundCounterMax()))
        .pTransactionRefundEffective(
            Bytes.minimalBytes(transactionProcessingMetadata.getGasRefunded()))
        .pTransactionCoinbaseAddressHi(highPart(coinbaseAddress))
        .pTransactionCoinbaseAddressLo(lowPart(coinbaseAddress));
  }
}
