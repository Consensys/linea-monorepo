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

package net.consensys.linea.sequencer.txselection;

import lombok.Builder;
import lombok.Getter;

/** The Linea configuration. */
@Getter
@Builder
public final class LineaTransactionSelectorConfiguration {
  private final int maxBlockCallDataSize;
  private final String moduleLimitsFilePath;
  private final long maxGasPerBlock;
  private final int verificationGasCost;
  private final int verificationCapacity;
  private final int gasPriceRatio;
  private final double minMargin;
  private final int adjustTxSize;
  private final int unprofitableCacheSize;
  private final int unprofitableRetryLimit;
}
