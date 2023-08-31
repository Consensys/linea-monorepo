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

package net.consensys.linea.zktracer.module.hub.chunks;

import java.math.BigInteger;

import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;

public record TransactionChunk(
    int batchNumber, MessageFrame frame, Transaction tx, boolean evmExecutes)
    implements TraceChunk {
  @Override
  public Trace.TraceBuilder trace(Trace.TraceBuilder trace) {
    final EWord deploymentAddress = EWord.of(Address.EMPTY); // TODO compute deployment address
    final EWord to = tx.getTo().map(EWord::of).orElse(deploymentAddress);
    final EWord from = EWord.of(tx.getSender());
    final EWord miner = EWord.of(frame.getMiningBeneficiary());

    return trace
        .peekAtTransaction(true)
        .pTransactionBatchNumber(BigInteger.valueOf(batchNumber))
        .pTransactionNonce(BigInteger.valueOf(tx.getNonce()))
        .pTransactionIsDeployment(tx.getTo().isEmpty())
        .pTransactionFromAddressHi(from.hiBigInt())
        .pTransactionFromAddressLo(from.loBigInt())
        .pTransactionToAddressHi(to.hiBigInt())
        .pTransactionToAddressLo(to.loBigInt())
        // .pTransactionGasPrice(tx.getGasPrice()) TODO (compute from TX)
        // .pTransactionBaseFee(frame.getBlockValues().getBaseFee().orElse(Wei.ZERO)) // TODO
        .pTransactionInitGas(Hub.computeInitGas(tx))
        // .pTransactionInitBalance() // TODO save the init balance from TX_INIT
        .pTransactionValue(tx.getValue().getAsBigInteger())
    //      .pTransactionMinerAddressHi(miner.hiBigInt())
    //      .pTransactionMinerAddressLo(miner.loBigInt())
    // .pTransactionCalldataSize(BigInteger.valueOf(tx.getData().map(Bytes::size).orElse(0)))
    //      .pTransactionRequiresEvmExecution(evmExecutes)
    //      .pTransactionLeftOverGas(0) // TODO retcon
    //      .pTransactionRefundCounter(0) // TODO retcon
    //      .pTransactionRefundAmount(0) // TODO retcon
    //        .pTransactionStatusCode(0) // TODO retcon
    ;
  }
}
