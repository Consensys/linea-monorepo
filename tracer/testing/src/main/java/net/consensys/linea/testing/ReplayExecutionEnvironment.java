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

import static org.assertj.core.api.Assertions.assertThat;

import java.io.IOException;
import java.io.Reader;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;

import com.google.gson.Gson;
import lombok.Builder;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.blockcapture.snapshots.AccountSnapshot;
import net.consensys.linea.blockcapture.snapshots.BlockSnapshot;
import net.consensys.linea.blockcapture.snapshots.ConflationSnapshot;
import net.consensys.linea.blockcapture.snapshots.StorageSnapshot;
import net.consensys.linea.blockcapture.snapshots.TransactionResultSnapshot;
import net.consensys.linea.blockcapture.snapshots.TransactionSnapshot;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.zktracer.ConflationAwareOperationTracer;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.module.hub.Hub;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.config.GenesisConfigFile;
import org.hyperledger.besu.config.GenesisConfigOptions;
import org.hyperledger.besu.datatypes.*;
import org.hyperledger.besu.ethereum.chain.BadBlockManager;
import org.hyperledger.besu.ethereum.core.*;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.mainnet.MainnetProtocolSchedule;
import org.hyperledger.besu.ethereum.mainnet.MainnetProtocolSpecFactory;
import org.hyperledger.besu.ethereum.mainnet.MainnetTransactionProcessor;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSchedule;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSpecBuilder;
import org.hyperledger.besu.ethereum.processing.TransactionProcessingResult;
import org.hyperledger.besu.ethereum.referencetests.ReferenceTestWorldState;
import org.hyperledger.besu.evm.account.MutableAccount;
import org.hyperledger.besu.evm.internal.EvmConfiguration;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.operation.BlockHashOperation;
import org.hyperledger.besu.evm.tracing.OperationTracer;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem;
import org.hyperledger.besu.plugin.data.BlockHeader;

/** Responsible for executing EVM transactions in replay tests. */
@Builder
@Slf4j
public class ReplayExecutionEnvironment {
  /** Chain ID for Linea mainnet */
  public static final BigInteger LINEA_MAINNET = BigInteger.valueOf(59144);

  /** Chain ID for Linea sepolia */
  public static final BigInteger LINEA_SEPOLIA = BigInteger.valueOf(59141);

  /** Used for checking resulting trace files. */
  private static final CorsetValidator CORSET_VALIDATOR = new CorsetValidator();

  /**
   * Determines whether transaction results should be checked against expected results embedded in
   * replay files. This gives an additional level of assurance that the tests properly reflect
   * mainnet (or e.g. sepolia as appropriate). When this is set to false, replay tests will not fail
   * even if the tx outcome differs from what actually occurred on the relevant chain (e.g.
   * mainnet). Such scenarios have arisen, for example, when <code>ZkTracer</code> has an unexpected
   * side-effect on tx execution (which it never should).
   */
  private final boolean txResultChecking;

  private final ZkTracer tracer = new ZkTracer();

