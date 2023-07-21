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

package net.consensys.linea.zktracer;

import java.math.BigInteger;
import java.util.HashMap;
import java.util.Map;

import net.consensys.linea.zktracer.module.lookuptable.ShfRtTrace;
import org.apache.tuweni.bytes.Bytes32;

public class ZkTraceBuilder {
  private final Map<String, Object> traceResults = new HashMap<>();

  public void addTrace(final String moduleName, final Object traceResult) {
    traceResults.put(moduleName, traceResult);
  }

  public ZkTrace build() {
    if (!traceResults.containsKey("shfRT")) {
      traceResults.put("shfRT", ShfRtTrace.generate());
    }

    return new ZkTrace(
        null, Bytes32.ZERO, BigInteger.ZERO, BigInteger.ZERO, BigInteger.ZERO, traceResults);
  }
}
