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
package net.consensys.linea.zktracer;

import static net.consensys.linea.zktracer.Trace.ETHEREUM_GAS_LIMIT_MAXIMUM;
import static net.consensys.linea.zktracer.Trace.ETHEREUM_GAS_LIMIT_MINIMUM;
import static net.consensys.linea.zktracer.Trace.LINEA_CHAIN_ID;
import static net.consensys.linea.zktracer.Trace.LINEA_GAS_LIMIT_MAXIMUM;
import static net.consensys.linea.zktracer.Trace.LINEA_GAS_LIMIT_MINIMUM;
import static net.consensys.linea.zktracer.Trace.LINEA_SEPOLIA_CHAIN_ID;

import java.math.BigInteger;

import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;

/**
 * Provides s single point of configuration for running the tracer on different chains which, aside
 * from Linea mainnet, are all currently for the purposes of testing.
 */
public class ChainConfig {
  /**
   * Represents Linea mainnet as it stands today which enforces the block gas limit (currently two
   * billion). As the name suggest, this is only intended for testing purposes.
   */
  public static final ChainConfig MAINNET_TESTCONFIG =
      new ChainConfig(
          LINEA_CHAIN_ID,
          true,
          BigInteger.valueOf(LINEA_GAS_LIMIT_MINIMUM),
          BigInteger.valueOf(LINEA_GAS_LIMIT_MAXIMUM),
          LineaL1L2BridgeSharedConfiguration.TEST_DEFAULT);

  /** Represents Ethereum mainnet for the purposes of running reference tests. */
  public static final ChainConfig ETHEREUM =
      new ChainConfig(
          1,
          false,
          BigInteger.valueOf(ETHEREUM_GAS_LIMIT_MINIMUM),
          ETHEREUM_GAS_LIMIT_MAXIMUM,
          LineaL1L2BridgeSharedConfiguration.TEST_DEFAULT);

  /**
   * Represents Linea mainnet prior to the block gas limit being enforced for the purposes of
   * running existing replay tests. As the name suggest, this is only intended for testing purposes.
   */
  public static final ChainConfig OLD_MAINNET_TESTCONFIG =
      new ChainConfig(
          LINEA_CHAIN_ID,
          false,
          BigInteger.valueOf(LINEA_GAS_LIMIT_MINIMUM),
          BigInteger.valueOf(LINEA_GAS_LIMIT_MAXIMUM),
          LineaL1L2BridgeSharedConfiguration.TEST_DEFAULT);

  /**
   * Represents Linea sepolia prior to the block gas limit being enforced for the purposes of
   * running existing replay tests. As the name suggest, this is only intended for testing purposes.
   */
  public static final ChainConfig OLD_SEPOLIA_TESTCONFIG =
      new ChainConfig(
          LINEA_SEPOLIA_CHAIN_ID,
          false,
          BigInteger.valueOf(LINEA_GAS_LIMIT_MINIMUM),
          BigInteger.valueOf(LINEA_GAS_LIMIT_MAXIMUM),
          LineaL1L2BridgeSharedConfiguration.TEST_DEFAULT);

  /** ChainID for this chain */
  public final BigInteger id;

  /** Determines whether block gas limit is enabled (or not). */
  public final boolean fixedGasLimitEnabled;

  /** Determines minimum gas limit for this chain */
  public final BigInteger gasLimitMinimum;

  /** Determines maximum gas limit for this chain */
  public final BigInteger gasLimitMaximum;

  /** Determines the bridge configuration being used. */
  public final LineaL1L2BridgeSharedConfiguration bridgeConfiguration;

  private ChainConfig(
      int chainId,
      boolean gasLimitEnabled,
      BigInteger gasLimitMinimum,
      BigInteger gasLimitMaximum,
      LineaL1L2BridgeSharedConfiguration bridgeConfig) {
    this(
        BigInteger.valueOf(chainId),
        gasLimitEnabled,
        gasLimitMinimum,
        gasLimitMaximum,
        bridgeConfig);
  }

  private ChainConfig(
      BigInteger chainId,
      boolean fixedGasLimitEnabled,
      BigInteger gasLimitMinimum,
      BigInteger gasLimitMaximum,
      LineaL1L2BridgeSharedConfiguration bridgeConfig) {
    // Sanity cehck chainId is non-negative.
    if (chainId.compareTo(BigInteger.ZERO) < 0) {
      throw new IllegalArgumentException("invalid chain id (" + chainId + ")");
    }
    //
    this.id = chainId;
    this.fixedGasLimitEnabled = fixedGasLimitEnabled;
    this.gasLimitMinimum = gasLimitMinimum;
    this.gasLimitMaximum = gasLimitMaximum;
    this.bridgeConfiguration = bridgeConfig;
  }

  /**
   * Construct a Linea chain configuration with a given chain ID. This constructor is used, for
   * example, to create an appropriate configuration when running in production, depending on
   * whether we are running e.g. on mainnet or sepolia.
   *
   * @param chainId
   * @return
   */
  public static ChainConfig LINEA_CHAIN(
      LineaL1L2BridgeSharedConfiguration bridgeConfig, BigInteger chainId) {
    return new ChainConfig(
        chainId,
        true,
        BigInteger.valueOf(LINEA_GAS_LIMIT_MINIMUM),
        BigInteger.valueOf(LINEA_GAS_LIMIT_MAXIMUM),
        bridgeConfig);
  }
}
