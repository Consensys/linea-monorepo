/*
 * Copyright ConsenSys AG.
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
package net.consensys.besu.tracing.rpc;

import static com.google.common.base.Preconditions.checkArgument;

import java.util.concurrent.atomic.AtomicReference;

import com.google.auto.service.AutoService;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.RpcEndpointService;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;

/** Test plugin with RPC endpoint */
@AutoService(BesuPlugin.class)
public class TestRpcEndpointServicePlugin implements BesuPlugin {

  /** namespace for custom trace RPC methods */
  public static final String NAMESPACE = "tests";

  private final AtomicReference<String> stringStorage = new AtomicReference<>("InitialValue");
  private final AtomicReference<Object[]> arrayStorage = new AtomicReference<>();

  /** Test plugin with RPC endpoint */
  public TestRpcEndpointServicePlugin() {
    System.out.println("made a new plugin with string storage. value=" + stringStorage.get());
  }

  private String setValue(final PluginRpcRequest request) {
    checkArgument(request.getParams().length == 1, "Only one parameter accepted");
    return stringStorage.updateAndGet(x -> request.getParams()[0].toString());
  }

  private String getValue(final PluginRpcRequest request) {
    return stringStorage.get();
  }

  private Object[] replaceValueList(final PluginRpcRequest request) {
    return arrayStorage.updateAndGet(x -> request.getParams());
  }

  private String throwException(final PluginRpcRequest request) {
    throw new RuntimeException("Kaboom");
  }

  @Override
  public void register(final BesuContext context) {
    System.out.println("Registering RPC plugin");
    context
        .getService(RpcEndpointService.class)
        .ifPresent(
            rpcEndpointService -> {
              System.out.println("Registering RPC plugin endpoints");
              rpcEndpointService.registerRPCEndpoint(NAMESPACE, "getValue", this::getValue);
              rpcEndpointService.registerRPCEndpoint(NAMESPACE, "setValue", this::setValue);
              rpcEndpointService.registerRPCEndpoint(
                  NAMESPACE, "replaceValueList", this::replaceValueList);
              rpcEndpointService.registerRPCEndpoint(
                  NAMESPACE, "throwException", this::throwException);
              rpcEndpointService.registerRPCEndpoint("notEnabled", "getValue", this::getValue);
            });
  }

  @Override
  public void start() {
    System.out.println("started TestRpcEndpointServicePlugin plugin");
  }

  @Override
  public void stop() {
    System.out.println("stopped TestRpcEndpointServicePlugin plugin");
  }
}
