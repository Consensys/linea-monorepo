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

import java.io.File;
import java.io.IOException;
import java.io.Reader;
import java.math.BigInteger;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;

import com.google.gson.Gson;
import lombok.Builder;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.blockcapture.BlockCapturer;
import net.consensys.linea.blockcapture.snapshots.AccountSnapshot;
import net.consensys.linea.blockcapture.snapshots.BlockSnapshot;
import net.consensys.linea.blockcapture.snapshots.ConflationSnapshot;
import net.consensys.linea.blockcapture.snapshots.StorageSnapshot;
import net.consensys.linea.blockcapture.snapshots.TransactionResultSnapshot;
import net.consensys.linea.blockcapture.snapshots.TransactionSnapshot;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.zktracer.ConflationAwareOperationTracer;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.module.constants.GlobalConstants;
import net.consensys.linea.zktracer.module.hub.Hub;
import org.apache.commons.io.FileUtils;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.consensus.clique.CliqueHelpers;
import org.hyperledger.besu.datatypes.*;
import org.hyperledger.besu.ethereum.core.*;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.mainnet.MainnetTransactionProcessor;
import org.hyperledger.besu.ethereum.processing.TransactionProcessingResult;
import org.hyperledger.besu.ethereum.referencetests.ReferenceTestWorldState;
import org.hyperledger.besu.evm.account.MutableAccount;
import org.hyperledger.besu.evm.blockhash.BlockHashLookup;
import org.hyperledger.besu.evm.internal.EvmConfiguration;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.tracing.OperationTracer;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;

/** Responsible for executing EVM transactions in replay tests. */
@Builder
@Slf4j
public class ReplayExecutionEnvironment {
  /** Chain ID for Linea mainnet */
  public static final BigInteger LINEA_MAINNET = BigInteger.valueOf(GlobalConstants.LINEA_CHAIN_ID);

  /** Chain ID for Linea sepolia */
  public static final BigInteger LINEA_SEPOLIA =
      BigInteger.valueOf(GlobalConstants.LINEA_SEPOLIA_CHAIN_ID);

  /** Used for checking resulting trace files. */
  private static final CorsetValidator CORSET_VALIDATOR = new CorsetValidator();

  /**
   * Determines whether to enable block capturing for conflations executed by this environment. This
   * is used for primarily for debugging the block capturer.
   */
  private static final boolean debugBlockCapturer = false;

  /**
   * Determines whether transaction results should be checked against expected results embedded in
   * replay files. This gives an additional level of assurance that the tests properly reflect
   * mainnet (or e.g. sepolia as appropriate). When this is set to false, replay tests will not fail
   * even if the tx outcome differs from what actually occurred on the relevant chain (e.g.
   * mainnet). Such scenarios have arisen, for example, when <code>ZkTracer</code> has an unexpected
   * side-effect on tx execution (which it never should).
   */
  private final boolean txResultChecking;

  /**
   * Setting this to true will disable the clique consensus protocol for parsing coinbase address
   * from block header. This is needed for manual tests like the Multi Block tests which do not have
   * encoded coinbase address in the block header.
   */
  @Builder.Default private final boolean useCoinbaseAddressFromBlockHeader = false;

  /** A transaction validator of each transaction; by default, it does not do anything. */
  @Builder.Default
  private final TransactionProcessingResultValidator transactionProcessingResultValidator =
      TransactionProcessingResultValidator.EMPTY_VALIDATOR;

  private ZkTracer zkTracer;

