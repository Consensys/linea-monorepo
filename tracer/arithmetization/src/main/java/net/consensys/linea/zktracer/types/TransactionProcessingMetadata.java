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

package net.consensys.linea.zktracer.types;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.*;
import static net.consensys.linea.zktracer.types.AddressUtils.effectiveToAddress;

import java.math.BigInteger;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;

import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.transients.Block;
import net.consensys.linea.zktracer.module.hub.transients.StorageInitialValues;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldView;

@Getter
public class TransactionProcessingMetadata implements PostTransactionDefer {

  final int absoluteTransactionNumber;
  final int relativeTransactionNumber;
  final int relativeBlockNumber;

  final Transaction besuTransaction;
  final Address coinbase;
  final long baseFee;

  final boolean isDeployment;

  @Accessors(fluent = true)
  final boolean requiresEvmExecution;

  @Accessors(fluent = true)
  final boolean copyTransactionCallData;

  final BigInteger initialBalance;

  final long dataCost;
  final long accessListCost;

  /* g in the EYP, defined by g = TG - g0 */
  final long initiallyAvailableGas;

  final Address effectiveTo;

  final long effectiveGasPrice;

  /* g' in the EYP*/
  @Setter long leftoverGas = -1;
  /*  Ar in the EYP*/
  @Setter long refundCounterMax = -1;
  /* g* - g' in the EYP*/
  @Setter long refundEffective = -1;
  /* Tg - g' in the EYP*/
  @Setter long gasUsed = -1;
  /* g* in the EYP */
  @Setter long gasRefunded = -1;
  /* Tg - g* in the EYP */
  @Setter long totalGasUsed = -1;

  @Accessors(fluent = true)
  @Setter
  boolean statusCode = false;

  @Setter int hubStampTransactionEnd;

  @Setter int accumulatedGasUsedInBlock = -1;

  @Accessors(fluent = true)
  @Setter
  boolean isSenderPreWarmed = false;

  @Accessors(fluent = true)
  @Setter
  boolean isReceiverPreWarmed = false;

  @Accessors(fluent = true)
  @Setter
  boolean isMinerWarmAtEndTx = false;

  @Setter List<Log> logs;

  final StorageInitialValues storage = new StorageInitialValues();

  @Setter int codeFragmentIndex = -1;

  @Setter Set<AccountSnapshot> destructedAccountsSnapshot;

  @Getter
  Map<EphemeralAccount, List<AttemptedSelfDestruct>> unexceptionalSelfDestructMap = new HashMap<>();

  @Getter Map<EphemeralAccount, Integer> effectiveSelfDestructMap = new HashMap<>();

  // Ephermeral accounts are both accounts that have been deployed on-chain
  // and accounts that live for a limited time
  public record EphemeralAccount(Address address, int deploymentNumber) {}
  ;

  public record AttemptedSelfDestruct(int hubStamp, CallFrame callFrame) {}
  ;

  public TransactionProcessingMetadata(
      final WorldView world,
      final Transaction transaction,
      final Block block,
      final int relativeTransactionNumber,
      final int absoluteTransactionNumber) {
    this.absoluteTransactionNumber = absoluteTransactionNumber;
    this.relativeBlockNumber = block.blockNumber();
    this.coinbase = block.minerAddress();
    this.baseFee = block.baseFee().toLong();

    this.besuTransaction = transaction;
    this.relativeTransactionNumber = relativeTransactionNumber;

    this.isDeployment = transaction.getTo().isEmpty();
    this.requiresEvmExecution = computeRequiresEvmExecution(world);
    this.copyTransactionCallData = computeCopyCallData();

    this.initialBalance = getInitialBalance(world);

    // Note: Besu's dataCost computation contains the 21_000 transaction cost
    this.dataCost =
        ZkTracer.gasCalculator.transactionIntrinsicGasCost(
                besuTransaction.getPayload(), isDeployment)
            - GAS_CONST_G_TRANSACTION;
    this.accessListCost =
        besuTransaction.getAccessList().map(ZkTracer.gasCalculator::accessListGasCost).orElse(0L);
    this.initiallyAvailableGas = getInitiallyAvailableGas();

    this.effectiveTo = effectiveToAddress(besuTransaction);

    this.effectiveGasPrice = computeEffectiveGasPrice();
  }

  public void setPreFinalisationValues(
      final long leftOverGas,
      final long refundCounterMax,
      final boolean minerIsWarmAtFinalisation,
      final int accumulatedGasUsedInBlockAtStartTx) {

    this.isMinerWarmAtEndTx(minerIsWarmAtFinalisation);
    this.refundCounterMax = refundCounterMax;
    this.setLeftoverGas(leftOverGas);
    this.gasUsed = computeGasUsed();
    this.refundEffective = computeRefundEffective();
    this.gasRefunded = computeRefunded();
    this.totalGasUsed = computeTotalGasUsed();
    this.accumulatedGasUsedInBlock = (int) (accumulatedGasUsedInBlockAtStartTx + totalGasUsed);
  }

