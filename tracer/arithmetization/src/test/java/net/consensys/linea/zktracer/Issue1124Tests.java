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

import static net.consensys.linea.testing.ReplayExecutionEnvironment.LINEA_MAINNET;
import static net.consensys.linea.zktracer.ReplayTests.replay;

import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

@Tag("nightly")
public class Issue1124Tests {

  // STP constraints were failing for these ranges

  // @Disabled
  @Test
  void issue_1124_range_4323962_4324012() {
    replay(LINEA_MAINNET, "4323962-4324012.json.gz", false);
  }

  // @Disabled
  @Test
  void issue_1124_range_4343434_4343473() {
    replay(LINEA_MAINNET, "4343434-4343473.json.gz", false);
  }
}
