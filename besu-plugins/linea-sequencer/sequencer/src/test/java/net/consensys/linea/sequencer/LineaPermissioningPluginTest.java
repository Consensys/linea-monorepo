/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Optional;
import net.consensys.linea.config.LineaPermissioningConfiguration;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.plugin.ServiceManager;
import org.hyperledger.besu.plugin.services.PermissioningService;
import org.hyperledger.besu.plugin.services.permissioning.TransactionPermissioningProvider;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class LineaPermissioningPluginTest {

  @Mock private ServiceManager serviceManager;
  @Mock private PermissioningService permissioningService;
  @Mock private Transaction transaction;
  @Mock private LineaPermissioningConfiguration lineaPermissioningConfiguration;

  private LineaPermissioningPlugin plugin;

  @BeforeEach
  public void setUp() {
    plugin =
        new LineaPermissioningPlugin() {
          @Override
          public LineaPermissioningConfiguration permissioningConfiguration() {
            return lineaPermissioningConfiguration;
          }
        };
    when(serviceManager.getService(PermissioningService.class))
        .thenReturn(Optional.of(permissioningService));
  }

  @Test
  public void shouldRegisterWithServiceManager() {
    // Act
    plugin.doRegister(serviceManager);

    // Assert
    verify(serviceManager).getService(PermissioningService.class);
  }

  @Test
  public void shouldThrowExceptionWhenPermissioningServiceNotAvailable() {
    // Arrange
    when(serviceManager.getService(PermissioningService.class)).thenReturn(Optional.empty());

    // Act/Assert
    assertThatThrownBy(() -> plugin.doRegister(serviceManager))
        .isInstanceOf(RuntimeException.class)
        .hasMessageContaining("Failed to obtain PermissioningService from the ServiceManager");
  }

  @Test
  public void shouldRegisterTransactionPermissioningProvider() {
    // Arrange
    plugin.doRegister(serviceManager);

    // Act
    plugin.doStart();

    // Assert
    verify(permissioningService)
        .registerTransactionPermissioningProvider(any(TransactionPermissioningProvider.class));
  }

  @Test
  public void shouldRejectBlobTransactionsByDefault() {
    // Arrange
    when(lineaPermissioningConfiguration.blobTxEnabled()).thenReturn(false);
    plugin.doRegister(serviceManager);
    plugin.doStart();
    final TransactionPermissioningProvider provider = this.getTransactionPermissioningProvider();

    // Act - BLOB transaction
    when(transaction.getType()).thenReturn(TransactionType.BLOB);
    boolean result = provider.isPermitted(transaction);

    // Assert
    assertThat(result).isFalse();
  }

  @Test
  public void shouldPermitEIP7702Transactions() {
    // Arrange
    when(lineaPermissioningConfiguration.blobTxEnabled()).thenReturn(false);
    plugin.doRegister(serviceManager);
    plugin.doStart();
    final TransactionPermissioningProvider provider = this.getTransactionPermissioningProvider();

    // Act - EIP7702 transaction
    when(transaction.getType()).thenReturn(TransactionType.DELEGATE_CODE);
    boolean result = provider.isPermitted(transaction);

    // Assert
    assertThat(result).isTrue();
  }

  @Test
  public void shouldPermitLegacyTransactions() {
    // Arrange
    when(lineaPermissioningConfiguration.blobTxEnabled()).thenReturn(false);
    plugin.doRegister(serviceManager);
    plugin.doStart();
    final TransactionPermissioningProvider provider = this.getTransactionPermissioningProvider();

    // Act - LEGACY/FRONTIER transaction
    when(transaction.getType()).thenReturn(TransactionType.FRONTIER);
    boolean result = provider.isPermitted(transaction);

    // Assert
    assertThat(result).isTrue();
  }

  @Test
  public void shouldPermitAccessListTransactions() {
    // Arrange
    when(lineaPermissioningConfiguration.blobTxEnabled()).thenReturn(false);
    plugin.doRegister(serviceManager);
    plugin.doStart();
    final TransactionPermissioningProvider provider = this.getTransactionPermissioningProvider();

    // Act - ACCESS_LIST transaction
    when(transaction.getType()).thenReturn(TransactionType.ACCESS_LIST);
    boolean result = provider.isPermitted(transaction);

    // Assert
    assertThat(result).isTrue();
  }

  @Test
  public void shouldPermitEIP1559Transactions() {
    // Arrange
    when(lineaPermissioningConfiguration.blobTxEnabled()).thenReturn(false);
    plugin.doRegister(serviceManager);
    plugin.doStart();
    final TransactionPermissioningProvider provider = this.getTransactionPermissioningProvider();

    // Act - EIP1559 transaction
    when(transaction.getType()).thenReturn(TransactionType.EIP1559);
    boolean result = provider.isPermitted(transaction);

    // Assert
    assertThat(result).isTrue();
  }

  @Test
  public void shouldPermitBlobTransactionsWhenEnabled() {
    // Arrange
    when(lineaPermissioningConfiguration.blobTxEnabled()).thenReturn(true);
    plugin.doRegister(serviceManager);
    plugin.doStart();
    final TransactionPermissioningProvider provider = this.getTransactionPermissioningProvider();

    // Act - BLOB transaction
    when(transaction.getType()).thenReturn(TransactionType.BLOB);
    boolean result = provider.isPermitted(transaction);

    // Assert
    assertThat(result).isTrue();
  }

  private TransactionPermissioningProvider getTransactionPermissioningProvider() {
    ArgumentCaptor<TransactionPermissioningProvider> providerCaptor =
        ArgumentCaptor.forClass(TransactionPermissioningProvider.class);
    verify(permissioningService).registerTransactionPermissioningProvider(providerCaptor.capture());
    return providerCaptor.getValue();
  }
}