  public void checkTracer(String inputFilePath) {
    // Generate the output file path based on the input file path
    Path inputPath = Paths.get(inputFilePath);
    String outputFileName = inputPath.getFileName().toString().replace(".json.gz", ".lt");
    Path outputPath = inputPath.getParent().resolve(outputFileName);
    this.zkTracer.writeToFile(outputPath);
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
    ExecutionEnvironment.checkTracer(zkTracer, CORSET_VALIDATOR, Optional.of(log));
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

  public void replay(BigInteger chainId, ConflationSnapshot conflation) {
    this.executeFrom(chainId, conflation);
    ExecutionEnvironment.checkTracer(zkTracer, CORSET_VALIDATOR, Optional.of(log));
  }

  /**
   * Loads the states and the conflation defined in a {@link ConflationSnapshot}, mimick the
   * accounts, storage and blocks state as it was on the blockchain before the conflation played
   * out, then execute and check it.
   *
   * @param conflation the conflation to replay
   */
  private void executeFrom(final BigInteger chainId, final ConflationSnapshot conflation) {
    ConflationAwareOperationTracer tracer = this.zkTracer;
    BlockCapturer capturer = null;
    // Configure block capturer (if applicable)
    if (debugBlockCapturer) {
      // Initialise world state from conflation
      MutableWorldState world = initWorld(conflation);
      capturer = new BlockCapturer();
      capturer.setWorld(world.updater());
      // Sequence zktracer and capturer
      tracer = ConflationAwareOperationTracer.sequence(tracer, capturer);
    }
    // Execute the conflation
    executeFrom(
        chainId,
        conflation,
        tracer,
        this.txResultChecking,
        this.useCoinbaseAddressFromBlockHeader,
        this.transactionProcessingResultValidator);
    //
    if (debugBlockCapturer) {
      writeCaptureToFile(chainId, conflation, capturer);
    }
  }

  private static void executeFrom(
      final BigInteger chainId,
      final ConflationSnapshot conflation,
      final ConflationAwareOperationTracer tracer,
      final boolean txResultChecking,
      final boolean useCoinbaseAddressFromBlockHeader,
      final TransactionProcessingResultValidator resultValidator) {
    BlockHashLookup blockHashLookup = conflation.toBlockHashLookup();
    // Initialise world state from conflation
    MutableWorldState world = initWorld(conflation);
    // Construct the transaction processor
    final MainnetTransactionProcessor transactionProcessor =
        ExecutionEnvironment.getProtocolSpec(chainId).getTransactionProcessor();
    // Begin
    tracer.traceStartConflation(conflation.blocks().size());
    //
    for (BlockSnapshot blockSnapshot : conflation.blocks()) {
      final BlockHeader header = blockSnapshot.header().toBlockHeader();

      final BlockBody body =
          new BlockBody(
              blockSnapshot.txs().stream().map(TransactionSnapshot::toTransaction).toList(),
              new ArrayList<>());
      final Address miningBeneficiary =
          useCoinbaseAddressFromBlockHeader
              ? header.getCoinbase()
              : CliqueHelpers.getProposerOfBlock(header);
      tracer.traceStartBlock(header, miningBeneficiary);

      for (TransactionSnapshot txs : blockSnapshot.txs()) {
        final Transaction tx = txs.toTransaction();
        final WorldUpdater updater = world.updater();
        // Process transaction leading to expected outcome
        final TransactionProcessingResult outcome =
            transactionProcessor.processTransaction(
                updater,
                header,
                tx,
                miningBeneficiary,
                buildOperationTracer(tx, txs.getOutcome(), tracer, txResultChecking),
                blockHashLookup,
                false,
                Wei.ZERO);
        resultValidator.accept(tx, outcome);
        // Commit transaction
        updater.commit();
      }
      tracer.traceEndBlock(header, body);
    }
    tracer.traceEndConflation(world.updater());
  }

  public Hub getHub() {
    return zkTracer.getHub();
  }

  /**
   * Initialise a fresh world state from a conflation.
   *
   * @param conflation The conflation from which to initialise.
   */
  private static MutableWorldState initWorld(final ConflationSnapshot conflation) {
    ReferenceTestWorldState world =
        ReferenceTestWorldState.create(new HashMap<>(), EvmConfiguration.DEFAULT);
    WorldUpdater updater = world.updater();
    for (AccountSnapshot account : conflation.accounts()) {
      // Construct contract address
      Address addr = Address.fromHexString(account.address());
      // Create account
      MutableAccount acc =
          updater.createAccount(
              Words.toAddress(addr), account.nonce(), Wei.fromHexString(account.balance()));
      // Update code
      acc.setCode(Bytes.fromHexString(account.code()));
    }
    // Initialise storage
    for (StorageSnapshot s : conflation.storage()) {
      UInt256 key = UInt256.fromHexString(s.key());
      UInt256 value = UInt256.fromHexString(s.value());
      // The following check is only necessary because of older replay files which captured storage
      // for accounts created in the conflation itself (see #1289).  Such assignments are always
      // zero values, but this confuses BESU into thinking their storage is not empty (leading to a
      // creation failure).  This fix simply prevents zero values from being assigned at all.
      // If/when all older replay files are recaptured, then this check should be redundant.
      if (!value.isZero()) {
        updater
            .getAccount(Words.toAddress(Bytes.fromHexString(s.address())))
            .setStorageValue(key, value);
      }
    }
    // Commit changes
    updater.commit();
    // Done
    return world;
  }

  /**
   * Construct an operation tracer which invokes the zkTracer appropriately, and can also
   * (optionally) check the transaction outcome matches what was expected.
   *
   * @param tx Transaction being processed
   * @param txs TransactionResultSnapshot which contains the expected result of this transaction.
   * @return An implementation of OperationTracer which packages up the appropriate behavour.
   */
  private static OperationTracer buildOperationTracer(
      Transaction tx,
      TransactionResultSnapshot txs,
      ConflationAwareOperationTracer tracer,
      boolean txResultChecking) {
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

  // Write the captured replay for a given conflation snapshot to a file.  This is used to debug the
  // BlockCapturer by making sure, for example, that captured replays still execute correctly.
  private static void writeCaptureToFile(
      BigInteger chainId, ConflationSnapshot conflation, BlockCapturer capturer) {
    // Extract capture name
    String json = capturer.toJson();
    // Determine suitable filename
    long startBlock = Long.MAX_VALUE;
    long endBlock = Long.MIN_VALUE;
    //
    for (BlockSnapshot blk : conflation.blocks()) {
      startBlock = Math.min(startBlock, blk.header().number());
      endBlock = Math.max(endBlock, blk.header().number());
    }
    // Convert ChainID to something useful
    String chain = getChainName(chainId);
    // Construct suitable filename for captured conflation.
    String filename =
        startBlock == endBlock
            ? String.format("capture-%d.%s.json", startBlock, chain)
            : String.format("capture-%d-%d.%s.json", startBlock, endBlock, chain);
    // Write the conflation.
    try {
      File file = new File(filename);
      log.info("Writing capture to " + file.getCanonicalPath());
      FileUtils.writeStringToFile(file, json);
    } catch (IOException e) {
      // Problem writing capture
      throw new RuntimeException(e);
    }
  }

  /**
   * Convert a chainID into a human-readable string.
   *
   * @param chainId
   * @return
   */
  private static String getChainName(BigInteger chainId) {
    if (chainId.equals(LINEA_MAINNET)) {
      return "mainnet";
    } else if (chainId.equals(LINEA_SEPOLIA)) {
      return "sepolia";
    } else {
      return String.format("chain%s", chainId.toString());
    }
  }
}
