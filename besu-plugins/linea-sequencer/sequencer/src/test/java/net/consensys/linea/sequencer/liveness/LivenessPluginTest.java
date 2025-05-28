/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.liveness;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.math.BigInteger;
import java.nio.file.Path;
import java.time.Instant;
import java.util.Optional;
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import net.consensys.linea.config.LivenessPluginConfiguration;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.plugin.ServiceManager;
import org.hyperledger.besu.plugin.data.AddedBlockContext;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.services.*;
import org.hyperledger.besu.plugin.services.metrics.Counter;
import org.hyperledger.besu.plugin.services.metrics.LabelledSuppliedMetric;
import org.hyperledger.besu.plugin.services.metrics.MetricCategoryRegistry;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.api.io.TempDir;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.mockito.junit.jupiter.MockitoSettings;
import org.mockito.quality.Strictness;

@ExtendWith(MockitoExtension.class)
@MockitoSettings(strictness = Strictness.LENIENT)
public class LivenessPluginTest {

  @TempDir private Path tempDir;

  @Mock private ServiceManager serviceManager;
  @Mock private BesuEvents besuEvents;
  @Mock private BesuConfiguration besuConfiguration;
  @Mock private BlockchainService blockchainService;
  @Mock private MetricCategoryRegistry metricCategoryRegistry;
  @Mock private MetricsSystem metricsSystem;
  @Mock private RpcEndpointService rpcEndpointService;
  @Mock private PicoCLIOptions picoCLIOptions;
  @Mock private Counter counter;
  @Mock private LabelledSuppliedMetric labelledSuppliedMetric;
  @Mock private BlockHeader blockHeader;
  @Mock private AddedBlockContext addedBlockContext;
  @Mock private LivenessPluginConfiguration livenessPluginConfiguration;
  @Mock private LineaTransactionSelectorConfiguration transactionSelectorConfiguration;

  private LivenessPlugin plugin;

  private static final String CONTRACT_ADDRESS = "0x1234567890123456789012345678901234567890";
  private static final String SIGNER_URL = "http://localhost:9000";
  private static final String SIGNER_KEY_ID = "test-key";
  private static final String SIGNER_ADDRESS = "0x1234567890123456789012345678901234567890";
  private static final long CHECK_INTERVAL_MS = 5000;
  private static final long MAX_BLOCK_AGE_MS = 10000;
  private static final long GAS_LIMIT = 100000;
  private static final long GAS_PRICE_GWEI = 0; // 0 = dynamic
  private static final int MAX_RETRY_ATTEMPTS = 3;
  private static final long RETRY_DELAY_MS = 1000;

  @BeforeEach
  public void setUp() {
    plugin =
        new LivenessPlugin() {
          @Override
          public LivenessPluginConfiguration livenessPluginConfiguration() {
            return livenessPluginConfiguration;
          }

          @Override
          public LineaTransactionSelectorConfiguration transactionSelectorConfiguration() {
            return transactionSelectorConfiguration;
          }
        };

    // Mock all required services
    when(serviceManager.getService(BesuEvents.class)).thenReturn(Optional.of(besuEvents));
    when(serviceManager.getService(BesuConfiguration.class))
        .thenReturn(Optional.of(besuConfiguration));
    when(serviceManager.getService(BlockchainService.class))
        .thenReturn(Optional.of(blockchainService));
    when(serviceManager.getService(MetricCategoryRegistry.class))
        .thenReturn(Optional.of(metricCategoryRegistry));
    when(serviceManager.getService(MetricsSystem.class)).thenReturn(Optional.of(metricsSystem));
    when(serviceManager.getService(RpcEndpointService.class))
        .thenReturn(Optional.of(rpcEndpointService));
    when(serviceManager.getService(PicoCLIOptions.class)).thenReturn(Optional.of(picoCLIOptions));

    // Mock besu configuration
    when(besuConfiguration.getDataPath()).thenReturn(tempDir);

    // Mock blockchain service
    when(blockchainService.getChainId()).thenReturn(Optional.of(BigInteger.valueOf(1337)));
    when(blockchainService.getChainHeadHeader()).thenReturn(blockHeader);

    // Mock block header
    when(blockHeader.getNumber()).thenReturn(100L);
    when(blockHeader.getTimestamp()).thenReturn(Instant.now().getEpochSecond());

    // Mock added block context
    when(addedBlockContext.getBlockHeader()).thenReturn(blockHeader);

    // Mock transaction selector configuration
    when(transactionSelectorConfiguration.maxBundlePoolSizeBytes()).thenReturn(1024L * 1024L);

    // Mock metrics
    when(metricsSystem.createCounter(any(), anyString(), anyString())).thenReturn(counter);
    when(metricsSystem.createLabelledSuppliedGauge(any(), anyString(), anyString(), anyString()))
        .thenReturn(labelledSuppliedMetric);
  }

