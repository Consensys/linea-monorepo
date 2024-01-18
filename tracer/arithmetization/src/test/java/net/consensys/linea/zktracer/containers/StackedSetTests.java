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

package net.consensys.linea.zktracer.containers;

import java.math.BigInteger;

import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.add.AddOperation;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

public class StackedSetTests {
  @Test
  public void push() {
    StackedSet<AddOperation> chunks = new StackedSet<>();
    chunks.enter();
    AddOperation addOperation =
        new AddOperation(
            OpCode.ADD,
            Bytes.wrap(BigInteger.ONE.toByteArray()),
            Bytes.wrap(BigInteger.ONE.toByteArray()));
    chunks.add(addOperation);
    chunks.add(addOperation);
    chunks.add(addOperation);
    Assertions.assertEquals(1, chunks.size());
    chunks.pop();
    Assertions.assertEquals(0, chunks.size());
  }
}
