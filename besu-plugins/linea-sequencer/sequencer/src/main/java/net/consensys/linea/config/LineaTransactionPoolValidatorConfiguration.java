/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.config;

import lombok.Builder;
import net.consensys.linea.plugins.LineaOptionsConfiguration;

/**
 * The Linea transaction pool validation configuration.
 *
 * @param denyListPath the path to the file containing the addresses that are denied.
 * @param maxTxGasLimit the maximum gas limit allowed for transactions
 * @param maxTxCalldataSize the maximum size of calldata allowed for transactions
 */
@Builder(toBuilder = true)
public record LineaTransactionPoolValidatorConfiguration(
    String denyListPath,
    int maxTxGasLimit,
    int maxTxCalldataSize,
    boolean txPoolSimulationCheckApiEnabled,
    boolean txPoolSimulationCheckP2pEnabled)
    implements LineaOptionsConfiguration {}
