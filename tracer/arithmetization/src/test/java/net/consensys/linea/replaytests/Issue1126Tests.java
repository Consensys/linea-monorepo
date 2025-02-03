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
import static net.consensys.linea.testing.ReplayExecutionEnvironment.LINEA_MAINNET;

import net.consensys.linea.UnitTestWatcher;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

/**
 * The following ranges blew up the MMU circuit:
 *
 * <p>* 3108622-3108633
 *
 * <p>* 3175608-3175636
 *
 * <p>* 3432730-3432768
 *
 * <p>* 4392225-4392280
 *
 * <p>* 4477086-4477226
 *
 * <p>See https://github.com/Consensys/linea-tracer/issues/1121
 */
@Tag("nightly")
@Disabled
@ExtendWith(UnitTestWatcher.class)
public class Issue1126Tests {
  @Test
  void issue_1126_3108622_3108633() {
    replay(LINEA_MAINNET, "3108622-3108633.mainnet.json.gz");
  }

  @Test
  void issue_1126_3175608_3175636() {
    replay(LINEA_MAINNET, "3175608-3175636.mainnet.json.gz");
  }

  @Test
  void issue_1126_3432730_3432768() {
    replay(LINEA_MAINNET, "3432730-3432768.mainnet.json.gz");
  }

  @Test
  void issue_1126_4392225_4392280() {
    replay(LINEA_MAINNET, "4392225-4392280.mainnet.json.gz");
  }

  @Test
  void issue_1126_4477086_4477226() {
    replay(LINEA_MAINNET, "4477086-4477226.mainnet.json.gz");
  }

  @Test
  void issue_1126_3290673_3290679() {
    replay(LINEA_MAINNET, "3290673-3290679.mainnet.json.gz");
  }

  @Test
  void issue_1126_3290746_3290752() {
    replay(LINEA_MAINNET, "3290746-3290752.mainnet.json.gz");
  }

  @Test
  void issue_1126_3315684_3315690() {
    replay(LINEA_MAINNET, "3315684-3315690.mainnet.json.gz");
  }

  @Test
  void issue_1126_3374278_3374284() {
    replay(LINEA_MAINNET, "3374278-3374284.mainnet.json.gz");
  }

  @Test
  void issue_1126_3385404_3385411() {
    replay(LINEA_MAINNET, "3385404-3385411.mainnet.json.gz");
  }

  @Test
  void issue_1126_3410170_3410240() {
    replay(LINEA_MAINNET, "3410170-3410240.mainnet.json.gz");
  }

  @Test
  void issue_1126_3423488_3423521() {
    replay(LINEA_MAINNET, "3423488-3423521.mainnet.json.gz");
  }

  @Test
  void issue_1126_3424829_3424864() {
    replay(LINEA_MAINNET, "3424829-3424864.mainnet.json.gz");
  }

  @Test
  void issue_1126_3429701_3429735() {
    replay(LINEA_MAINNET, "3429701-3429735.mainnet.json.gz");
  }

  @Test
  void issue_1126_3431193_3431232() {
    replay(LINEA_MAINNET, "3431193-3431232.mainnet.json.gz");
  }

  @Test
  void issue_1126_3431567_3431601() {
    replay(LINEA_MAINNET, "3431567-3431601.mainnet.json.gz");
  }

  @Test
  void issue_1126_4323962_4324012() {
    replay(LINEA_MAINNET, "4323962-4324012.mainnet.json.gz");
  }

  @Test
  void issue_1126_4343434_4343473() {
    replay(LINEA_MAINNET, "4343434-4343473.mainnet.json.gz");
  }

  @Test
  void issue_1126_4519246_4519309() {
    replay(LINEA_MAINNET, "4519246-4519309.mainnet.json.gz");
  }

  @Test
  void issue_1126_4556007_4556115() {
    replay(LINEA_MAINNET, "4556007-4556115.mainnet.json.gz");
  }

  @Test
  void issue_1126_4583379_4583463() {
    replay(LINEA_MAINNET, "4583379-4583463.mainnet.json.gz");
  }
}
