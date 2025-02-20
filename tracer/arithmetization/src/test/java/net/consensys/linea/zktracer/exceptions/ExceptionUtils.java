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
package net.consensys.linea.zktracer.exceptions;

import static net.consensys.linea.zktracer.module.hub.signals.TracedException.OUT_OF_GAS_EXCEPTION;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotEquals;

import net.consensys.linea.testing.BytecodeRunner;

public class ExceptionUtils {
  static void assertEqualsOutOfGasIfCornerCaseMinusOneElseAssertNotEquals(
      int cornerCase, BytecodeRunner bytecodeRunner) {
    if (cornerCase == -1) {
      assertEquals(
          OUT_OF_GAS_EXCEPTION,
          bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
    } else {
      assertNotEquals(
          OUT_OF_GAS_EXCEPTION,
          bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
    }
  }
}