  public void completeLineaTransaction(
      Hub hub, final boolean statusCode, final List<Log> logs, final Set<Address> selfDestructs) {
    this.statusCode = statusCode;
    this.hubStampTransactionEnd = hub.stamp();
    this.logs = logs;
    for (Address address : selfDestructs) {
      this.destructedAccountsSnapshot.add(AccountSnapshot.canonical(hub, address));
    }
  }

  private boolean computeCopyCallData() {
    return requiresEvmExecution && !isDeployment && !besuTransaction.getData().get().isEmpty();
  }

  private boolean computeRequiresEvmExecution(WorldView world) {
    if (!isDeployment) {
      return Optional.ofNullable(world.get(this.besuTransaction.getTo().get()))
          .map(a -> !a.getCode().isEmpty())
          .orElse(false);
    }

    return !this.besuTransaction.getInit().get().isEmpty();
  }

  private BigInteger getInitialBalance(WorldView world) {
    final Address sender = besuTransaction.getSender();
    return world.get(sender).getBalance().getAsBigInteger();
  }

  public long getUpfrontGasCost() {
    return dataCost
        + (isDeployment ? GAS_CONST_G_CREATE : 0)
        + GAS_CONST_G_TRANSACTION
        + accessListCost;
  }

  public long getInitiallyAvailableGas() {
    return besuTransaction.getGasLimit() - getUpfrontGasCost();
  }

  private long computeRefundEffective() {
    final long maxRefundableAmount = this.getGasUsed() / MAX_REFUND_QUOTIENT;
    return Math.min(maxRefundableAmount, refundCounterMax);
  }

  private long computeEffectiveGasPrice() {
    final Transaction tx = besuTransaction;
    switch (tx.getType()) {
      case FRONTIER, ACCESS_LIST -> {
        return tx.getGasPrice().get().getAsBigInteger().longValueExact();
      }
      case EIP1559 -> {
        final long baseFee = this.baseFee;
        final long maxPriorityFee =
            tx.getMaxPriorityFeePerGas().get().getAsBigInteger().longValueExact();
        final long maxFeePerGas = tx.getMaxFeePerGas().get().getAsBigInteger().longValueExact();
        return Math.min(baseFee + maxPriorityFee, maxFeePerGas);
      }
      default -> throw new IllegalArgumentException("Transaction type not supported");
    }
  }

  public Address getSender() {
    return besuTransaction.getSender();
  }

  public boolean requiresPrewarming() {
    return requiresEvmExecution && (accessListCost != 0);
  }

  public boolean requiresCfiUpdate() {
    return requiresEvmExecution && isDeployment;
  }

  /* Tg - g' in the EYP*/
  public long computeGasUsed() {
    return besuTransaction.getGasLimit() - leftoverGas;
  }

  /* g* in the EYP */
  public long computeRefunded() {
    return leftoverGas + this.refundEffective;
  }

  /* Tg - g* in the EYP */
  public long computeTotalGasUsed() {
    return besuTransaction.getGasLimit() - getGasRefunded();
  }

  public long weiPerGasForMiner() {
    return switch (besuTransaction.getType()) {
      case FRONTIER, ACCESS_LIST -> effectiveGasPrice;
      case EIP1559 -> effectiveGasPrice - baseFee;
      default -> throw new IllegalStateException(
          "Transaction Type not supported: " + besuTransaction.getType());
    };
  }

  public Wei getMinerReward() {
    return Wei.of(
        BigInteger.valueOf(totalGasUsed).multiply(BigInteger.valueOf(weiPerGasForMiner())));
  }

  public Wei getGasRefundInWei() {
    return Wei.of(BigInteger.valueOf(gasRefunded).multiply(BigInteger.valueOf(effectiveGasPrice)));
  }

  public int numberWarmedAddress() {
    return this.besuTransaction.getAccessList().isPresent()
        ? this.besuTransaction.getAccessList().get().size()
        : 0;
  }

  public int numberWarmedKey() {
    return this.besuTransaction.getAccessList().isPresent()
        ? this.besuTransaction.getAccessList().get().stream()
            .mapToInt(accessListEntry -> accessListEntry.storageKeys().size())
            .sum()
        : 0;
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    for (Map.Entry<EphemeralAccount, List<AttemptedSelfDestruct>> entry :
        this.unexceptionalSelfDestructMap.entrySet()) {

      EphemeralAccount ephemeralAccount = entry.getKey();
      List<AttemptedSelfDestruct> attemptedSelfDestructs = entry.getValue();

      // For each address, deployment number, we find selfDestructTime as
      // the time in which the first unexceptional and un-reverted SELFDESTRUCT occurs
      // Then we add this value in a new map
      for (AttemptedSelfDestruct attemptedSelfDestruct : attemptedSelfDestructs) {
        if (attemptedSelfDestruct.callFrame().revertStamp() == 0) {
          int selfDestructTime = attemptedSelfDestruct.hubStamp();
          this.effectiveSelfDestructMap.put(ephemeralAccount, selfDestructTime);
          break;
        }
      }
    }
  }
}
