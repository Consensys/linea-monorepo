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

package net.consensys.linea.plugins;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.plugins.config.LineaL1L2BridgeCliOptions;
import net.consensys.linea.plugins.config.LineaL1L2BridgeConfiguration;
import net.consensys.linea.plugins.config.LineaTracerCliOptions;
import net.consensys.linea.plugins.config.LineaTracerConfiguration;
import org.hyperledger.besu.plugin.BesuContext;

/**
 * In this class we put CLI options that are private to plugins in this repo.
 *
 * <p>For the moment is just a placeholder since there are no private options
 */
@Slf4j
public abstract class AbstractLineaPrivateOptionsPlugin extends AbstractLineaSharedOptionsPlugin {
  private static final String CLI_OPTIONS_PREFIX = "linea";
  private static boolean cliOptionsRegistered = false;
  private static boolean configured = false;
  private static LineaTracerCliOptions tracerCliOptions;
  private static LineaL1L2BridgeCliOptions l1L2BridgeCliOptions;
  protected static LineaTracerConfiguration tracerConfiguration;
  protected static LineaL1L2BridgeConfiguration l1L2BridgeConfiguration;

  @Override
  public synchronized void register(final BesuContext context) {
    super.register(context);
    if (!cliOptionsRegistered) {
      //      final PicoCLIOptions cmdlineOptions =
      //          context
      //              .getService(PicoCLIOptions.class)
      //              .orElseThrow(
      //                  () ->
      //                      new IllegalStateException(
      //                          "Failed to obtain PicoCLI options from the BesuContext"));
      cliOptionsRegistered = true;
    }
  }

  @Override
  public void beforeExternalServices() {
    super.beforeExternalServices();
    if (!configured) {
      configured = true;
    }
  }

  @Override
  public void start() {
    super.start();
  }

  @Override
  public void stop() {
    super.stop();
    cliOptionsRegistered = false;
    configured = false;
  }
}
