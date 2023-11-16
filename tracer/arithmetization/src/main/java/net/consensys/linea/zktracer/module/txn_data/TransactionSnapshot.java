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

package net.consensys.linea.zktracer.module.txn_data;

import java.math.BigInteger;
import java.util.List;
import java.util.Optional;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
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
public final class TransactionSnapshot {
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
  /** CodeFragmentIndex, given by the {@link net.consensys.linea.zktracer.module.romLex.RomLex} */
  private final int codeIdBeforeLex;
  /** The sender balance when it sent the transaction */
  private final BigInteger initialSenderBalance;
  /** The payload of the transaction, calldata or initcode */
  private final Bytes payload;

  private final long gasLimit;
  private final BigInteger effectiveGasPrice;
  private final Optional<? extends Quantity> maxFeePerGas;
  private final Optional<? extends Quantity> maxPriorityFeePerGas;
  @Setter private boolean status;
  @Setter private long refundCounter;
  @Setter private long leftoverGas;
  @Setter private long effectiveGasRefund;
  @Setter private long cumulativeGasConsumption;

  public TransactionSnapshot(
      BigInteger value,
      Address from,
      Address to,
      long nonce,
      int prewarmedAddressesCount,
      int prewarmedStorageKeysCount,
      boolean isDeployment,
      boolean requiresEvmExecution,
      TransactionType type,
      int codeFragmentIndex,
      BigInteger initialSenderBalance,
      Bytes payload,
      long gasLimit,
      BigInteger effectiveGasPrice,
      Optional<? extends Quantity> maxFeePerGas,
      Optional<? extends Quantity> maxPriorityFeePerGas) {

    this.value = value;
    this.from = from;
    this.to = to;
    this.nonce = nonce;
    this.prewarmedAddressesCount = prewarmedAddressesCount;
    this.prewarmedStorageKeysCount = prewarmedStorageKeysCount;
    this.isDeployment = isDeployment;
    this.requiresEvmExecution = requiresEvmExecution;
    this.type = type;
    this.codeIdBeforeLex = codeFragmentIndex;
    this.initialSenderBalance = initialSenderBalance;
    this.payload = payload;
    this.gasLimit = gasLimit;
    this.effectiveGasPrice = effectiveGasPrice;
    this.maxFeePerGas = maxFeePerGas;
    this.maxPriorityFeePerGas = maxPriorityFeePerGas;
  }

  public static TransactionSnapshot fromTransaction(
      int codeIdBeforeLex, Transaction tx, WorldView world, Optional<Wei> baseFee) {

    return new TransactionSnapshot(
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
        codeIdBeforeLex,
        Optional.ofNullable(tx.getSender())
            .map(world::get)
            .map(x -> x.getBalance().getAsBigInteger())
            .orElse(BigInteger.ZERO),
        tx.getPayload().copy(),
        tx.getGasLimit(),
        computeEffectiveGasPrice(baseFee, tx),
        tx.getMaxFeePerGas(),
        tx.getMaxPriorityFeePerGas());
  }

  // dataCost returns the gas cost of the call data / init code
  // 0x00 costs 4 gas, any other byte costs 16 = 4 + 12
  public long dataCost() {
    Bytes payload = this.payload();
    long dataCost = 4 * (long) payload.size();
    for (int i = 0; i < payload.size(); i++) {
      if (payload.get(i) != 0) {
        dataCost += 12;
      }
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

  /**
   * Computes minimumNecessaryBalance := T_v + T_g * T_p where
   *
   * <p>T_v = value field of transaction
   *
   * <p>T_p = gas price for Type 0 & Type 1 transactions | max priority fee for Type 2 transactions
   *
   * @return
   */
  BigInteger getMaximalUpfrontCost() {
    BigInteger minimumNecessaryBalance = this.value();
    BigInteger gasLimit = BigInteger.valueOf(this.gasLimit());
    BigInteger maximalGasPrice =
        switch (this.type()) {
          case FRONTIER, ACCESS_LIST -> this.effectiveGasPrice;
          case EIP1559 -> this.maxFeePerGas().get().getAsBigInteger();
          default -> throw new RuntimeException("transaction type not supported");
        };
    return minimumNecessaryBalance.add(gasLimit.multiply(maximalGasPrice));
  }

  /**
   * Converts the {@link TransactionType} to an integer.
   *
   * @return an integer encoding the transaction type
   */
  int typeAsInt() {
    return switch (this.type()) {
      case FRONTIER -> 0;
      case ACCESS_LIST -> 1;
      case EIP1559 -> 2;
      default -> throw new RuntimeException("transaction type not supported");
    };
  }

  // getData should return either:
  // - call data (message call)
  // - init code (contract creation)
  long maxCounter() {
    return switch (this.type()) {
      case FRONTIER -> 1 + TxnDataTrace.nROWS0;
      case ACCESS_LIST -> 1 + TxnDataTrace.nROWS1;
      case EIP1559 -> 1 + TxnDataTrace.nROWS2;
      default -> throw new RuntimeException("transaction type not supported");
    };
  }

  long getUpfrontGasCost() {
    long initialCost = this.dataCost();

    if (this.isDeployment()) {
      initialCost += TxnDataTrace.G_txcreate;
    }

    initialCost += TxnDataTrace.G_transaction;

    if (this.type() != TransactionType.FRONTIER) {
      initialCost += (long) this.prewarmedAddressesCount() * TxnDataTrace.G_accesslistaddress;
      initialCost += (long) this.prewarmedStorageKeysCount() * TxnDataTrace.G_accessliststorage;
    }

    Preconditions.checkArgument(this.gasLimit() >= initialCost, "gasLimit < initialGasCost");

    return initialCost;
  }

  BigInteger getLimitMinusLeftoverGas() {
    return BigInteger.valueOf(this.gasLimit() - this.leftoverGas());
  }

  BigInteger getLimitMinusLeftoverGasDividedByTwo() {
    return this.getLimitMinusLeftoverGas().divide(BigInteger.TWO);
  }
}
