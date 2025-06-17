/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.rpc.services;

import com.google.auto.service.AutoService;
import net.consensys.linea.AbstractLineaRequiredPlugin;
import net.consensys.linea.rpc.methods.LineaCancelBundle;
import net.consensys.linea.rpc.methods.LineaSendBundle;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.ServiceManager;

@AutoService(BesuPlugin.class)
public class LineaBundleEndpointsPlugin extends AbstractLineaRequiredPlugin {
  private LineaSendBundle lineaSendBundleMethod;
  private LineaCancelBundle lineaCancelBundleMethod;

  /**
   * Register the bundle RPC service.
   *
   * @param serviceManager the ServiceManager to be used.
   */
  @Override
  public void doRegister(final ServiceManager serviceManager) {
    lineaSendBundleMethod = new LineaSendBundle(blockchainService);

    rpcEndpointService.registerRPCEndpoint(
        lineaSendBundleMethod.getNamespace(),
        lineaSendBundleMethod.getName(),
        lineaSendBundleMethod::execute);

    lineaCancelBundleMethod = new LineaCancelBundle();

    rpcEndpointService.registerRPCEndpoint(
        lineaCancelBundleMethod.getNamespace(),
        lineaCancelBundleMethod.getName(),
        lineaCancelBundleMethod::execute);
  }

  /**
   * Starts this plugin and in case the extra data pricing is enabled, as first thing it tries to
   * extract extra data pricing configuration from the chain head, then it starts listening for new
   * imported block, in order to update the extra data pricing on every incoming block.
   */
  @Override
  public void doStart() {
    // set the pool
    lineaSendBundleMethod.init(bundlePoolService);
    lineaCancelBundleMethod.init(bundlePoolService);
  }

  @Override
  public void stop() {
    bundlePoolService.saveToDisk();
    super.stop();
  }
}
