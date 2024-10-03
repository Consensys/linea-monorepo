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

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatExceptionOfType;

import java.net.MalformedURLException;
import java.net.URI;

import org.hyperledger.besu.plugin.services.PicoCLIOptions;
import org.hyperledger.besu.services.PicoCLIOptionsImpl;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.EnumSource;
import org.junit.jupiter.params.provider.ValueSource;
import picocli.CommandLine;
import picocli.CommandLine.Command;
import picocli.CommandLine.Option;

class LineaRejectedTxReportingCliOptionsTest {

  @Command
  static final class MockLineaBesuCommand {
    @Option(names = "--mock-option")
    String mockOption;
  }

  private MockLineaBesuCommand command;
  private LineaRejectedTxReportingCliOptions lineaRejectedTxReportingCliOptions;
  private CommandLine commandLine;
  private PicoCLIOptions picoCliService;

  @BeforeEach
  public void setup() {
    command = new MockLineaBesuCommand();
    commandLine = new CommandLine(command);
    picoCliService = new PicoCLIOptionsImpl(commandLine);

    lineaRejectedTxReportingCliOptions = LineaRejectedTxReportingCliOptions.create();
    picoCliService.addPicoCLIOptions("linea", lineaRejectedTxReportingCliOptions);
  }

  @Test
  void emptyLineaRejectedTxReportingCliOptions() {
    commandLine.parseArgs("--mock-option", "mockValue");

    assertThat(command.mockOption).isEqualTo("mockValue");
    assertThat(lineaRejectedTxReportingCliOptions.dependentOptions).isNull();
  }

  @ParameterizedTest
  @EnumSource(LineaNodeType.class)
  void lineaRejectedTxOptionBothOptionsRequired(final LineaNodeType lineaNodeType)
      throws MalformedURLException {
    commandLine.parseArgs(
        "--plugin-linea-rejected-tx-endpoint",
        "http://localhost:8080",
        "--plugin-linea-node-type",
        lineaNodeType.name());

    assertThat(lineaRejectedTxReportingCliOptions.dependentOptions.rejectedTxEndpoint)
        .isEqualTo(URI.create("http://localhost:8080").toURL());
    assertThat(lineaRejectedTxReportingCliOptions.dependentOptions.lineaNodeType)
        .isEqualTo(lineaNodeType);
  }

  @Test
  void lineaRejectedTxReportingCliOptionsOnlyEndpointCauseException() {
    assertThatExceptionOfType(CommandLine.ParameterException.class)
        .isThrownBy(
            () ->
                commandLine.parseArgs(
                    "--plugin-linea-rejected-tx-endpoint", "http://localhost:8080"))
        .withMessageContaining(
            "Error: Missing required argument(s): --plugin-linea-node-type=<NODE_TYPE>");
  }

  @Test
  void lineaRejectedTxReportingCliOptionsOnlyNodeTypeCauseException() {
    assertThatExceptionOfType(CommandLine.ParameterException.class)
        .isThrownBy(
            () -> commandLine.parseArgs("--plugin-linea-node-type", LineaNodeType.SEQUENCER.name()))
        .withMessageContaining(
            "Error: Missing required argument(s): --plugin-linea-rejected-tx-endpoint=<URL>");
  }

  @Test
  void lineaRejectedTxReportingInvalidNodeTypeCauseException() {
    assertThatExceptionOfType(CommandLine.ParameterException.class)
        .isThrownBy(
            () ->
                commandLine.parseArgs(
                    "--plugin-linea-rejected-tx-endpoint",
                    "http://localhost:8080",
                    "--plugin-linea-node-type",
                    "INVALID_NODE_TYPE"))
        .withMessageContaining(
            "Invalid value for option '--plugin-linea-node-type': expected one of [SEQUENCER, RPC, P2P] (case-sensitive) but was 'INVALID_NODE_TYPE'");
  }

  @ParameterizedTest
  @ValueSource(strings = {"", "http://localhost:8080:8080", "invalid"})
  void lineaRejectedTxReportingCliOptionsInvalidEndpointCauseException(final String endpoint) {
    assertThatExceptionOfType(CommandLine.ParameterException.class)
        .isThrownBy(
            () ->
                commandLine.parseArgs(
                    "--plugin-linea-rejected-tx-endpoint",
                    endpoint,
                    "--plugin-linea-node-type",
                    "SEQUENCER"))
        .withMessageContaining(
            "Invalid value for option '--plugin-linea-rejected-tx-endpoint': cannot convert '"
                + endpoint
                + "' to URL (java.net.MalformedURLException:");
  }
}
