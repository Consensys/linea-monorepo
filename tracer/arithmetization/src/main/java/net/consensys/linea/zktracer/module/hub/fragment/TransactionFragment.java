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

import static net.consensys.linea.zktracer.types.AddressUtils.effectiveToAddress;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.TransactionStack;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.worldstate.WorldView;

public final class TransactionFragment implements TraceFragment, PostTransactionDefer {
  @Setter private TraceSection parentSection;
  private final int batchNumber;
  private final Address minerAddress;
  private final Transaction tx;
  private final boolean evmExecutes;
  private final Wei gasPrice;
  private final Wei baseFee;
  private boolean txSuccess;
  private final long initialGas;

  private TransactionFragment(
      int batchNumber,
      Address minerAddress,
      Transaction tx,
      boolean evmExecutes,
      Wei gasPrice,
      Wei baseFee,
      boolean txSuccess,
      long initialGas) {
    this.batchNumber = batchNumber;
    this.minerAddress = minerAddress;
    this.tx = tx;
    this.evmExecutes = evmExecutes;
    this.gasPrice = gasPrice;
    this.baseFee = baseFee;
    this.txSuccess = txSuccess;
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
        batchNumber, minerAddress, tx, evmExecutes, gasPrice, baseFee, false, initialGas);
  }

  @Override
  public Trace trace(Trace trace) {
    final EWord to = EWord.of(effectiveToAddress(tx));
    final EWord from = EWord.of(tx.getSender());
    final EWord miner = EWord.of(minerAddress);
    long gasRefundAmount = this.parentSection.parentTrace().refundedGas();
    long leftoverGas = this.parentSection.parentTrace().leftoverGas();
    long gasRefundFinalCounter = this.parentSection.parentTrace().gasRefundFinalCounter();

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
        .pTransactionGasInitiallyAvailable(
            Bytes.ofUnsignedLong(TransactionStack.computeInitGas(tx)))
        .pTransactionInitialBalance(Bytes.ofUnsignedLong(initialGas))
        .pTransactionValue(bigIntegerToBytes(tx.getValue().getAsBigInteger()))
        .pTransactionCoinbaseAddressHi(miner.hi())
        .pTransactionCoinbaseAddressLo(miner.lo())
        .pTransactionCallDataSize(Bytes.ofUnsignedInt(tx.getData().map(Bytes::size).orElse(0)))
        .pTransactionRequiresEvmExecution(evmExecutes)
        .pTransactionGasLeftover(Bytes.ofUnsignedLong(leftoverGas))
        .pTransactionRefundCounterInfinity(Bytes.ofUnsignedLong(gasRefundFinalCounter))
        .pTransactionRefundEffective(Bytes.ofUnsignedLong(gasRefundAmount))
        .pTransactionStatusCode(txSuccess);
  }

  @Override
  public void runPostTx(Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    this.txSuccess = isSuccessful;
  }
}
