/*
 * Copyright ConsenSys Inc.
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

import static com.google.common.base.Preconditions.checkArgument;
import static net.consensys.linea.testing.ToyExecutionEnvironmentV2.*;
import static net.consensys.linea.zktracer.ChainConfig.MAINNET_TESTCONFIG;
import static net.consensys.linea.zktracer.Fork.*;

import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;

import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.ExecutionEnvironment;
import net.consensys.linea.zktracer.ZkCounter;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.container.module.Module;
import org.hyperledger.besu.ethereum.core.BlockHeader;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.junit.jupiter.api.Test;

public class LineCountingTracerTest extends TracerTestBase {

  @Test
  void noDuplicateNamesInModules() {
    final ZkTracer tracer = new ZkTracer(chainConfig);
    final List<String> tracerToCount =
        tracer.getModulesToCount().stream().map(Module::moduleKey).toList();
    final List<String> tracedModules =
        tracer.getHub().getModulesToTrace().stream().map(Module::moduleKey).toList();
    checkArgument(tracerToCount.containsAll(tracedModules), "Some traced modules are not counted");
    checkArgument(
        tracerToCount.size() == tracerToCount.stream().distinct().toList().size(),
        "Duplicate has been found");

    final ZkCounter counter = new ZkCounter(chainConfig.bridgeConfiguration);
    final List<String> counterToCount =
        counter.getModulesToCount().stream().map(Module::moduleKey).toList();
    checkArgument(
        counterToCount.size() == counterToCount.stream().distinct().toList().size(),
        "Duplicate has been found");
  }

  @Test
  void sameModuleAcrossAllForkWithZkCounterAndTracer() {

    final List<String> londonTracer =
        new ZkTracer(MAINNET_TESTCONFIG(LONDON))
            .getModulesToCount().stream().map(Module::moduleKey).toList();
    final List<String> parisTracer =
        new ZkTracer(MAINNET_TESTCONFIG(PARIS))
            .getModulesToCount().stream().map(Module::moduleKey).toList();
    final List<String> shanghaiTracer =
        new ZkTracer(MAINNET_TESTCONFIG(SHANGHAI))
            .getModulesToCount().stream().map(Module::moduleKey).toList();
    final List<String> cancunTracer =
        new ZkTracer(MAINNET_TESTCONFIG(CANCUN))
            .getModulesToCount().stream().map(Module::moduleKey).toList();
    final List<String> pragueTracer =
        new ZkTracer(MAINNET_TESTCONFIG(PRAGUE))
            .getModulesToCount().stream().map(Module::moduleKey).toList();

    // check that paris ⊆ london
    for (String module : parisTracer) {
      checkArgument(
          londonTracer.contains(module), "Module " + module + " is in Paris but not in London");
    }

    // check that london ⊆ paris
    for (String module : londonTracer) {
      checkArgument(
          parisTracer.contains(module), "Module " + module + " is in London but not in Paris");
    }

    // check that shanghai ⊆ london
    for (String module : shanghaiTracer) {
      checkArgument(
          londonTracer.contains(module), "Module " + module + " is in Shanghai but not in London");
    }

    // check that london ⊆ Shanghai
    for (String module : londonTracer) {
      checkArgument(
          shanghaiTracer.contains(module),
          "Module " + module + " is in London but not in Shanghai");
    }

    // check that cancun ⊆ london
    for (String module : cancunTracer) {
      checkArgument(
          londonTracer.contains(module), "Module " + module + " is in Cancun but not in London");
    }

    // check that london ⊆ cancun
    for (String module : londonTracer) {
      checkArgument(
          cancunTracer.contains(module), "Module " + module + " is in London but not in Cancun");
    }

    // check that prague ⊆ london
    for (String module : pragueTracer) {
      checkArgument(
          londonTracer.contains(module), "Module " + module + " is in Prague but not in London");
    }

    // check that london ⊆ prague
    for (String module : londonTracer) {
      checkArgument(
          pragueTracer.contains(module), "Module " + module + " is in London but not in Prague");
    }
  }

  @Test
  void startBlockStuffAreNotPopped() {
    final WorldView world = WorldView.EMPTY;
    final BlockHeader blockHeader =
        ExecutionEnvironment.getLineaBlockHeaderBuilder(Optional.empty())
            .number(DEFAULT_BLOCK_NUMBER)
            .coinbase(DEFAULT_COINBASE_ADDRESS)
            .timestamp(DEFAULT_TIME_STAMP)
            .parentHash(DEFAULT_HASH)
            .baseFee(DEFAULT_BASE_FEE)
            .buildBlockHeader();

    final ZkTracer tracer = new ZkTracer(chainConfig);
    tracer.traceStartConflation(1);
    tracer.traceStartBlock(world, blockHeader, DEFAULT_COINBASE_ADDRESS);
    final Map<String, Integer> sizeBeforeTracer = tracer.getModulesLineCount();
    tracer.popTransactionBundle();
    final Map<String, Integer> sizeAfterTracer = tracer.getModulesLineCount();
    for (String module : sizeBeforeTracer.keySet()) {
      checkArgument(
          Objects.equals(sizeAfterTracer.get(module), sizeBeforeTracer.get(module)),
          "Tracer: some block stuff has been removed in Module " + module);
    }

    final ZkCounter counter = new ZkCounter(chainConfig.bridgeConfiguration);
    counter.traceStartConflation(1);
    counter.traceStartBlock(world, blockHeader, DEFAULT_COINBASE_ADDRESS);
    final Map<String, Integer> sizeBeforeCounter = counter.getModulesLineCount();
    counter.popTransactionBundle();
    final Map<String, Integer> sizeAfterCounter = counter.getModulesLineCount();
    for (String module : sizeBeforeCounter.keySet()) {
      checkArgument(
          Objects.equals(sizeAfterCounter.get(module), sizeBeforeCounter.get(module)),
          "Counter: some block stuff has been removed in Module " + module);
    }
  }
}
