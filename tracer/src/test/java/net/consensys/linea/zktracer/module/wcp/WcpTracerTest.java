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
package net.consensys.linea.zktracer.module.wcp;

import static net.consensys.linea.zktracer.OpCode.GT;

import java.util.List;
import java.util.stream.Stream;

import net.consensys.linea.zktracer.AbstractModuleTracerCorsetTest;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.provider.Arguments;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
class WcpTracerTest extends AbstractModuleTracerCorsetTest {

  @Override
  protected ModuleTracer getModuleTracer() {
    return new WcpTracer();
  }

  @Override
  protected Stream<Arguments> provideNonRandomArguments() {
    Bytes32 arg1 =
        Bytes32.fromHexString("0xdcd5cf52e4daec5389587d0d0e996e6ce2d0546b63d3ea0a0dc48ad984d180a9");
    Bytes32 arg2 =
        Bytes32.fromHexString("0x0479484af4a59464a48818b3980174687661bafb13d06f49537995fa6c02159e");
    return Stream.of(Arguments.of(GT, List.of(arg1, arg2)));
  }
}
