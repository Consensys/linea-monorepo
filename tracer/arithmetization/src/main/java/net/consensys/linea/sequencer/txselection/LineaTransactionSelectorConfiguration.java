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

/** The Linea configuration. */
public final class LineaTransactionSelectorConfiguration {
  private final int maxBlockCallDataSize;
  private final String moduleLimitsFilePath;
  private final long maxGasPerBlock;
  private final int verificationGasCost;
  private final int verificationCapacity;
  private final int gasPriceRatio;
  private final double minMargin;

  private LineaTransactionSelectorConfiguration(
      final int maxBlockCallDataSize,
      final String moduleLimitsFilePath,
      final long maxGasPerBlock,
      final int verificationGasCost,
      final int verificationCapacity,
      final int gasPriceRatio,
      final double minMargin) {
    this.maxBlockCallDataSize = maxBlockCallDataSize;
    this.moduleLimitsFilePath = moduleLimitsFilePath;
    this.maxGasPerBlock = maxGasPerBlock;
    this.verificationGasCost = verificationGasCost;
    this.verificationCapacity = verificationCapacity;
    this.gasPriceRatio = gasPriceRatio;
    this.minMargin = minMargin;
  }

  public int maxBlockCallDataSize() {
    return maxBlockCallDataSize;
  }

  public String moduleLimitsFilePath() {
    return moduleLimitsFilePath;
  }

  public long maxGasPerBlock() {
    return maxGasPerBlock;
  }

  public int getVerificationGasCost() {
    return verificationGasCost;
  }

  public int getVerificationCapacity() {
    return verificationCapacity;
  }

  public int getGasPriceRatio() {
    return gasPriceRatio;
  }

  public double getMinMargin() {
    return minMargin;
  }

  public static class Builder {
    private int maxBlockCallDataSize;
    private String moduleLimitsFilePath;
    private long maxGasPerBlock;
    private int verificationGasCost;
    private int verificationCapacity;
    private int gasPriceRatio;
    private double minMargin;

    public Builder maxBlockCallDataSize(final int maxBlockCallDataSize) {
      this.maxBlockCallDataSize = maxBlockCallDataSize;
      return this;
    }

    public Builder moduleLimits(final String moduleLimitFilePath) {
      this.moduleLimitsFilePath = moduleLimitFilePath;
      return this;
    }

    public Builder maxGasPerBlock(final long maxGasPerBlock) {
      this.maxGasPerBlock = maxGasPerBlock;
      return this;
    }

    public Builder verificationGasCost(final int verificationGasCost) {
      this.verificationGasCost = verificationGasCost;
      return this;
    }

    public Builder verificationCapacity(final int verificationCapacity) {
      this.verificationCapacity = verificationCapacity;
      return this;
    }

    public Builder gasPriceRatio(final int gasPriceRatio) {
      this.gasPriceRatio = gasPriceRatio;
      return this;
    }

    public Builder minMargin(final double minMargin) {
      this.minMargin = minMargin;
      return this;
    }

    public LineaTransactionSelectorConfiguration build() {
      return new LineaTransactionSelectorConfiguration(
          maxBlockCallDataSize,
          moduleLimitsFilePath,
          maxGasPerBlock,
          verificationGasCost,
          verificationCapacity,
          gasPriceRatio,
          minMargin);
    }
  }
}
