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
package net.consensys.linea.continoustracing;

import com.google.common.base.MoreObjects;
import picocli.CommandLine;

public class ContinuousTracingCliOptions {
  public static final String CONTINUOUS_TRACING_ENABLED =
      "--plugin-linea-continuous-tracing-enabled";
  public static final String CONTINUOUS_TRACING_ZK_EVM_BIN =
      "--plugin-linea-continuous-tracing-zk-evm-bin";

  @CommandLine.Option(
      names = {CONTINUOUS_TRACING_ENABLED},
      hidden = true,
      paramLabel = "<BOOLEAN>",
      description = "Enable continuous tracing (default: false)")
  private boolean continuousTracingEnabled = false;

  @CommandLine.Option(
      names = {CONTINUOUS_TRACING_ZK_EVM_BIN},
      hidden = true,
      paramLabel = "<PATH>",
      description = "Path to the ZkEvm binary")
  private String zkEvmBin = null;

  private ContinuousTracingCliOptions() {}

  public static ContinuousTracingCliOptions create() {
    return new ContinuousTracingCliOptions();
  }

  public ContinuousTracingConfiguration toDomainObject() {
    return new ContinuousTracingConfiguration(continuousTracingEnabled, zkEvmBin);
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(CONTINUOUS_TRACING_ENABLED, continuousTracingEnabled)
        .add(CONTINUOUS_TRACING_ZK_EVM_BIN, zkEvmBin)
        .toString();
  }
}
