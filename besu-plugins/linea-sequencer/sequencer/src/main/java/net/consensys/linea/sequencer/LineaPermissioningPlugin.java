/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.sequencer;

import com.google.auto.service.AutoService;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.AbstractLineaRequiredPlugin;
import net.consensys.linea.config.LineaPermissioningConfiguration;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.ServiceManager;
import org.hyperledger.besu.plugin.services.PermissioningService;

/**
 * This plugin uses the {@link PermissioningService} to filter transactions at multiple critical
 * lifecycle stages: i.) Block import - when a Besu validator receives a block via P2P gossip ii.)
 * Transaction pool - when a Besu node adds to its local transaction pool iii.) Block production -
 * when a Besu node builds a block
 *
 * <p>As PermissioningService executes rules over a broad scope, we may in the future consolidate
 * logic from the {@code LineaTransactionSelectorPlugin} and {@code
 * LineaTransactionPoolValidatorPlugin} to unify transaction filtering logic.
 *
 * <p>In addition to transaction permissioning, {@link PermissioningService} also supports node- and
 * message-level permissioning, which can be implemented to control peer connections and devp2p
 * message exchanges.
 */
@Slf4j
@AutoService(BesuPlugin.class)
public class LineaPermissioningPlugin extends AbstractLineaRequiredPlugin {
  private PermissioningService permissioningService;

  @Override
  public void doRegister(final ServiceManager serviceManager) {
    permissioningService =
        serviceManager
            .getService(PermissioningService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain PermissioningService from the ServiceManager."));
  }

  @Override
  public void doStart() {
    LineaPermissioningConfiguration config = permissioningConfiguration();
    log.info("Linea Permissioning Plugin starting with blobTxEnabled={}", config.blobTxEnabled());

    permissioningService.registerTransactionPermissioningProvider(
        (tx) -> {
          if (tx.getType() == TransactionType.BLOB && !config.blobTxEnabled()) return false;
          return true;
        });
  }

  @Override
  public void stop() {
    super.stop();
  }
}
