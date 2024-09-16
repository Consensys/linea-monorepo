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
public class Issue1169Tests {

  @Test
  public void issue_1145_block_3318494_InsufficientBalanceMainnet() {
    replay(LINEA_MAINNET, "2746060.mainnet.json.gz");
  }
}
