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

package net.consensys.linea.zktracer;

import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.tables.instructionDecoder.InstructionDecoder;
import net.consensys.linea.zktracer.module.tables.shf.ShfRtTrace;

public class ZkTraceBuilder {
  private final Map<String, Object> traceResults = new HashMap<>();

  public ZkTraceBuilder addTrace(Module module) {
    Optional.ofNullable(module.commit())
        .ifPresent(
            v -> {
              if (v.length() != module.lineCount()) {
                throw new IllegalStateException(
                    "["
                        + module.jsonKey()
                        + "] lines expected: "
                        + module.lineCount()
                        + " -- lines found: "
                        + v.length());
              }
              assert v.length() == module.lineCount();
              traceResults.put(module.jsonKey(), v);
            });
    return this;
  }

  public ZkTrace build() {
    // TODO: add other reference tables
    traceResults.put("shfRT", ShfRtTrace.generate());
    traceResults.put("instruction-decoder", InstructionDecoder.generate());

    return new ZkTrace(traceResults);
  }
}
