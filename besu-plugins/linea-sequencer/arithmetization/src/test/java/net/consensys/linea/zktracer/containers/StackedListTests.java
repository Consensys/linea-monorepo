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

import static org.assertj.core.api.Assertions.assertThat;

import java.math.BigInteger;

import com.google.common.collect.ImmutableList;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import net.consensys.linea.zktracer.module.add.AddOperation;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

public class StackedListTests {
  private static final AddOperation ONE_PLUS_ONE =
      new AddOperation(
          OpCode.ADD,
          Bytes.wrap(BigInteger.ONE.toByteArray()),
          Bytes.wrap(BigInteger.ONE.toByteArray()));

  private static final AddOperation ONE_PLUS_TWO =
      new AddOperation(
          OpCode.ADD,
          Bytes.wrap(BigInteger.ONE.toByteArray()),
          Bytes.wrap(BigInteger.TWO.toByteArray()));

  @RequiredArgsConstructor
  private static class IntegerModuleOperation extends ModuleOperation {
    private final int x;

    @Override
    protected int computeLineCount() {
      return x;
    }
  }

  @Test
  void testAddedToFront() {
    final StackedList<IntegerModuleOperation> state = new StackedList<>();

    state.enter();
    state.add(new IntegerModuleOperation(1));
    assertThat(state.lineCount()).isEqualTo(1);

    state.enter();
    state.add(new IntegerModuleOperation(3));
    assertThat(state.lineCount()).isEqualTo(4);

    state.pop();
    assertThat(state.lineCount()).isEqualTo(1);
  }

  @Test
  public void push() {
    StackedList<AddOperation> chunks = new StackedList<>();
    chunks.enter();

    chunks.add(ONE_PLUS_ONE);
    chunks.add(ONE_PLUS_ONE);
    chunks.add(ONE_PLUS_ONE);
    Assertions.assertEquals(3, chunks.size());
    chunks.pop();
    Assertions.assertEquals(0, chunks.size());
  }

  @Test
  public void multiplePushPop() {
    StackedList<AddOperation> chunks = new StackedList<>();
    chunks.enter();
    chunks.add(ONE_PLUS_ONE);
    chunks.add(ONE_PLUS_ONE);
    Assertions.assertEquals(2, chunks.size());

    chunks.enter();
    chunks.add(ONE_PLUS_ONE);
    Assertions.assertEquals(3, chunks.size());
    chunks.add(ONE_PLUS_TWO);
    Assertions.assertEquals(4, chunks.size());
    chunks.pop();
    Assertions.assertEquals(2, ImmutableList.copyOf(chunks.iterator()).size());
  }
}
