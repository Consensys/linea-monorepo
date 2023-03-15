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
package tracers;

import net.consensys.zktracer.ZkTracer;
import net.consensys.zktracer.ZkTracerBuilder;

import java.io.IOException;

import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.hyperledger.besu.plugin.data.OperationTracerWrapper;
import tracing.OperationTracerPluginWrapper;

public class ZkTracerFactory implements TracerFactory {

  @Override
  public OperationTracerWrapper create(final JsonGenerator jsonGenerator) {
    final ObjectMapper mapper = new ObjectMapper();

    // This extends the ZkTracerBuilder interface
    ZkTracerBuilder zkTracerBuilder =
        (s, o) -> {
          try {
            mapper.writeValue(jsonGenerator, o);
          } catch (IOException e) {
            throw new RuntimeException(e);
          }
        };
    ZkTracer tracer = new ZkTracer(zkTracerBuilder);
    return OperationTracerPluginWrapper.create(tracer);
  }
}
