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

package net.consensys.linea.zktracer.specs;

import static net.consensys.linea.zktracer.module.alu.add.Add.ADD_JSON_KEY;

import net.consensys.linea.zktracer.AbstractModuleBySpecTest;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.alu.add.Add;

/** Implementation of a module tracer by spec class for the ADD module. */
public class AddTracerBySpecTest extends AbstractModuleBySpecTest {

  public static Object[][] specs() {
    return findSpecFiles(ADD_JSON_KEY);
  }

  @Override
  protected Module getModuleTracer() {
    return new Add();
  }
}