  @Test
  public void shouldRegisterSuccessfully() {
    assertThatNoException().isThrownBy(() -> plugin.register(serviceManager));
  }

  private void setupDefaultConfiguration() {
    when(livenessPluginConfiguration.enabled()).thenReturn(true);
    when(livenessPluginConfiguration.metricCategoryEnabled()).thenReturn(true);
    when(livenessPluginConfiguration.contractAddress()).thenReturn(CONTRACT_ADDRESS);
    when(livenessPluginConfiguration.signerUrl()).thenReturn(SIGNER_URL);
    when(livenessPluginConfiguration.signerKeyId()).thenReturn(SIGNER_KEY_ID);
    when(livenessPluginConfiguration.signerAddress()).thenReturn(SIGNER_ADDRESS);
    when(livenessPluginConfiguration.checkIntervalMilliseconds()).thenReturn(CHECK_INTERVAL_MS);
    when(livenessPluginConfiguration.maxBlockAgeMilliseconds()).thenReturn(MAX_BLOCK_AGE_MS);
    when(livenessPluginConfiguration.gasLimit()).thenReturn(GAS_LIMIT);
    when(livenessPluginConfiguration.gasPriceGwei()).thenReturn(GAS_PRICE_GWEI);
    when(livenessPluginConfiguration.maxRetryAttempts()).thenReturn(MAX_RETRY_ATTEMPTS);
    when(livenessPluginConfiguration.retryDelayMilliseconds()).thenReturn(RETRY_DELAY_MS);
  }

  @Test
  public void shouldThrowExceptionWhenMetricCategoryRegistryNotAvailable() {
    when(serviceManager.getService(MetricCategoryRegistry.class)).thenReturn(Optional.empty());

    assertThatThrownBy(() -> plugin.register(serviceManager))
        .isInstanceOf(RuntimeException.class)
        .hasMessageContaining("Failed to obtain MetricCategoryRegistry from the ServiceManager");
  }

  @Test
  public void shouldStartSuccessfullyWhenEnabled() {
    setupDefaultConfiguration();
    when(metricCategoryRegistry.isMetricCategoryEnabled(any())).thenReturn(true);

    plugin.register(serviceManager);

    assertThatNoException().isThrownBy(() -> plugin.start());
  }

  @Test
  public void shouldStartSuccessfullyWhenDisabled() {
    when(livenessPluginConfiguration.enabled()).thenReturn(false);
    plugin.register(serviceManager);

    assertThatNoException().isThrownBy(() -> plugin.start());
  }

  @Test
  public void shouldHandleBlockAddedWhenEnabled() {
    setupDefaultConfiguration();
    when(metricCategoryRegistry.isMetricCategoryEnabled(any())).thenReturn(true);

    plugin.register(serviceManager);
    plugin.start();

    long newBlockTimestamp = Instant.now().getEpochSecond();
    when(blockHeader.getTimestamp()).thenReturn(newBlockTimestamp);

    assertThatNoException().isThrownBy(() -> plugin.onBlockAdded(addedBlockContext));
  }

  @Test
  public void shouldHandleBlockAddedWhenDisabled() {
    when(livenessPluginConfiguration.enabled()).thenReturn(false);
    plugin.register(serviceManager);
    plugin.start();

    assertThatNoException().isThrownBy(() -> plugin.onBlockAdded(addedBlockContext));
  }

  @Test
  public void shouldCreateValidFunctionCallData() throws Exception {
    when(livenessPluginConfiguration.enabled()).thenReturn(true);
    when(livenessPluginConfiguration.metricCategoryEnabled()).thenReturn(true);
    when(livenessPluginConfiguration.contractAddress()).thenReturn(CONTRACT_ADDRESS);
    when(livenessPluginConfiguration.signerUrl()).thenReturn(SIGNER_URL);
    when(livenessPluginConfiguration.signerKeyId()).thenReturn(SIGNER_KEY_ID);
    when(livenessPluginConfiguration.signerAddress()).thenReturn(SIGNER_ADDRESS);
    when(livenessPluginConfiguration.checkIntervalMilliseconds()).thenReturn(CHECK_INTERVAL_MS);
    when(livenessPluginConfiguration.maxBlockAgeMilliseconds()).thenReturn(MAX_BLOCK_AGE_MS);
    when(livenessPluginConfiguration.gasLimit()).thenReturn(GAS_LIMIT);
    when(livenessPluginConfiguration.gasPriceGwei()).thenReturn(GAS_PRICE_GWEI);
    when(metricCategoryRegistry.isMetricCategoryEnabled(any())).thenReturn(true);

    plugin.register(serviceManager);
    plugin.start();

    var method =
        LivenessPlugin.class.getDeclaredMethod("createFunctionCallData", boolean.class, long.class);
    method.setAccessible(true);

    Bytes result = (Bytes) method.invoke(plugin, true, 1234567890L);

    assertThat(result).isNotNull();
    assertThat(result.size()).isGreaterThan(4); // At least function selector plus some data
    assertThat(result.toHexString()).startsWith("0x");
  }

