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

package net.consensys.linea.osakaReplayTests;

import static net.consensys.linea.ReplayTestTools.replay;
import static net.consensys.linea.zktracer.ChainConfig.MAINNET_TESTCONFIG;
import static net.consensys.linea.zktracer.Fork.OSAKA;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;

@Tag("replay")
@ExtendWith(UnitTestWatcher.class)
public class Incident1445 extends TracerTestBase {

  @Test
  void block_28279135_28279249(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(OSAKA), "osaka/incident-1445-28279135-28279249.json" , testInfo, false);
  }

  // Faulty block from block_28279135_28279249 test
  // This test can be run with Besu node setup
  @Test
  void block_2827980(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(OSAKA), "osaka/incident-1445-28279180.mainnet.json.gz", testInfo, false);
  }

}