  public void checkTracer() {
    try {
      final Path traceFile = Files.createTempFile(null, ".lt");
      this.tracer.writeToFile(traceFile);
      log.info("trace written to `{}`", traceFile);
      assertThat(CORSET_VALIDATOR.validate(traceFile).isValid()).isTrue();
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  public void checkTracer(String inputFilePath) {
    // Generate the output file path based on the input file path
    Path inputPath = Paths.get(inputFilePath);
    String outputFileName = inputPath.getFileName().toString().replace(".json.gz", ".lt");
    Path outputPath = inputPath.getParent().resolve(outputFileName);
    this.tracer.writeToFile(outputPath);
    log.info("trace written to `{}`", outputPath);
    // validation is disabled by default for replayBulk
    // assertThat(CORSET_VALIDATOR.validate(outputPath).isValid()).isTrue();
  }

  /**
   * Given a file containing the JSON serialization of a {@link ConflationSnapshot}, loads it,
   * updates this's state to mirror it, and replays it.
   *
   * @param replayFile the file containing the conflation
   */
  public void replay(BigInteger chainId, final Reader replayFile) {
    Gson gson = new Gson();
    ConflationSnapshot conflation;
    try {
      conflation = gson.fromJson(replayFile, ConflationSnapshot.class);
    } catch (Exception e) {
      log.error(e.getMessage());
      return;
    }
    this.executeFrom(chainId, conflation);
    this.checkTracer();
  }

  public void replay(BigInteger chainId, final Reader replayFile, String inputFilePath) {
    Gson gson = new Gson();
    ConflationSnapshot conflation;
    try {
      conflation = gson.fromJson(replayFile, ConflationSnapshot.class);
    } catch (Exception e) {
      log.error(e.getMessage());
      return;
    }
    this.executeFrom(chainId, conflation);
    this.checkTracer(inputFilePath);
  }

  /**
   * Loads the states and the conflation defined in a {@link ConflationSnapshot}, mimick the
   * accounts, storage and blocks state as it was on the blockchain before the conflation played
   * out, then execute and check it.
   *
   * @param conflation the conflation to replay
   */
  private void executeFrom(final BigInteger chainId, final ConflationSnapshot conflation) {
    BlockHashOperation.BlockHashLookup blockHashLookup = conflation.toBlockHashLookup();
    ReferenceTestWorldState world =
        ReferenceTestWorldState.create(new HashMap<>(), EvmConfiguration.DEFAULT);
    // Initialise world state from conflation
    initWorld(world.updater(), conflation);
    // Construct the transaction processor
    final MainnetTransactionProcessor transactionProcessor =
        getMainnetTransactionProcessor(chainId);
    // Begin
    tracer.traceStartConflation(conflation.blocks().size());
    //
    for (BlockSnapshot blockSnapshot : conflation.blocks()) {
      final BlockHeader header = blockSnapshot.header().toBlockHeader();

      final BlockBody body =
          new BlockBody(
              blockSnapshot.txs().stream().map(TransactionSnapshot::toTransaction).toList(),
              new ArrayList<>());
      tracer.traceStartBlock(header, body);

      for (TransactionSnapshot txs : blockSnapshot.txs()) {
        final Transaction tx = txs.toTransaction();
        final WorldUpdater updater = world.updater();
        // Process transaction leading to expected outcome
        final TransactionProcessingResult outcome =
            transactionProcessor.processTransaction(
                updater,
                (ProcessableBlockHeader) header,
                tx,
                header.getCoinbase(),
                buildOperationTracer(tx, txs.getOutcome()),
                blockHashLookup,
                false,
                Wei.ZERO);
        // Commit transaction
        updater.commit();
      }
      tracer.traceEndBlock(header, body);
    }
    tracer.traceEndConflation(world.updater());
  }

  /**
   * Construct transaction processor a given chain (e.g. 59144 for mainnet, 59141 for sepolia, etc).
   *
   * @return
   */
  private MainnetTransactionProcessor getMainnetTransactionProcessor(BigInteger chainId) {
    EvmConfiguration evmConfig = EvmConfiguration.DEFAULT;
    // Read genesis config for linea
    GenesisConfigFile configFile =
        GenesisConfigFile.fromSource(GenesisConfigFile.class.getResource("/linea.json"));
    GenesisConfigOptions options = configFile.getConfigOptions();
    BadBlockManager badBlockManager = new BadBlockManager();
    // Create the schedule
    ProtocolSchedule schedule =
        MainnetProtocolSchedule.fromConfig(
            options,
            MiningParameters.MINING_DISABLED,
            badBlockManager,
            false,
            new NoOpMetricsSystem());
    // Create
    ProtocolSpecBuilder builder =
        new MainnetProtocolSpecFactory(
                Optional.of(chainId),
                true,
                OptionalLong.empty(),
                evmConfig,
                MiningParameters.MINING_DISABLED,
                false,
                new NoOpMetricsSystem())
            .londonDefinition(options); // .lineaOpCodesDefinition(options);
    //
    builder.privacyParameters(PrivacyParameters.DEFAULT);
    builder.badBlocksManager(badBlockManager);
    //
    return builder.build(schedule).getTransactionProcessor();
  }

  public Hub getHub() {
    return tracer.getHub();
  }

  /**
   * Initialise a world updater given a conflation. Observe this can be applied to any WorldUpdater,
   * such as SimpleWorld.
   *
   * @param world The world to be initialised.
   * @param conflation The conflation from which to initialise.
   */
  private static void initWorld(WorldUpdater world, final ConflationSnapshot conflation) {
    WorldUpdater updater = world.updater();
    for (AccountSnapshot account : conflation.accounts()) {
      // Construct contract address
      Address addr = Address.fromHexString(account.address());
      // Create account
      MutableAccount acc =
          world.createAccount(
              Words.toAddress(addr), account.nonce(), Wei.fromHexString(account.balance()));
      // Update code
      acc.setCode(Bytes.fromHexString(account.code()));
    }
    // Initialise storage
    for (StorageSnapshot s : conflation.storage()) {
      world
          .getAccount(Words.toAddress(Bytes.fromHexString(s.address())))
          .setStorageValue(UInt256.fromHexString(s.key()), UInt256.fromHexString(s.value()));
    }
    //
    world.commit();
  }

  /**
   * Construct an operation tracer which invokes the zkTracer appropriately, and can also
   * (optionally) check the transaction outcome matches what was expected.
   *
   * @param tx Transaction being processed
   * @param txs TransactionResultSnapshot which contains the expected result of this transaction.
   * @return An implementation of OperationTracer which packages up the appropriate behavour.
   */
  private OperationTracer buildOperationTracer(Transaction tx, TransactionResultSnapshot txs) {
    if (txs == null) {
      String hash = tx.getHash().toHexString();
      log.info("tx `{}` outcome not checked (missing)", hash);
      return tracer;
    } else if (!txResultChecking) {
      String hash = tx.getHash().toHexString();
      log.info("tx `{}` outcome not checked (disabled)", hash);
      return tracer;
    } else {
      return ConflationAwareOperationTracer.sequence(txs.check(), tracer);
    }
  }
}
