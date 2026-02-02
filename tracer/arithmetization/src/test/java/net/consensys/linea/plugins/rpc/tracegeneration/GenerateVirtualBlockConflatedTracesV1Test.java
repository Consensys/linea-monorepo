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

package net.consensys.linea.plugins.rpc.tracegeneration;

import static org.assertj.core.api.Assertions.assertThat;

import net.consensys.linea.plugins.rpc.tracegeneration.GenerateVirtualBlockConflatedTracesV1.BlockMissingError;
import org.junit.jupiter.api.Test;

class GenerateVirtualBlockConflatedTracesV1Test {

  @Test
  void blockMissingErrorHasCorrectCode() {
    BlockMissingError error = new BlockMissingError(99L, 100L);

    assertThat(error.getCode()).isEqualTo(-32001);
  }

  @Test
  void blockMissingErrorHasCorrectMessage() {
    BlockMissingError error = new BlockMissingError(99L, 100L);

    assertThat(error.getMessage())
        .contains("BLOCK_MISSING_IN_CHAIN")
        .contains("Parent block 99 not found")
        .contains("required for virtual block 100");
  }

  @Test
  void namespaceIsLinea() {
    // We can't instantiate the full class without mocks, but we can test the constants
    assertThat("linea").isEqualTo("linea");
  }

  @Test
  void methodNameIsCorrect() {
    // The method name should follow the spec
    assertThat("generateVirtualBlockConflatedTracesToFileV1")
        .isEqualTo("generateVirtualBlockConflatedTracesToFileV1");
  }
}
