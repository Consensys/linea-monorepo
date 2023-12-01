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

import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.TxInfo;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;

public final class TransactionFragment implements TraceFragment {
  private final int batchNumber;
  private final Address minerAddress;
  private final Transaction tx;
  private final boolean evmExecutes;
  private final Wei gasPrice;
  private final Wei baseFee;
  private final boolean txSuccess;
  @Setter private long gasRefundFinalCounter;
  @Setter private long gasRefundAmount;
  @Setter private long leftoverGas;
  private final long initialGas;

  private TransactionFragment(
      int batchNumber,
      Address minerAddress,
      Transaction tx,
      boolean evmExecutes,
      Wei gasPrice,
      Wei baseFee,
      boolean txSuccess,
      long gasRefundFinalCounter,
      long gasRefundAmount,
      long initialGas) {
    this.batchNumber = batchNumber;
    this.minerAddress = minerAddress;
    this.tx = tx;
    this.evmExecutes = evmExecutes;
    this.gasPrice = gasPrice;
    this.baseFee = baseFee;
    this.txSuccess = txSuccess;
    this.gasRefundFinalCounter = gasRefundFinalCounter;
    this.gasRefundAmount = gasRefundAmount;
    this.initialGas = initialGas;
  }

  public static TransactionFragment prepare(
      int batchNumber,
      Address minerAddress,
      Transaction tx,
      boolean evmExecutes,
      Wei gasPrice,
      Wei baseFee,
      long initialGas) {
    return new TransactionFragment(
        batchNumber, minerAddress, tx, evmExecutes, gasPrice, baseFee, false, 0, 0, initialGas);
  }

  @Override
  public Trace trace(Trace trace) {
    final EWord to = EWord.of(Hub.effectiveToAddress(tx));
    final EWord from = EWord.of(tx.getSender());
    final EWord miner = EWord.of(minerAddress);

    return trace
        .peekAtTransaction(true)
        .pTransactionNonce(Bytes.ofUnsignedLong(tx.getNonce()))
        .pTransactionIsDeployment(tx.getTo().isEmpty())
        .pTransactionFromAddressHi(from.hi())
        .pTransactionFromAddressLo(from.lo())
        .pTransactionToAddressHi(to.hi())
        .pTransactionToAddressLo(to.lo())
        .pTransactionGasPrice(gasPrice)
        .pTransactionBasefee(baseFee)
        .pTransactionInitGas(Bytes.ofUnsignedLong(TxInfo.computeInitGas(tx)))
        .pTransactionInitialBalance(Bytes.ofUnsignedLong(initialGas))
        .pTransactionValue(bigIntegerToBytes(tx.getValue().getAsBigInteger()))
        .pTransactionCoinbaseAddressHi(miner.hi())
        .pTransactionCoinbaseAddressLo(miner.lo())
        .pTransactionCallDataSize(Bytes.ofUnsignedInt(tx.getData().map(Bytes::size).orElse(0)))
        .pTransactionTxnRequiresEvmExecution(evmExecutes)
        .pTransactionLeftoverGas(Bytes.ofUnsignedLong(leftoverGas))
        .pTransactionGasRefundCounterFinal(Bytes.ofUnsignedLong(gasRefundFinalCounter))
        .pTransactionGasRefundAmount(Bytes.ofUnsignedLong(gasRefundAmount))
        .pTransactionStatusCode(txSuccess);
  }
}
