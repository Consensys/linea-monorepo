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

import java.math.BigInteger;
import java.util.Optional;
import java.util.OptionalLong;

import org.hyperledger.besu.config.GenesisConfigFile;
import org.hyperledger.besu.ethereum.chain.BadBlockManager;
import org.hyperledger.besu.ethereum.core.MiningParameters;
import org.hyperledger.besu.ethereum.core.PrivacyParameters;
import org.hyperledger.besu.ethereum.mainnet.MainnetProtocolSchedule;
import org.hyperledger.besu.ethereum.mainnet.MainnetProtocolSpecFactory;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSchedule;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSpec;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSpecBuilder;
import org.hyperledger.besu.evm.internal.EvmConfiguration;
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem;

public class ExecutionEnvironment {
  static GenesisConfigFile GENESIS_CONFIG =
      GenesisConfigFile.fromSource(GenesisConfigFile.class.getResource("/linea.json"));

  public static ProtocolSpec getProtocolSpec(BigInteger chainId) {
    BadBlockManager badBlockManager = new BadBlockManager();

    ProtocolSchedule schedule =
        MainnetProtocolSchedule.fromConfig(
            GENESIS_CONFIG.getConfigOptions(),
            MiningParameters.MINING_DISABLED,
            badBlockManager,
            false,
            new NoOpMetricsSystem());

    ProtocolSpecBuilder builder =
        new MainnetProtocolSpecFactory(
                Optional.of(chainId),
                true,
                OptionalLong.empty(),
                EvmConfiguration.DEFAULT,
                MiningParameters.MINING_DISABLED,
                false,
                new NoOpMetricsSystem())
            .londonDefinition(GENESIS_CONFIG.getConfigOptions());
    // .lineaOpCodesDefinition(GENESIS_CONFIG.getConfigOptions());

    builder.privacyParameters(PrivacyParameters.DEFAULT);
    builder.badBlocksManager(badBlockManager);
    return builder.build(schedule);
  }
}
