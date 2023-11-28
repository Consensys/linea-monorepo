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

import java.util.Optional;

import com.google.auto.service.AutoService;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.BesuEvents;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;
import org.hyperledger.besu.plugin.services.TraceService;

@Slf4j
@AutoService(BesuPlugin.class)
public class ContinuousTracingPlugin implements BesuPlugin {
  public static final String NAME = "linea-continuous";
  public static final String ENV_WEBHOOK_URL = "SLACK_SHADOW_NODE_WEBHOOK_URL";

  private final ContinuousTracingCliOptions options;
  private BesuContext context;

  public ContinuousTracingPlugin() {
    options = ContinuousTracingCliOptions.create();
  }

  @Override
  public Optional<String> getName() {
    return Optional.of(NAME);
  }

  @Override
  public void register(final BesuContext context) {
    final PicoCLIOptions cmdlineOptions =
        context
            .getService(PicoCLIOptions.class)
            .orElseThrow(
                () ->
                    new IllegalStateException(
                        "Expecting a PicoCLI options to register CLI options with, but none found."));

    cmdlineOptions.addPicoCLIOptions(getName().get(), options);

    this.context = context;
  }

  @Override
  public void start() {
    log.info("Starting {} with configuration: {}", NAME, options);

    // BesuEvents can only be requested after the plugin has been registered.
    final BesuEvents besuEvents =
        context
            .getService(BesuEvents.class)
            .orElseThrow(
                () ->
                    new IllegalStateException(
                        "Expecting a BesuEvents to register events with, but none found."));

    final TraceService traceService =
        context
            .getService(TraceService.class)
            .orElseThrow(
                () -> new IllegalStateException("Expecting a TraceService, but none found."));

    final ContinuousTracingConfiguration tracingConfiguration = options.toDomainObject();

    if (tracingConfiguration.continuousTracing() && (tracingConfiguration.zkEvmBin() == null)) {
      log.error("zkEvmBin must be specified when continuousTracing is enabled");
      System.exit(1);
    }

    final String webHookUrl = System.getenv(ENV_WEBHOOK_URL);
    if (tracingConfiguration.continuousTracing() && (webHookUrl == null)) {
      log.error(
          "Webhook URL must be specified as environment variable {} when continuousTracing is enabled",
          ENV_WEBHOOK_URL);
      System.exit(1);
    }

    if (tracingConfiguration.continuousTracing()) {
      OpCodes.load(); // must be loaded explicitly

      besuEvents.addBlockAddedListener(
          new ContinuousTracingBlockAddedListener(
              new ContinuousTracer(traceService, new CorsetValidator()),
              new TraceFailureHandler(SlackNotificationService.create(webHookUrl)),
              tracingConfiguration.zkEvmBin()));
    }
  }

  @Override
  public void stop() {}
}
