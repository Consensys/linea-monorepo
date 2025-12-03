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

@Tag("nightly")
@Tag("replay")
@ExtendWith(UnitTestWatcher.class)
public class NightlyReplayTests extends TracerTestBase {

  // ============================================================================
  // Blocks 25080040 -- 25080104
  // ============================================================================

  @Test
  void block_25080040_25080044(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080040-25080044.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080045_25080049(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080045-25080049.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080050_25080054(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080050-25080054.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080055_25080059(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080055-25080059.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080060_25080064(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080060-25080064.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080065_25080069(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080065-25080069.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080070_25080074(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080070-25080074.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080075_25080079(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080075-25080079.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080080_25080084(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080080-25080084.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080085_25080089(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080085-25080089.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080090_25080094(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080090-25080094.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080095_25080099(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080095-25080099.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25080100_25080104(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25080100-25080104.mainnet.prague.json.gz", testInfo);
  }

  // ==========================================================================
  // Blocks 25091000 -- 25091099
  // ==========================================================================

  @Test
  void block_25091000_25091004(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091000-25091004.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091005_25091009(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091005-25091009.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091010_25091014(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091010-25091014.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091015_25091019(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091015-25091019.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091020_25091024(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091020-25091024.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091025_25091029(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091025-25091029.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091030_25091034(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091030-25091034.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091035_25091039(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091035-25091039.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091040_25091044(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091040-25091044.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091045_25091049(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091045-25091049.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091050_25091054(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091050-25091054.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091055_25091059(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091055-25091059.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091060_25091064(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091060-25091064.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091065_25091069(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091065-25091069.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091070_25091074(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091070-25091074.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091075_25091079(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091075-25091079.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091080_25091084(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091080-25091084.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091085_25091089(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091085-25091089.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091090_25091094(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091090-25091094.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25091095_25091099(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25091095-25091099.mainnet.prague.json.gz", testInfo);
  }

  // ==========================================================================
  // Blocks 25098000 -- 25098099
  // ==========================================================================

  @Test
  void block_25098000_25098004(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098000-25098004.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098005_25098009(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098005-25098009.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098010_25098014(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098010-25098014.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098015_25098019(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098015-25098019.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098020_25098024(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098020-25098024.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098025_25098029(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098025-25098029.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098030_25098034(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098030-25098034.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098035_25098039(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098035-25098039.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098040_25098044(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098040-25098044.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098045_25098049(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098045-25098049.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098050_25098054(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098050-25098054.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098055_25098059(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098055-25098059.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098060_25098064(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098060-25098064.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098065_25098069(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098065-25098069.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098070_25098074(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098070-25098074.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098075_25098079(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098075-25098079.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098080_25098084(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098080-25098084.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098085_25098089(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098085-25098089.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098090_25098094(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098090-25098094.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25098095_25098099(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25098095-25098099.mainnet.prague.json.gz", testInfo);
  }

  // ==========================================================================
  // Blocks 25102000 -- 25103004
  // ==========================================================================

  @Test
  void block_25102000_25102004(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102000-25102004.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102005_25102009(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102005-25102009.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102010_25102014(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102010-25102014.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102015_25102019(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102015-25102019.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102020_25102024(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102020-25102024.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102025_25102029(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102025-25102029.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102030_25102034(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102030-25102034.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102035_25102039(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102035-25102039.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102040_25102044(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102040-25102044.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102045_25102049(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102045-25102049.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102050_25102054(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102050-25102054.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102055_25102059(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102055-25102059.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102060_25102064(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102060-25102064.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102065_25102069(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102065-25102069.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102070_25102074(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102070-25102074.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102075_25102079(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102075-25102079.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102080_25102084(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102080-25102084.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102085_25102089(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102085-25102089.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102090_25102094(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102090-25102094.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102095_25102099(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102095-25102099.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102100_25102104(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102100-25102104.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102105_25102109(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102105-25102109.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102110_25102114(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102110-25102114.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102115_25102119(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102115-25102119.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102120_25102124(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102120-25102124.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102125_25102129(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102125-25102129.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102130_25102134(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102130-25102134.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102135_25102139(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102135-25102139.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102140_25102144(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102140-25102144.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102145_25102149(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102145-25102149.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102150_25102154(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102150-25102154.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102155_25102159(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102155-25102159.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102160_25102164(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102160-25102164.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102165_25102169(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102165-25102169.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102170_25102174(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102170-25102174.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102175_25102179(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102175-25102179.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102180_25102184(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102180-25102184.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102185_25102189(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102185-25102189.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102190_25102194(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102190-25102194.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102195_25102199(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102195-25102199.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102200_25102204(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102200-25102204.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102205_25102209(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102205-25102209.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102210_25102214(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102210-25102214.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102215_25102219(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102215-25102219.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102220_25102224(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102220-25102224.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102225_25102229(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102225-25102229.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102230_25102234(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102230-25102234.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102235_25102239(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102235-25102239.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102240_25102244(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102240-25102244.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102245_25102249(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102245-25102249.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102250_25102254(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102250-25102254.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102255_25102259(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102255-25102259.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102260_25102264(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102260-25102264.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102265_25102269(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102265-25102269.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102270_25102274(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102270-25102274.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102275_25102279(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102275-25102279.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102280_25102284(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102280-25102284.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102285_25102289(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102285-25102289.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102290_25102294(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102290-25102294.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102295_25102299(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102295-25102299.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102300_25102304(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102300-25102304.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102305_25102309(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102305-25102309.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102310_25102314(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102310-25102314.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102315_25102319(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102315-25102319.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102320_25102324(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102320-25102324.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102325_25102329(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102325-25102329.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102330_25102334(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102330-25102334.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102335_25102339(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102335-25102339.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102340_25102344(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102340-25102344.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102345_25102349(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102345-25102349.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102350_25102354(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102350-25102354.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102355_25102359(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102355-25102359.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102360_25102364(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102360-25102364.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102365_25102369(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102365-25102369.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102370_25102374(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102370-25102374.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102375_25102379(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102375-25102379.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102380_25102384(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102380-25102384.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102385_25102389(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102385-25102389.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102390_25102394(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102390-25102394.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102395_25102399(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102395-25102399.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102400_25102404(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102400-25102404.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102405_25102409(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102405-25102409.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102410_25102414(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102410-25102414.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102415_25102419(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102415-25102419.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102420_25102424(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102420-25102424.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102425_25102429(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102425-25102429.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102430_25102434(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102430-25102434.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102435_25102439(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102435-25102439.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102440_25102444(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102440-25102444.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102445_25102449(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102445-25102449.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102450_25102454(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102450-25102454.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102455_25102459(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102455-25102459.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102460_25102464(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102460-25102464.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102465_25102469(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102465-25102469.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102470_25102474(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102470-25102474.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102475_25102479(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102475-25102479.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102480_25102484(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102480-25102484.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102485_25102489(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102485-25102489.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102490_25102494(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102490-25102494.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102495_25102499(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102495-25102499.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102500_25102504(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102500-25102504.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102505_25102509(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102505-25102509.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102510_25102514(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102510-25102514.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102515_25102519(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102515-25102519.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102520_25102524(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102520-25102524.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102525_25102529(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102525-25102529.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102530_25102534(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102530-25102534.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102535_25102539(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102535-25102539.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102540_25102544(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102540-25102544.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102545_25102549(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102545-25102549.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102550_25102554(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102550-25102554.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102555_25102559(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102555-25102559.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102560_25102564(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102560-25102564.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102565_25102569(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102565-25102569.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102570_25102574(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102570-25102574.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102575_25102579(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102575-25102579.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102580_25102584(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102580-25102584.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102585_25102589(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102585-25102589.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102590_25102594(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102590-25102594.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102595_25102599(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102595-25102599.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102600_25102604(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102600-25102604.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102605_25102609(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102605-25102609.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102610_25102614(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102610-25102614.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102615_25102619(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102615-25102619.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102620_25102624(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102620-25102624.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102625_25102629(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102625-25102629.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102630_25102634(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102630-25102634.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102635_25102639(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102635-25102639.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102640_25102644(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102640-25102644.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102645_25102649(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102645-25102649.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102650_25102654(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102650-25102654.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102655_25102659(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102655-25102659.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102660_25102664(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102660-25102664.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102665_25102669(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102665-25102669.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102670_25102674(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102670-25102674.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102675_25102679(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102675-25102679.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102680_25102684(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102680-25102684.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102685_25102689(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102685-25102689.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102690_25102694(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102690-25102694.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102695_25102699(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102695-25102699.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102700_25102704(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102700-25102704.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102705_25102709(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102705-25102709.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102710_25102714(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102710-25102714.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102715_25102719(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102715-25102719.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102720_25102724(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102720-25102724.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102725_25102729(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102725-25102729.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102730_25102734(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102730-25102734.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102735_25102739(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102735-25102739.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102740_25102744(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102740-25102744.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102745_25102749(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102745-25102749.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102750_25102754(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102750-25102754.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102755_25102759(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102755-25102759.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102760_25102764(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102760-25102764.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102765_25102769(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102765-25102769.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102770_25102774(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102770-25102774.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102775_25102779(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102775-25102779.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102780_25102784(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102780-25102784.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102785_25102789(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102785-25102789.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102790_25102794(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102790-25102794.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102795_25102799(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102795-25102799.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102800_25102804(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102800-25102804.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102805_25102809(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102805-25102809.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102810_25102814(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102810-25102814.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102815_25102819(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102815-25102819.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102820_25102824(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102820-25102824.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102825_25102829(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102825-25102829.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102830_25102834(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102830-25102834.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102835_25102839(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102835-25102839.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102840_25102844(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102840-25102844.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102845_25102849(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102845-25102849.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102850_25102854(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102850-25102854.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102855_25102859(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102855-25102859.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102860_25102864(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102860-25102864.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102865_25102869(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102865-25102869.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102870_25102874(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102870-25102874.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102875_25102879(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102875-25102879.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102880_25102884(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102880-25102884.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102885_25102889(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102885-25102889.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102890_25102894(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102890-25102894.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102895_25102899(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102895-25102899.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102900_25102904(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102900-25102904.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102905_25102909(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102905-25102909.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102910_25102914(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102910-25102914.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102915_25102919(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102915-25102919.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102920_25102924(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102920-25102924.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102925_25102929(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102925-25102929.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102930_25102934(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102930-25102934.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102935_25102939(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102935-25102939.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102940_25102944(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102940-25102944.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102945_25102949(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102945-25102949.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102950_25102954(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102950-25102954.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102955_25102959(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102955-25102959.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102960_25102964(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102960-25102964.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102965_25102969(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102965-25102969.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102970_25102974(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102970-25102974.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102975_25102979(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102975-25102979.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102980_25102984(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102980-25102984.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102985_25102989(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102985-25102989.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102990_25102994(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102990-25102994.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25102995_25102999(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25102995-25102999.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25103000_25103004(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25103000-25103004.mainnet.prague.json.gz", testInfo);
  }
}
