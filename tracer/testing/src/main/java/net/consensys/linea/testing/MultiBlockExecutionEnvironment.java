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

package net.consensys.linea.testing;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LINEA_BLOCK_GAS_LIMIT;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Optional;

import lombok.Builder;
import lombok.Singular;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.blockcapture.snapshots.*;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.ethereum.core.*;

@Builder
@Slf4j
public class MultiBlockExecutionEnvironment {
  @Singular("addAccount")
  private final List<ToyAccount> accounts;

  private final List<BlockSnapshot> blocks;

  /**
   * A transaction validator of each transaction; by default, it asserts that the transaction was
   * successfully processed.
   */
  @Builder.Default
  private final TransactionProcessingResultValidator transactionProcessingResultValidator =
      TransactionProcessingResultValidator.DEFAULT_VALIDATOR;

  public static class MultiBlockExecutionEnvironmentBuilder {

    private List<BlockSnapshot> blocks = new ArrayList<>();

    public MultiBlockExecutionEnvironmentBuilder addBlock(List<Transaction> transactions) {
      return addBlock(transactions, LINEA_BLOCK_GAS_LIMIT);
    }

    public MultiBlockExecutionEnvironmentBuilder addBlock(
        List<Transaction> transactions, long gasLimit) {
      BlockHeaderBuilder blockHeaderBuilder =
          this.blocks.isEmpty()
              ? ExecutionEnvironment.getLineaBlockHeaderBuilder(Optional.empty())
              : ExecutionEnvironment.getLineaBlockHeaderBuilder(
                  Optional.of(this.blocks.getLast().header().toBlockHeader()));
      blockHeaderBuilder.coinbase(ToyExecutionEnvironmentV2.DEFAULT_COINBASE_ADDRESS);
      blockHeaderBuilder.gasLimit(gasLimit);
      BlockBody blockBody = new BlockBody(transactions, Collections.emptyList());
      this.blocks.add(BlockSnapshot.of(blockHeaderBuilder.buildBlockHeader(), blockBody));

      return this;
    }
  }

  public void run() {
    ReplayExecutionEnvironment.builder()
        .zkTracer(new ZkTracer(ToyExecutionEnvironmentV2.CHAIN_ID))
        .useCoinbaseAddressFromBlockHeader(true)
        .transactionProcessingResultValidator(this.transactionProcessingResultValidator)
        .build()
        .replay(ToyExecutionEnvironmentV2.CHAIN_ID, this.buildConflationSnapshot());
  }

  private ConflationSnapshot buildConflationSnapshot() {
    List<AccountSnapshot> accountSnapshots =
        accounts.stream()
            .map(
                toyAccount ->
                    new AccountSnapshot(
                        toyAccount.getAddress().toHexString(),
                        toyAccount.getNonce(),
                        toyAccount.getBalance().toHexString(),
                        toyAccount.getCode().toHexString()))
            .toList();

    List<StorageSnapshot> storageSnapshots =
        accounts.stream()
            .flatMap(
                account ->
                    account.storage.entrySet().stream()
                        .map(
                            storageEntry ->
                                new StorageSnapshot(
                                    account.getAddress().toHexString(),
                                    storageEntry.getKey().toHexString(),
                                    storageEntry.getValue().toHexString())))
            .toList();

    List<BlockHashSnapshot> blockHashSnapshots =
        blocks.stream()
            .map(
                blockSnapshot -> {
                  BlockHeader blockHeader = blockSnapshot.header().toBlockHeader();
                  return BlockHashSnapshot.of(blockHeader.getNumber(), blockHeader.getBlockHash());
                })
            .toList();

    return new ConflationSnapshot(
        this.blocks, accountSnapshots, storageSnapshots, blockHashSnapshots);
  }
}
