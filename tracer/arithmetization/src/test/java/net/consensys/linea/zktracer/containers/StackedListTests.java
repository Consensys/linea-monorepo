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

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import org.junit.jupiter.api.Test;

public class StackedListTests {
  @RequiredArgsConstructor
  private class IntegerModuleOperation extends ModuleOperation {
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
}
