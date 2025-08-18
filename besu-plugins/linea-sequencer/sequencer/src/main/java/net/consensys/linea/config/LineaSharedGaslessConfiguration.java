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
package net.consensys.linea.config;

import net.consensys.linea.plugins.LineaOptionsConfiguration;

/**
 * Shared configuration parameters for gasless transaction features (RLN, RPC modifications).
 *
 * @param denyListPath Path to the text file storing addresses of users on the deny list. This file
 *     is read by the RPC estimateGas method and read/written by the RLN Validator.
 * @param denyListRefreshSeconds Interval in seconds at which the deny list file should be reloaded
 *     by components.
 * @param premiumGasPriceThresholdGWei Minimum gas price (in GWei) for a transaction to be
 *     considered premium.
 * @param denyListEntryMaxAgeMinutes Maximum age in minutes for an entry on the deny list before it
 *     expires.
 */
public record LineaSharedGaslessConfiguration(
    String denyListPath,
    long denyListRefreshSeconds,
    long premiumGasPriceThresholdGWei,
    long denyListEntryMaxAgeMinutes)
    implements LineaOptionsConfiguration {

  public static final String DEFAULT_DENY_LIST_PATH = "/var/lib/besu/gasless-deny-list.txt";
  public static final long DEFAULT_DENY_LIST_REFRESH_SECONDS = 300L; // 5 minutes
  public static final long DEFAULT_PREMIUM_GAS_PRICE_THRESHOLD_GWEI = 100L; // 100 Gwei
  public static final long DEFAULT_DENY_LIST_ENTRY_MAX_AGE_MINUTES = 10L; // 10 minutes

  public static LineaSharedGaslessConfiguration V1_DEFAULT =
      new LineaSharedGaslessConfiguration(
          DEFAULT_DENY_LIST_PATH,
          DEFAULT_DENY_LIST_REFRESH_SECONDS,
          DEFAULT_PREMIUM_GAS_PRICE_THRESHOLD_GWEI,
          DEFAULT_DENY_LIST_ENTRY_MAX_AGE_MINUTES);

  // Constructor allowing easy overriding of the path if needed from other config sources
  public LineaSharedGaslessConfiguration {
    if (denyListPath == null || denyListPath.isBlank()) {
      throw new IllegalArgumentException("Deny list path cannot be null or blank.");
    }
    if (denyListRefreshSeconds <= 0) {
      throw new IllegalArgumentException("Deny list refresh seconds must be positive.");
    }
    if (premiumGasPriceThresholdGWei <= 0) {
      throw new IllegalArgumentException("Premium gas price threshold GWei must be positive.");
    }
    if (denyListEntryMaxAgeMinutes <= 0) {
      throw new IllegalArgumentException("Deny list entry max age minutes must be positive.");
    }
  }
}
