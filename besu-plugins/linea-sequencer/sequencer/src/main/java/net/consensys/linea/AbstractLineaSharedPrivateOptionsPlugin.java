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

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.atomic.AtomicBoolean;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.bundles.BundlePoolService;
import net.consensys.linea.bundles.LineaLimitedBundlePool;
import net.consensys.linea.config.LineaBundleCliOptions;
import net.consensys.linea.config.LineaBundleConfiguration;
import net.consensys.linea.config.LineaProfitabilityCliOptions;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import net.consensys.linea.config.LineaRejectedTxReportingCliOptions;
import net.consensys.linea.config.LineaRejectedTxReportingConfiguration;
import net.consensys.linea.config.LineaRpcCliOptions;
import net.consensys.linea.config.LineaRpcConfiguration;
import net.consensys.linea.config.LineaTracerCliOptions;
import net.consensys.linea.config.LineaTracerConfiguration;
import net.consensys.linea.config.LineaTransactionPoolValidatorCliOptions;
import net.consensys.linea.config.LineaTransactionPoolValidatorConfiguration;
import net.consensys.linea.config.LineaTransactionSelectorCliOptions;
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import net.consensys.linea.plugins.AbstractLineaSharedOptionsPlugin;
import net.consensys.linea.plugins.LineaOptionsPluginConfiguration;
import net.consensys.linea.utils.Compressor;
import org.hyperledger.besu.plugin.ServiceManager;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.BesuEvents;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.MetricsSystem;
import org.hyperledger.besu.plugin.services.RpcEndpointService;
import org.hyperledger.besu.plugin.services.metrics.MetricCategoryRegistry;

/**
 * This abstract class is used as superclass for all the plugins that share one or more
 * configuration options, services and common initializations.
 *
 * <p>Configuration options that are exclusive of a single plugin, are not required to be added
 * here, but they could stay in the class that implement a plugin, but in case that configuration
 * becomes to be used by multiple plugins, then to avoid code duplications and possible different
 * management of the options, it is better to move the configuration here so all plugins will
 * automatically see it.
 *
 * <p>Same for services and other initialization tasks, that are shared by more than one plugin,
 * like registration of metrics categories or check to perform once at startup
 */
