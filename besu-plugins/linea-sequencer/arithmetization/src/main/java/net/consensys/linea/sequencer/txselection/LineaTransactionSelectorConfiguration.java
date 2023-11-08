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

  private LineaTransactionSelectorConfiguration(
      int maxBlockCallDataSize, final String moduleLimitsFilePath) {
    this.maxBlockCallDataSize = maxBlockCallDataSize;
    this.moduleLimitsFilePath = moduleLimitsFilePath;
  }

  public int maxBlockCallDataSize() {
    return maxBlockCallDataSize;
  }

  public String moduleLimitsFilePath() {
    return moduleLimitsFilePath;
  }

  public static class Builder {
    private int maxBlockCallDataSize;
    private String moduleLimitsFilePath;

    public Builder maxBlockCallDataSize(final int maxBlockCallDataSize) {
      this.maxBlockCallDataSize = maxBlockCallDataSize;
      return this;
    }

    public Builder moduleLimits(final String moduleLimitFilePath) {
      this.moduleLimitsFilePath = moduleLimitFilePath;
      return this;
    }

    public LineaTransactionSelectorConfiguration build() {
      return new LineaTransactionSelectorConfiguration(maxBlockCallDataSize, moduleLimitsFilePath);
    }
  }
}
