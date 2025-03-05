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

import static net.consensys.linea.zktracer.Trace.*;
import static org.assertj.core.api.Assertions.assertThat;

import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Optional;
import java.util.OptionalLong;

import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.zktracer.ZkTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.config.GenesisConfig;
import org.hyperledger.besu.config.GenesisConfigOptions;
import org.hyperledger.besu.consensus.clique.CliqueBlockHeaderFunctions;
import org.hyperledger.besu.consensus.clique.CliqueForksSchedulesFactory;
import org.hyperledger.besu.consensus.clique.CliqueProtocolSchedule;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SignatureAlgorithm;
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory;
import org.hyperledger.besu.cryptoservices.KeyPairSecurityModule;
import org.hyperledger.besu.cryptoservices.NodeKey;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.chain.BadBlockManager;
import org.hyperledger.besu.ethereum.core.BlockHeader;
import org.hyperledger.besu.ethereum.core.BlockHeaderBuilder;
import org.hyperledger.besu.ethereum.core.Difficulty;
import org.hyperledger.besu.ethereum.core.MiningConfiguration;
import org.hyperledger.besu.ethereum.core.PrivacyParameters;
import org.hyperledger.besu.ethereum.mainnet.MainnetProtocolSpecFactory;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSchedule;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSpec;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSpecBuilder;
import org.hyperledger.besu.evm.internal.EvmConfiguration;
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem;
import org.slf4j.Logger;

public class ExecutionEnvironment {
  public static final String CORSET_VALIDATION_RESULT = "Corset validation result: ";

  static GenesisConfig GENESIS_CONFIG =
      GenesisConfig.fromSource(ExecutionEnvironment.class.getResource("/linea.json"));

  static final BlockHeaderBuilder DEFAULT_BLOCK_HEADER_BUILDER =
      BlockHeaderBuilder.createDefault()
          .number(ToyExecutionEnvironmentV2.DEFAULT_BLOCK_NUMBER)
          .timestamp(123456789)
          .parentHash(Hash.EMPTY_TRIE_HASH)
          .baseFee(ToyExecutionEnvironmentV2.DEFAULT_BASE_FEE)
          .nonce(0)
          .blockHeaderFunctions(new CliqueBlockHeaderFunctions());

  public static void checkTracer(
      ZkTracer zkTracer,
      CorsetValidator corsetValidator,
      Optional<Logger> logger,
      long startBlock,
      long endBlock) {
    Path traceFilePath = null;
    boolean traceValidated = false;
    try {
      String prefix = constructTestPrefix();
      traceFilePath = Files.createTempFile(prefix, ".lt");
      zkTracer.writeToFile(traceFilePath, startBlock, endBlock);
      final Path finalTraceFilePath = traceFilePath;
      logger.ifPresent(log -> log.debug("trace written to {}", finalTraceFilePath));
      CorsetValidator.Result corsetValidationResult = corsetValidator.validate(traceFilePath);
      traceValidated = corsetValidationResult.isValid();
      assertThat(traceValidated)
          .withFailMessage(CORSET_VALIDATION_RESULT + "%s", corsetValidationResult.corsetOutput())
          .isTrue();
    } catch (IOException e) {
      throw new RuntimeException(e);
    } finally {
      if (traceFilePath != null && traceValidated) {
        if (System.getenv("PRESERVE_TRACE_FILES") == null) {
          boolean traceFileDeleted = traceFilePath.toFile().delete();
          final Path finalTraceFilePath = traceFilePath;
          logger.ifPresent(
              log -> log.debug("trace file {} deleted {}", finalTraceFilePath, traceFileDeleted));
        }
      }
    }
  }

  public static BlockHeaderBuilder getLineaBlockHeaderBuilder(
      Optional<BlockHeader> parentBlockHeader) {

    BlockHeaderBuilder blockHeaderBuilder =
        parentBlockHeader.isPresent()
            ? BlockHeaderBuilder.fromHeader(parentBlockHeader.get())
                .number(parentBlockHeader.get().getNumber() + 1)
                .timestamp(parentBlockHeader.get().getTimestamp() + 100)
                .parentHash(parentBlockHeader.get().getHash())
                .nonce(parentBlockHeader.get().getNonce() + 1)
                .blockHeaderFunctions(new CliqueBlockHeaderFunctions())
            : DEFAULT_BLOCK_HEADER_BUILDER;

    return blockHeaderBuilder
        .baseFee(Wei.of(LINEA_BASE_FEE))
        // TODO: refacto this block gas limit
        .gasLimit(LINEA_BLOCK_GAS_LIMIT)
        .difficulty(Difficulty.of(LINEA_DIFFICULTY));
  }

  public static ProtocolSpec getProtocolSpec(BigInteger chainId) {
    BadBlockManager badBlockManager = new BadBlockManager();
    final GenesisConfigOptions genesisConfigOptions = GENESIS_CONFIG.getConfigOptions();

    ProtocolSchedule schedule =
        CliqueProtocolSchedule.create(
            genesisConfigOptions,
            CliqueForksSchedulesFactory.create(genesisConfigOptions),
            createNodeKey(),
            PrivacyParameters.DEFAULT,
            false,
            EvmConfiguration.DEFAULT,
            MiningConfiguration.MINING_DISABLED,
            badBlockManager,
            false,
            new NoOpMetricsSystem());

    ProtocolSpecBuilder builder =
        new MainnetProtocolSpecFactory(
                Optional.of(chainId),
                true,
                OptionalLong.empty(),
                EvmConfiguration.DEFAULT,
                MiningConfiguration.MINING_DISABLED,
                false,
                new NoOpMetricsSystem())
            .londonDefinition(GENESIS_CONFIG.getConfigOptions());
    // .lineaOpCodesDefinition(GENESIS_CONFIG.getConfigOptions());

    return builder
        .privacyParameters(PrivacyParameters.DEFAULT)
        .badBlocksManager(badBlockManager)
        .build(schedule);
  }

  private static NodeKey createNodeKey() {
    final Bytes32 keyPairPrvKey =
        Bytes32.fromHexString("0xf7a58d5e755d51fa2f6206e91dd574597c73248aaf946ec1964b8c6268d6207b");
    final SignatureAlgorithm signatureAlgorithm = SignatureAlgorithmFactory.getInstance();
    final KeyPair keyPair =
        signatureAlgorithm.createKeyPair(signatureAlgorithm.createPrivateKey(keyPairPrvKey));
    final KeyPairSecurityModule keyPairSecurityModule = new KeyPairSecurityModule(keyPair);

    return new NodeKey(keyPairSecurityModule);
  }

  private static final String LINEA_PACKAGE = "net.consensys.linea.";

  /**
   * Construct a suitable prefix for the temporary lt file generated based on the method name of the
   * test. This is done by walking up the stack looking for a calling method whose classname ends
   * with "Test". Having found such a method, its name is then used as the test prefix. If no method
   * is found, then this simply returns null --- which is completely safe in this context.
   *
   * @return
   */
  public static String constructTestPrefix() {
    for (StackTraceElement ste : Thread.currentThread().getStackTrace()) {
      String className = ste.getClassName();
      if (className.endsWith("Test") || className.endsWith("Tests")) {
        // Yes, it is.  Now tidy up the name.
        String name = ste.getClassName().replace(LINEA_PACKAGE, "").replace(".", "_");
        // Done
        return name + "_" + ste.getMethodName() + "_";
      }
    }
    // Failed, so return null.  This is fine as it just means the generate lt file will not have an
    // informative
    // prefix.
    return null;
  }
}