@Slf4j
public abstract class AbstractLineaSharedPrivateOptionsPlugin
    extends AbstractLineaSharedOptionsPlugin {
  protected static BesuConfiguration besuConfiguration;
  protected static BlockchainService blockchainService;
  protected static MetricsSystem metricsSystem;
  protected static BesuEvents besuEvents;
  protected static BundlePoolService bundlePoolService;
  protected static MetricCategoryRegistry metricCategoryRegistry;
  protected static RpcEndpointService rpcEndpointService;

  private static final AtomicBoolean sharedRegisterTasksDone = new AtomicBoolean(false);
  private static final AtomicBoolean sharedStartTasksDone = new AtomicBoolean(false);

  static {
    // force the initialization of the gnark compress native library to fail fast in case of issues
    Compressor.instance.compressedSize(new byte[1024]);
  }

  private ServiceManager serviceManager;

  @Override
  public Map<String, LineaOptionsPluginConfiguration> getLineaPluginConfigMap() {
    final var configMap = new HashMap<>(super.getLineaPluginConfigMap());

    configMap.put(
        LineaTransactionSelectorCliOptions.CONFIG_KEY,
        LineaTransactionSelectorCliOptions.create().asPluginConfig());
    configMap.put(
        LineaTransactionPoolValidatorCliOptions.CONFIG_KEY,
        LineaTransactionPoolValidatorCliOptions.create().asPluginConfig());
    configMap.put(LineaRpcCliOptions.CONFIG_KEY, LineaRpcCliOptions.create().asPluginConfig());
    configMap.put(
        LineaProfitabilityCliOptions.CONFIG_KEY,
        LineaProfitabilityCliOptions.create().asPluginConfig());
    configMap.put(
        LineaTracerCliOptions.CONFIG_KEY, LineaTracerCliOptions.create().asPluginConfig());
    configMap.put(
        LineaRejectedTxReportingCliOptions.CONFIG_KEY,
        LineaRejectedTxReportingCliOptions.create().asPluginConfig());
    configMap.put(
        LineaBundleCliOptions.CONFIG_KEY, LineaBundleCliOptions.create().asPluginConfig());
    return configMap;
  }

  public LineaTransactionSelectorConfiguration transactionSelectorConfiguration() {
    return (LineaTransactionSelectorConfiguration)
        getConfigurationByKey(LineaTransactionSelectorCliOptions.CONFIG_KEY).optionsConfig();
  }

  public LineaTransactionPoolValidatorConfiguration transactionPoolValidatorConfiguration() {
    return (LineaTransactionPoolValidatorConfiguration)
        getConfigurationByKey(LineaTransactionPoolValidatorCliOptions.CONFIG_KEY).optionsConfig();
  }

  public LineaRpcConfiguration lineaRpcConfiguration() {
    return (LineaRpcConfiguration)
        getConfigurationByKey(LineaRpcCliOptions.CONFIG_KEY).optionsConfig();
  }

  public LineaProfitabilityConfiguration profitabilityConfiguration() {
    return (LineaProfitabilityConfiguration)
        getConfigurationByKey(LineaProfitabilityCliOptions.CONFIG_KEY).optionsConfig();
  }

  public LineaTracerConfiguration tracerConfiguration() {
    return (LineaTracerConfiguration)
        getConfigurationByKey(LineaTracerCliOptions.CONFIG_KEY).optionsConfig();
  }

  public LineaRejectedTxReportingConfiguration rejectedTxReportingConfiguration() {
    return (LineaRejectedTxReportingConfiguration)
        getConfigurationByKey(LineaRejectedTxReportingCliOptions.CONFIG_KEY).optionsConfig();
  }

  public LineaBundleConfiguration bundleConfiguration() {
    return (LineaBundleConfiguration)
        getConfigurationByKey(LineaBundleCliOptions.CONFIG_KEY).optionsConfig();
  }

  @Override
  public synchronized void register(final ServiceManager serviceManager) {
    super.register(serviceManager);

    this.serviceManager = serviceManager;

    if (sharedRegisterTasksDone.compareAndSet(false, true)) {
      performSharedRegisterTasksOnce(serviceManager);
    }
  }

  protected static void performSharedRegisterTasksOnce(final ServiceManager serviceManager) {
    besuConfiguration =
        serviceManager
            .getService(BesuConfiguration.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain BesuConfiguration from the ServiceManager."));
    blockchainService =
        serviceManager
            .getService(BlockchainService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain BlockchainService from the ServiceManager."));

    metricCategoryRegistry =
        serviceManager
            .getService(MetricCategoryRegistry.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain MetricCategoryRegistry from the ServiceManager."));

    rpcEndpointService =
        serviceManager
            .getService(RpcEndpointService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain RpcEndpointService from the ServiceManager."));
  }

  @Override
  public void start() {
    super.start();

    if (sharedStartTasksDone.compareAndSet(false, true)) {
      performSharedStartTasksOnce(serviceManager);
    }
  }

  private void performSharedStartTasksOnce(final ServiceManager serviceManager) {

    blockchainService
        .getChainId()
        .ifPresentOrElse(
            chainId -> {
              if (chainId.signum() <= 0) {
                throw new IllegalArgumentException("Chain id must be greater than zero.");
              }
            },
            () -> {
              throw new IllegalArgumentException("Chain id required");
            });

    metricsSystem =
        serviceManager
            .getService(MetricsSystem.class)
            .orElseThrow(
                () ->
                    new RuntimeException("Failed to obtain MetricSystem from the ServiceManager."));

    besuEvents =
        serviceManager
            .getService(BesuEvents.class)
            .orElseThrow(
                () -> new RuntimeException("Failed to obtain BesuEvents from the ServiceManager."));

    bundlePoolService =
        new LineaLimitedBundlePool(
            besuConfiguration.getDataPath(),
            transactionSelectorConfiguration().maxBundlePoolSizeBytes(),
            besuEvents,
            blockchainService);
    bundlePoolService.loadFromDisk();
  }

  @Override
  public void stop() {
    super.stop();
    sharedRegisterTasksDone.set(false);
    sharedStartTasksDone.set(false);
    blockchainService = null;
    metricsSystem = null;
  }
}
