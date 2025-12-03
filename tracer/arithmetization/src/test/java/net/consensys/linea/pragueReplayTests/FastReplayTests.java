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
package net.consensys.linea.pragueReplayTests;

import static net.consensys.linea.ReplayTestTools.replay;
import static net.consensys.linea.zktracer.ChainConfig.MAINNET_TESTCONFIG;
import static net.consensys.linea.zktracer.Fork.PRAGUE;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;

@Tag("replay")
@ExtendWith(UnitTestWatcher.class)
public class FastReplayTests extends TracerTestBase {
  @Test
  void block_25080000_25080001(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080000-25080001.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080002_25080003(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080002-25080003.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080004_25080005(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080004-25080005.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080006_25080007(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080006-25080007.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080008_25080009(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080008-25080009.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080010_25080011(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080010-25080011.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080012_25080013(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080012-25080013.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080014_25080015(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080014-25080015.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080016_25080017(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080016-25080017.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080018_25080019(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080018-25080019.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080020_25080021(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080020-25080021.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080022_25080023(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080022-25080023.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080025_25080029(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080025-25080029.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080030_25080034(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080030-25080034.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080035_25080039(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080035-25080039.mainnet.prague.json.gz", testInfo);
  }
}
