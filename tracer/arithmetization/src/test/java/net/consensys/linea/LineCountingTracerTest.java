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

import java.util.List;

import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.zktracer.ZkCounter;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.container.module.Module;
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
}
