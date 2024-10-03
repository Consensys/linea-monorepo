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

import java.net.URL;
import java.util.Optional;

import com.google.common.base.MoreObjects;
import net.consensys.linea.plugins.LineaCliOptions;
import picocli.CommandLine.ArgGroup;
import picocli.CommandLine.Option;

/** The Linea Rejected Transaction Reporting CLI options. */
public class LineaRejectedTxReportingCliOptions implements LineaCliOptions {
  /**
   * The configuration key used in AbstractLineaPrivateOptionsPlugin to identify the cli options.
   */
  public static final String CONFIG_KEY = "rejected-tx-reporting-config";

  /** The rejected transaction endpoint. */
  public static final String REJECTED_TX_ENDPOINT = "--plugin-linea-rejected-tx-endpoint";

  /** The Linea node type. */
  public static final String LINEA_NODE_TYPE = "--plugin-linea-node-type";

  @ArgGroup(exclusive = false)
  DependentOptions dependentOptions; // will be null if no options from this group are specified

  static class DependentOptions {
    @Option(
        names = {REJECTED_TX_ENDPOINT},
        hidden = true,
        required = true, // required within the group
        paramLabel = "<URL>",
        description =
            "Endpoint URI for reporting rejected transactions. Specify a valid URI to enable reporting.")
    URL rejectedTxEndpoint = null;

    @Option(
        names = {LINEA_NODE_TYPE},
        hidden = true,
        required = true, // required within the group
        paramLabel = "<NODE_TYPE>",
        description =
            "Linea Node type to use when reporting rejected transactions. (default: ${DEFAULT-VALUE}. Valid values: ${COMPLETION-CANDIDATES})")
    LineaNodeType lineaNodeType = null;
  }

  /** Default constructor. */
  private LineaRejectedTxReportingCliOptions() {}

  /**
   * Create Linea Rejected Transaction Reporting CLI options.
   *
   * @return the Linea Rejected Transaction Reporting CLI options
   */
  public static LineaRejectedTxReportingCliOptions create() {
    return new LineaRejectedTxReportingCliOptions();
  }

  /**
   * Instantiates a new Linea rejected tx reporting cli options from Configuration object
   *
   * @param config An instance of LineaRejectedTxReportingConfiguration
   */
  public static LineaRejectedTxReportingCliOptions fromConfig(
      final LineaRejectedTxReportingConfiguration config) {
    final LineaRejectedTxReportingCliOptions options = create();
    // both options are required.
    if (config.rejectedTxEndpoint() != null && config.lineaNodeType() != null) {
      final var depOpts = new DependentOptions();
      depOpts.rejectedTxEndpoint = config.rejectedTxEndpoint();
      depOpts.lineaNodeType = config.lineaNodeType();
      options.dependentOptions = depOpts;
    }

    return options;
  }

  @Override
  public LineaRejectedTxReportingConfiguration toDomainObject() {
    final var rejectedTxEndpoint =
        Optional.ofNullable(dependentOptions).map(o -> o.rejectedTxEndpoint).orElse(null);
    final var lineaNodeType =
        Optional.ofNullable(dependentOptions).map(o -> o.lineaNodeType).orElse(null);

    return LineaRejectedTxReportingConfiguration.builder()
        .rejectedTxEndpoint(rejectedTxEndpoint)
        .lineaNodeType(lineaNodeType)
        .build();
  }

  @Override
  public String toString() {
    final var rejectedTxEndpoint =
        Optional.ofNullable(dependentOptions).map(o -> o.rejectedTxEndpoint).orElse(null);
    final var lineaNodeType =
        Optional.ofNullable(dependentOptions).map(o -> o.lineaNodeType).orElse(null);

    return MoreObjects.toStringHelper(this)
        .add(REJECTED_TX_ENDPOINT, rejectedTxEndpoint)
        .add(LINEA_NODE_TYPE, lineaNodeType)
        .toString();
  }
}
