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
package net.consensys.linea.replaytests;

import static net.consensys.linea.replaytests.ReplayTestTools.replay;
import static net.consensys.linea.testing.ReplayExecutionEnvironment.LINEA_SEPOLIA;

import net.consensys.linea.UnitTestWatcher;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

/**
 * This test contains a STATICCALL to the ECRECOVER precompile in the first transaction, at
 * HUB_STAMP = 1187 (before update). This call is special and we should extract several unit tests
 * from it.
 *
 * <p>At the time of the STATICCALL the context owns 3124 gas. With no value transfer (and hence no
 * account creation) and (I believe) no memory expansion the net upfront gas cost of this STATICCALL
 * is 100. The current context is therefore left with 3024 gas after having paid for the upfront gas
 * cost. The STATICCALL allows the transfer of all its GAS (=3124) to the child context. That is: it
 * allows it to transfer (after upfronts costs) all 3024 remaing units of gas.
 *
 * <p>However, the (63/64)-ths rule kicks in and the child may only receive 3024 - (3024/64) = 3024
 * - 47 = 2977 to its child. With the cost of ECRECOVER being set at 3000 ECRECOVER fails in the
 * only way it knows how to and we are left with a 0 on the stack.
 *
 * <p><a href="https://github.com/Consensys/linea-tracer/issues/1153">Related GitHub issue</a>
 */
@Tag("replay")
@Tag("nightly")
@ExtendWith(UnitTestWatcher.class)
public class Issue1136Tests {

  @Test
  void issue_1136_block_3110546() {
    replay(LINEA_SEPOLIA, "3110546.sepolia.json.gz");
  }
}
