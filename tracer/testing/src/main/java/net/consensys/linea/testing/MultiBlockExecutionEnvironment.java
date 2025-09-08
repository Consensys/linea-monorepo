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

import static net.consensys.linea.testing.ToyExecutionEnvironmentV2.DEFAULT_BLOCK_NUMBER;
import static net.consensys.linea.zktracer.Trace.LINEA_BLOCK_GAS_LIMIT;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Optional;

import lombok.Builder;
import lombok.Singular;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.blockcapture.snapshots.*;
import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.module.hub.Hub;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.ethereum.core.*;
import org.junit.jupiter.api.TestInfo;

@Builder
@Slf4j
public class MultiBlockExecutionEnvironment {
  public static final short DEFAULT_DELTA_TIMESTAMP_BETWEEN_BLOCKS = 2;

  @Singular("addAccount")
  private final List<ToyAccount> accounts;

  private final List<BlockSnapshot> blocks;

  public static final BigInteger CHAIN_ID = BigInteger.valueOf(1337);
  private final ZkTracer tracer;

  public final ChainConfig testsChain;
  public final TestInfo testInfo;

  @Builder.Default private final long startingBlockNumber = DEFAULT_BLOCK_NUMBER;
  @Builder.Default private final boolean systemContractDeployedPriorToConflation = true;

  /**
   * A transaction validator of each transaction; by default, it asserts that the transaction was
   * successfully processed.
   */
  @Builder.Default
  private final TransactionProcessingResultValidator transactionProcessingResultValidator =
      TransactionProcessingResultValidator.DEFAULT_VALIDATOR;

  public static MultiBlockExecutionEnvironment.MultiBlockExecutionEnvironmentBuilder builder(
      ChainConfig chainConfig, TestInfo testInfo) {
    return new MultiBlockExecutionEnvironmentBuilder()
        .tracer(new ZkTracer(chainConfig))
        .testsChain(chainConfig)
        .testInfo(testInfo);
  }

  public static MultiBlockExecutionEnvironment.MultiBlockExecutionEnvironmentBuilder builder(
      ChainConfig chainConfig,
      TestInfo testInfo,
      boolean systemContractDeployedPriorConflation,
      long firstBlockNumber) {
    return new MultiBlockExecutionEnvironmentBuilder()
        .tracer(new ZkTracer(chainConfig))
        .testsChain(chainConfig)
        .testInfo(testInfo)
        .systemContractDeployedPriorToConflation(systemContractDeployedPriorConflation)
        .startingBlockNumber(firstBlockNumber);
  }

  public static class MultiBlockExecutionEnvironmentBuilder {

    private List<BlockSnapshot> blocks = new ArrayList<>();

    public MultiBlockExecutionEnvironmentBuilder addBlock(List<Transaction> transactions) {
      return addBlock(transactions, LINEA_BLOCK_GAS_LIMIT);
    }

    public MultiBlockExecutionEnvironmentBuilder addBlock(
        List<Transaction> transactions, long gasLimit) {
      final boolean firstBlock = this.blocks.isEmpty();
      final BlockHeaderBuilder blockHeaderBuilder =
          firstBlock
              ? ExecutionEnvironment.getLineaBlockHeaderBuilder(Optional.empty())
              : ExecutionEnvironment.getLineaBlockHeaderBuilder(
                  Optional.of(this.blocks.getLast().header().toBlockHeader()));
      blockHeaderBuilder.coinbase(ToyExecutionEnvironmentV2.DEFAULT_COINBASE_ADDRESS);
      blockHeaderBuilder.gasLimit(gasLimit);
      blockHeaderBuilder.number(startingBlockNumber$value + blocks.size());
      // Note: as per https://eips.ethereum.org/EIPS/eip-4788: "If this EIP is active in a genesis
      // block, the genesis header’s parent_beacon_block_root must be 0x0 and no system transaction
      // may occur."
      if (firstBlock) {
        blockHeaderBuilder.parentBeaconBlockRoot(
            startingBlockNumber$value == 0 ? Bytes32.ZERO : Bytes32.fromHexString("0xBADDADD7"));
      } else {
        blockHeaderBuilder.parentBeaconBlockRoot(blocks.getLast().header().parentBeaconBlockRoot());
      }

      final BlockBody blockBody = new BlockBody(transactions, Collections.emptyList());
      this.blocks.add(BlockSnapshot.of(blockHeaderBuilder.buildBlockHeader(), blockBody));

      return this;
    }
  }

  public void run() {
    ReplayExecutionEnvironment.builder()
        .zkTracer(tracer)
        .useCoinbaseAddressFromBlockHeader(true)
        .transactionProcessingResultValidator(this.transactionProcessingResultValidator)
        .systemContractDeployedPriorToConflation(systemContractDeployedPriorToConflation)
        .build()
        .replay(testsChain, testInfo, this.buildConflationSnapshot());
  }

  public Hub getHub() {
    return tracer.getHub();
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