  @Test
  public void shouldCreateDifferentFunctionCallDataForUpAndDown() throws Exception {
    when(livenessPluginConfiguration.enabled()).thenReturn(true);
    when(livenessPluginConfiguration.metricCategoryEnabled()).thenReturn(true);
    when(livenessPluginConfiguration.contractAddress()).thenReturn(CONTRACT_ADDRESS);
    when(livenessPluginConfiguration.signerUrl()).thenReturn(SIGNER_URL);
    when(livenessPluginConfiguration.signerKeyId()).thenReturn(SIGNER_KEY_ID);
    when(livenessPluginConfiguration.signerAddress()).thenReturn(SIGNER_ADDRESS);
    when(livenessPluginConfiguration.checkIntervalMilliseconds()).thenReturn(CHECK_INTERVAL_MS);
    when(livenessPluginConfiguration.maxBlockAgeMilliseconds()).thenReturn(MAX_BLOCK_AGE_MS);
    when(livenessPluginConfiguration.gasLimit()).thenReturn(GAS_LIMIT);
    when(livenessPluginConfiguration.gasPriceGwei()).thenReturn(GAS_PRICE_GWEI);
    when(metricCategoryRegistry.isMetricCategoryEnabled(any())).thenReturn(true);

    plugin.register(serviceManager);
    plugin.start();

    var method =
        LivenessPlugin.class.getDeclaredMethod("createFunctionCallData", boolean.class, long.class);
    method.setAccessible(true);

    long timestamp = 1234567890L;

    Bytes upData = (Bytes) method.invoke(plugin, true, timestamp);
    Bytes downData = (Bytes) method.invoke(plugin, false, timestamp);

    assertThat(upData).isNotEqualTo(downData);
    assertThat(upData).isNotNull();
    assertThat(downData).isNotNull();
  }

  @Test
  public void shouldStopGracefully() {
    when(livenessPluginConfiguration.enabled()).thenReturn(true);
    when(livenessPluginConfiguration.metricCategoryEnabled()).thenReturn(true);
    when(livenessPluginConfiguration.contractAddress()).thenReturn(CONTRACT_ADDRESS);
    when(livenessPluginConfiguration.signerUrl()).thenReturn(SIGNER_URL);
    when(livenessPluginConfiguration.signerKeyId()).thenReturn(SIGNER_KEY_ID);
    when(livenessPluginConfiguration.signerAddress()).thenReturn(SIGNER_ADDRESS);
    when(livenessPluginConfiguration.checkIntervalMilliseconds()).thenReturn(CHECK_INTERVAL_MS);
    when(livenessPluginConfiguration.maxBlockAgeMilliseconds()).thenReturn(MAX_BLOCK_AGE_MS);
    when(livenessPluginConfiguration.gasLimit()).thenReturn(GAS_LIMIT);
    when(livenessPluginConfiguration.gasPriceGwei()).thenReturn(GAS_PRICE_GWEI);
    when(metricCategoryRegistry.isMetricCategoryEnabled(any())).thenReturn(true);

    plugin.register(serviceManager);
    plugin.start();

    assertThatNoException().isThrownBy(() -> plugin.stop());
  }

  @Test
  public void shouldStopGracefullyWhenNotStarted() {
    assertThatNoException().isThrownBy(() -> plugin.stop());
  }

