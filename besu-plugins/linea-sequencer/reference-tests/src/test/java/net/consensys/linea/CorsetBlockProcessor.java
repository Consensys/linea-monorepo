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

package net.consensys.linea;

import static org.hyperledger.besu.ethereum.mainnet.feemarket.ExcessBlobGasCalculator.calculateExcessBlobGasForParent;

import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.BlockProcessingOutputs;
import org.hyperledger.besu.ethereum.BlockProcessingResult;
import org.hyperledger.besu.ethereum.bonsai.worldview.BonsaiWorldState;
import org.hyperledger.besu.ethereum.bonsai.worldview.BonsaiWorldStateUpdateAccumulator;
import org.hyperledger.besu.ethereum.chain.Blockchain;
import org.hyperledger.besu.ethereum.core.BlockBody;
import org.hyperledger.besu.ethereum.core.BlockHeader;
import org.hyperledger.besu.ethereum.core.Deposit;
import org.hyperledger.besu.ethereum.core.MutableWorldState;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.core.TransactionReceipt;
import org.hyperledger.besu.ethereum.core.Withdrawal;
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockProcessor;
import org.hyperledger.besu.ethereum.mainnet.MainnetTransactionProcessor;
import org.hyperledger.besu.ethereum.mainnet.MiningBeneficiaryCalculator;
import org.hyperledger.besu.ethereum.mainnet.ParentBeaconBlockRootHelper;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSchedule;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSpec;
import org.hyperledger.besu.ethereum.mainnet.TransactionValidationParams;
import org.hyperledger.besu.ethereum.mainnet.WithdrawalsProcessor;
import org.hyperledger.besu.ethereum.privacy.storage.PrivateMetadataUpdater;
import org.hyperledger.besu.ethereum.processing.TransactionProcessingResult;
import org.hyperledger.besu.ethereum.vm.BlockHashLookup;
import org.hyperledger.besu.ethereum.vm.CachingBlockHashLookup;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;

@Slf4j
public class CorsetBlockProcessor extends MainnetBlockProcessor {
  private final ProtocolSchedule protocolSchedule;
  private final ZkTracer zkTracer;

  public CorsetBlockProcessor(
      final MainnetTransactionProcessor transactionProcessor,
      final TransactionReceiptFactory transactionReceiptFactory,
      final Wei blockReward,
      final MiningBeneficiaryCalculator miningBeneficiaryCalculator,
      final boolean skipZeroBlockRewards,
      final ProtocolSchedule protocolSchedule,
      final ZkTracer zkTracer) {
    super(
        transactionProcessor,
        transactionReceiptFactory,
        blockReward,
        miningBeneficiaryCalculator,
        skipZeroBlockRewards,
        protocolSchedule);
    this.protocolSchedule = protocolSchedule;
    this.zkTracer = zkTracer;
  }

  @Override
  public BlockProcessingResult processBlock(
      final Blockchain blockchain,
      final MutableWorldState worldState,
      final BlockHeader blockHeader,
      final List<Transaction> transactions,
      final List<BlockHeader> ommers,
      final Optional<List<Withdrawal>> maybeWithdrawals,
      final Optional<List<Deposit>> maybeDeposits,
      final PrivateMetadataUpdater privateMetadataUpdater) {
    final List<TransactionReceipt> receipts = new ArrayList<>();
    long currentGasUsed = 0;
    BlockBody blockBody = new BlockBody(transactions, new ArrayList<>());

    final ProtocolSpec protocolSpec = protocolSchedule.getByBlockHeader(blockHeader);

    if (blockHeader.getParentBeaconBlockRoot().isPresent()) {
      final WorldUpdater updater = worldState.updater();
      ParentBeaconBlockRootHelper.storeParentBeaconBlockRoot(
          updater, blockHeader.getTimestamp(), blockHeader.getParentBeaconBlockRoot().get());
    }

    for (final Transaction transaction : transactions) {
      if (!hasAvailableBlockBudget(blockHeader, transaction, currentGasUsed)) {
        return new BlockProcessingResult(Optional.empty(), "provided gas insufficient");
      }

      final WorldUpdater worldStateUpdater = worldState.updater();

      final BlockHashLookup blockHashLookup = new CachingBlockHashLookup(blockHeader, blockchain);
      final Address miningBeneficiary =
          miningBeneficiaryCalculator.calculateBeneficiary(blockHeader);

      Optional<BlockHeader> maybeParentHeader =
          blockchain.getBlockHeader(blockHeader.getParentHash());

      Wei blobGasPrice =
          maybeParentHeader
              .map(
                  (parentHeader) ->
                      protocolSpec
                          .getFeeMarket()
                          .blobGasPricePerGas(
                              calculateExcessBlobGasForParent(protocolSpec, parentHeader)))
              .orElse(Wei.ZERO);

      zkTracer.traceStartConflation(1);
      zkTracer.traceStartBlock(blockHeader, blockBody);
      final TransactionProcessingResult result =
          transactionProcessor.processTransaction(
              blockchain,
              worldStateUpdater,
              blockHeader,
              transaction,
              miningBeneficiary,
              zkTracer,
              blockHashLookup,
              true,
              TransactionValidationParams.processingBlock(),
              privateMetadataUpdater,
              blobGasPrice);

      if (result.isInvalid()) {
        String errorMessage =
            MessageFormat.format(
                "Block processing error: transaction invalid {0}. Block {1} Transaction {2}",
                result.getValidationResult().getErrorMessage(),
                blockHeader.getHash().toHexString(),
                transaction.getHash().toHexString());
        log.info(errorMessage);
        if (worldState instanceof BonsaiWorldState) {
          ((BonsaiWorldStateUpdateAccumulator) worldStateUpdater).reset();
        }
        return new BlockProcessingResult(Optional.empty(), errorMessage);
      }
      worldStateUpdater.commit();

      currentGasUsed += transaction.getGasLimit() - result.getGasRemaining();
      final TransactionReceipt transactionReceipt =
          transactionReceiptFactory.create(
              transaction.getType(), result, worldState, currentGasUsed);
      receipts.add(transactionReceipt);
    }

    final Optional<WithdrawalsProcessor> maybeWithdrawalsProcessor =
        protocolSpec.getWithdrawalsProcessor();
    if (maybeWithdrawalsProcessor.isPresent() && maybeWithdrawals.isPresent()) {
      try {
        maybeWithdrawalsProcessor
            .get()
            .processWithdrawals(maybeWithdrawals.get(), worldState.updater());
      } catch (final Exception e) {
        log.error("failed processing withdrawals", e);
        return new BlockProcessingResult(Optional.empty(), e);
      }
    }

    if (!rewardCoinbase(worldState, blockHeader, ommers, skipZeroBlockRewards)) {
      // no need to log, rewardCoinbase logs the error.
      if (worldState instanceof BonsaiWorldState) {
        ((BonsaiWorldStateUpdateAccumulator) worldState.updater()).reset();
      }
      return new BlockProcessingResult(Optional.empty(), "ommer too old");
    }

    try {
      worldState.persist(blockHeader);
      //    } catch (MerkleTrieException e) {
      //      if (worldState instanceof BonsaiWorldState) {
      //        ((BonsaiWorldStateUpdateAccumulator) worldState.updater()).reset();
      //      }
      //      throw e;
    } catch (Exception e) {
      log.error("failed persisting block", e);
      return new BlockProcessingResult(Optional.empty(), e);
    }
    zkTracer.traceEndBlock(blockHeader, blockBody);
    zkTracer.traceEndConflation();

    return new BlockProcessingResult(Optional.of(new BlockProcessingOutputs(worldState, receipts)));
  }
}
