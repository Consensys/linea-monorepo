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

package net.consensys.linea;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.config.LineaTracerCliOptions;
import net.consensys.linea.config.LineaTracerConfiguration;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;

@Slf4j
public abstract class AbstractLineaSharedOptionsPlugin implements BesuPlugin {
  private static final String CLI_OPTIONS_PREFIX = "linea";
  private static boolean cliOptionsRegistered = false;
  private static boolean configured = false;
  protected static LineaTracerCliOptions tracerCliOptions;
  protected static LineaTracerConfiguration tracerConfiguration;

  @Override
  public synchronized void register(final BesuContext context) {
    if (!cliOptionsRegistered) {
      final PicoCLIOptions cmdlineOptions =
          context
              .getService(PicoCLIOptions.class)
              .orElseThrow(
                  () ->
                      new IllegalStateException(
                          "Failed to obtain PicoCLI options from the BesuContext"));
      tracerCliOptions = LineaTracerCliOptions.create();

      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, tracerCliOptions);
      cliOptionsRegistered = true;
    }
  }

  @Override
  public void beforeExternalServices() {
    if (!configured) {
      tracerConfiguration = tracerCliOptions.toDomainObject();
      configured = true;
    }

    log.debug("Configured plugin {} with tracer configuration: {}", getName(), tracerConfiguration);
  }

  @Override
  public void start() {}

  @Override
  public void stop() {
    cliOptionsRegistered = false;
    configured = false;
  }
}