  @Test
  public void shouldHandleMetricsDisabled() {
    when(livenessPluginConfiguration.enabled()).thenReturn(true);
    when(livenessPluginConfiguration.metricCategoryEnabled()).thenReturn(false); // Metrics disabled
    when(livenessPluginConfiguration.contractAddress()).thenReturn(CONTRACT_ADDRESS);
    when(livenessPluginConfiguration.signerUrl()).thenReturn(SIGNER_URL);
    when(livenessPluginConfiguration.signerKeyId()).thenReturn(SIGNER_KEY_ID);
    when(livenessPluginConfiguration.signerAddress()).thenReturn(SIGNER_ADDRESS);
    when(livenessPluginConfiguration.checkIntervalMilliseconds()).thenReturn(CHECK_INTERVAL_MS);
    when(livenessPluginConfiguration.maxBlockAgeMilliseconds()).thenReturn(MAX_BLOCK_AGE_MS);
    when(livenessPluginConfiguration.gasLimit()).thenReturn(GAS_LIMIT);
    when(livenessPluginConfiguration.gasPriceGwei()).thenReturn(GAS_PRICE_GWEI);

    plugin.register(serviceManager);

    assertThatNoException().isThrownBy(() -> plugin.start());
  }

  // ======= NEW TESTS FOR BUG FIXES =======

  @Test
  public void shouldValidateGasLimitPositive() throws Exception {
    setupDefaultConfiguration();
    when(livenessPluginConfiguration.gasLimit()).thenReturn(50000L); // Valid gas limit

    plugin.register(serviceManager);
    plugin.start();

    var method = LivenessPlugin.class.getDeclaredMethod("getValidatedGasLimit");
    method.setAccessible(true);

    long result = (long) method.invoke(plugin);

    assertThat(result).isEqualTo(50000L);
  }

  @Test
  public void shouldThrowExceptionForZeroGasLimit() throws Exception {
    setupDefaultConfiguration();
    when(livenessPluginConfiguration.gasLimit()).thenReturn(0L); // Invalid gas limit

    plugin.register(serviceManager);
    plugin.start();

    var method = LivenessPlugin.class.getDeclaredMethod("getValidatedGasLimit");
    method.setAccessible(true);

    assertThatThrownBy(() -> method.invoke(plugin))
        .isInstanceOf(InvocationTargetException.class)
        .hasCauseInstanceOf(IOException.class)
        .hasRootCauseMessage("Gas limit must be positive, but was: 0");
  }

  @Test
  public void shouldThrowExceptionForNegativeGasLimit() throws Exception {
    setupDefaultConfiguration();
    when(livenessPluginConfiguration.gasLimit()).thenReturn(-1000L); // Invalid gas limit

    plugin.register(serviceManager);
    plugin.start();

    var method = LivenessPlugin.class.getDeclaredMethod("getValidatedGasLimit");
    method.setAccessible(true);

    assertThatThrownBy(() -> method.invoke(plugin))
        .isInstanceOf(InvocationTargetException.class)
        .hasCauseInstanceOf(IOException.class)
        .hasRootCauseMessage("Gas limit must be positive, but was: -1000");
  }

  @Test
  public void shouldUseMinimumGasLimitWhenTooLow() throws Exception {
    setupDefaultConfiguration();
    when(livenessPluginConfiguration.gasLimit()).thenReturn(1000L); // Below minimum

    plugin.register(serviceManager);
    plugin.start();

    var method = LivenessPlugin.class.getDeclaredMethod("getValidatedGasLimit");
    method.setAccessible(true);

    long result = (long) method.invoke(plugin);

    assertThat(result).isEqualTo(21000L); // Should use minimum
  }

  @Test
  public void shouldUseMaximumGasLimitWhenTooHigh() throws Exception {
    setupDefaultConfiguration();
    when(livenessPluginConfiguration.gasLimit()).thenReturn(50_000_000L); // Above maximum

    plugin.register(serviceManager);
    plugin.start();

    var method = LivenessPlugin.class.getDeclaredMethod("getValidatedGasLimit");
    method.setAccessible(true);

    long result = (long) method.invoke(plugin);

    assertThat(result).isEqualTo(10_000_000L); // Should use maximum
  }

  @Test
  public void shouldUseConfiguredGasPriceWhenPositive() throws Exception {
    setupDefaultConfiguration();
    when(livenessPluginConfiguration.gasPriceGwei()).thenReturn(5L); // 5 Gwei

    plugin.register(serviceManager);
    plugin.start();

    var method = LivenessPlugin.class.getDeclaredMethod("getGasPrice");
    method.setAccessible(true);

    org.hyperledger.besu.datatypes.Wei result =
        (org.hyperledger.besu.datatypes.Wei) method.invoke(plugin);

    // 5 Gwei = 5 * 1_000_000_000 Wei
    assertThat(result.getAsBigInteger().longValue()).isEqualTo(5_000_000_000L);
  }
}
