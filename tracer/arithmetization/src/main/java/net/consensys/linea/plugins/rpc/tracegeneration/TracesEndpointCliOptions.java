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

package net.consensys.linea.plugins.rpc.tracegeneration;

import com.google.common.base.MoreObjects;
import net.consensys.linea.plugins.LineaCliOptions;
import picocli.CommandLine;

public class TracesEndpointCliOptions implements LineaCliOptions {

  static final String CONFIG_KEY = "traces-endpoint-config";

  static final String CONFLATED_TRACE_GENERATION_TRACES_OUTPUT_PATH =
      "--plugin-linea-conflated-trace-generation-traces-output-path";

  static final String CONFLATED_TRACE_GENERATION_CONCURRENT_REQUESTS_LIMIT =
      "--plugin-linea-conflated-trace-generation-concurrent-requests-limit";

  @CommandLine.Option(
      required = true,
      names = {CONFLATED_TRACE_GENERATION_TRACES_OUTPUT_PATH},
      hidden = true,
      paramLabel = "<PATH>",
      description = "Path to where traces will be written")
  private String tracesOutputPath = null;

  private TracesEndpointCliOptions() {}

  /**
   * Create Linea cli options.
   *
   * @return the Linea cli options
   */
  public static TracesEndpointCliOptions create() {
    return new TracesEndpointCliOptions();
  }

  /**
   * Linea cli options from config.
   *
   * @param config the config
   * @return the Linea cli options
   */
  static TracesEndpointCliOptions fromConfig(final TracesEndpointConfiguration config) {
    final TracesEndpointCliOptions options = create();
    options.tracesOutputPath = config.tracesOutputPath();
    return options;
  }

  /**
   * To domain object Linea factory configuration.
   *
   * @return the Linea factory configuration
   */
  @Override
  public TracesEndpointConfiguration toDomainObject() {
    return TracesEndpointConfiguration.builder().tracesOutputPath(tracesOutputPath).build();
  }

  @Override
  public String toString() {
    return MoreObjects.toStringHelper(this)
        .add(CONFLATED_TRACE_GENERATION_TRACES_OUTPUT_PATH, tracesOutputPath)
        .toString();
  }
}
