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

@Tag("weekly")
@Tag("replay")
@ExtendWith(UnitTestWatcher.class)
public class WeeklyReplayTests extends TracerTestBase {

  // ============================================================================
  // Blocks 25085000 --- 25086004
  // ============================================================================

  @Test
  void block_25085000_25085004(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085000-25085004.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085005_25085009(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085005-25085009.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085010_25085014(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085010-25085014.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085015_25085019(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085015-25085019.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085020_25085024(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085020-25085024.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085025_25085029(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085025-25085029.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085030_25085034(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085030-25085034.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085035_25085039(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085035-25085039.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085040_25085044(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085040-25085044.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085045_25085049(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085045-25085049.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085050_25085054(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085050-25085054.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085055_25085059(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085055-25085059.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085060_25085064(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085060-25085064.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085065_25085069(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085065-25085069.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085070_25085074(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085070-25085074.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085075_25085079(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085075-25085079.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085080_25085084(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085080-25085084.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085085_25085089(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085085-25085089.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085090_25085094(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085090-25085094.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085095_25085099(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085095-25085099.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085100_25085104(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085100-25085104.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085105_25085109(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085105-25085109.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085110_25085114(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085110-25085114.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085115_25085119(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085115-25085119.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085120_25085124(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085120-25085124.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085125_25085129(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085125-25085129.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085130_25085134(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085130-25085134.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085135_25085139(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085135-25085139.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085140_25085144(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085140-25085144.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085145_25085149(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085145-25085149.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085150_25085154(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085150-25085154.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085155_25085159(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085155-25085159.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085160_25085164(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085160-25085164.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085165_25085169(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085165-25085169.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085170_25085174(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085170-25085174.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085175_25085179(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085175-25085179.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085180_25085184(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085180-25085184.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085185_25085189(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085185-25085189.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085190_25085194(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085190-25085194.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085195_25085199(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085195-25085199.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085200_25085204(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085200-25085204.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085205_25085209(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085205-25085209.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085210_25085214(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085210-25085214.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085215_25085219(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085215-25085219.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085220_25085224(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085220-25085224.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085225_25085229(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085225-25085229.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085230_25085234(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085230-25085234.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085235_25085239(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085235-25085239.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085240_25085244(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085240-25085244.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085245_25085249(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085245-25085249.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085250_25085254(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085250-25085254.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085255_25085259(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085255-25085259.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085260_25085264(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085260-25085264.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085265_25085269(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085265-25085269.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085270_25085274(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085270-25085274.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085275_25085279(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085275-25085279.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085280_25085284(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085280-25085284.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085285_25085289(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085285-25085289.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085290_25085294(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085290-25085294.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085295_25085299(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085295-25085299.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085300_25085304(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085300-25085304.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085305_25085309(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085305-25085309.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085310_25085314(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085310-25085314.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085315_25085319(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085315-25085319.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085320_25085324(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085320-25085324.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085325_25085329(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085325-25085329.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085330_25085334(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085330-25085334.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085335_25085339(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085335-25085339.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085340_25085344(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085340-25085344.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085345_25085349(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085345-25085349.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085350_25085354(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085350-25085354.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085355_25085359(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085355-25085359.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085360_25085364(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085360-25085364.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085365_25085369(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085365-25085369.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085370_25085374(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085370-25085374.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085375_25085379(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085375-25085379.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085380_25085384(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085380-25085384.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085385_25085389(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085385-25085389.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085390_25085394(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085390-25085394.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085395_25085399(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085395-25085399.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085400_25085404(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085400-25085404.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085405_25085409(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085405-25085409.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085410_25085414(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085410-25085414.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085415_25085419(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085415-25085419.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085420_25085424(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085420-25085424.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085425_25085429(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085425-25085429.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085430_25085434(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085430-25085434.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085435_25085439(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085435-25085439.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085440_25085444(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085440-25085444.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085445_25085449(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085445-25085449.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085450_25085454(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085450-25085454.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085455_25085459(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085455-25085459.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085460_25085464(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085460-25085464.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085465_25085469(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085465-25085469.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085470_25085474(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085470-25085474.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085475_25085479(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085475-25085479.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085480_25085484(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085480-25085484.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085485_25085489(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085485-25085489.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085490_25085494(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085490-25085494.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085495_25085499(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085495-25085499.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085500_25085504(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085500-25085504.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085505_25085509(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085505-25085509.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085510_25085514(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085510-25085514.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085515_25085519(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085515-25085519.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085520_25085524(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085520-25085524.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085525_25085529(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085525-25085529.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085530_25085534(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085530-25085534.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085535_25085539(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085535-25085539.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085540_25085544(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085540-25085544.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085545_25085549(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085545-25085549.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085550_25085554(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085550-25085554.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085555_25085559(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085555-25085559.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085560_25085564(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085560-25085564.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085565_25085569(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085565-25085569.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085570_25085574(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085570-25085574.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085575_25085579(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085575-25085579.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085580_25085584(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085580-25085584.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085585_25085589(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085585-25085589.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085590_25085594(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085590-25085594.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085595_25085599(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085595-25085599.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085600_25085604(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085600-25085604.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085605_25085609(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085605-25085609.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085610_25085614(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085610-25085614.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085615_25085619(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085615-25085619.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085620_25085624(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085620-25085624.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085625_25085629(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085625-25085629.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085630_25085634(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085630-25085634.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085635_25085639(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085635-25085639.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085640_25085644(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085640-25085644.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085645_25085649(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085645-25085649.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085650_25085654(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085650-25085654.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085655_25085659(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085655-25085659.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085660_25085664(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085660-25085664.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085665_25085669(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085665-25085669.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085670_25085674(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085670-25085674.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085675_25085679(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085675-25085679.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085680_25085684(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085680-25085684.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085685_25085689(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085685-25085689.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085690_25085694(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085690-25085694.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085695_25085699(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085695-25085699.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085700_25085704(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085700-25085704.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085705_25085709(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085705-25085709.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085710_25085714(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085710-25085714.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085715_25085719(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085715-25085719.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085720_25085724(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085720-25085724.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085725_25085729(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085725-25085729.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085730_25085734(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085730-25085734.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085735_25085739(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085735-25085739.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085740_25085744(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085740-25085744.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085745_25085749(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085745-25085749.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085750_25085754(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085750-25085754.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085755_25085759(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085755-25085759.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085760_25085764(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085760-25085764.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085765_25085769(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085765-25085769.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085770_25085774(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085770-25085774.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085775_25085779(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085775-25085779.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085780_25085784(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085780-25085784.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085785_25085789(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085785-25085789.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085790_25085794(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085790-25085794.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085795_25085799(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085795-25085799.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085800_25085804(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085800-25085804.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085805_25085809(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085805-25085809.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085810_25085814(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085810-25085814.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085815_25085819(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085815-25085819.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085820_25085824(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085820-25085824.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085825_25085829(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085825-25085829.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085830_25085834(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085830-25085834.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085835_25085839(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085835-25085839.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085840_25085844(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085840-25085844.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085845_25085849(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085845-25085849.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085850_25085854(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085850-25085854.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085855_25085859(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085855-25085859.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085860_25085864(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085860-25085864.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085865_25085869(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085865-25085869.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085870_25085874(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085870-25085874.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085875_25085879(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085875-25085879.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085880_25085884(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085880-25085884.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085885_25085889(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085885-25085889.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085890_25085894(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085890-25085894.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085895_25085899(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085895-25085899.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085900_25085904(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085900-25085904.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085905_25085909(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085905-25085909.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085910_25085914(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085910-25085914.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085915_25085919(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085915-25085919.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085920_25085924(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085920-25085924.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085925_25085929(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085925-25085929.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085930_25085934(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085930-25085934.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085935_25085939(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085935-25085939.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085940_25085944(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085940-25085944.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085945_25085949(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085945-25085949.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085950_25085954(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085950-25085954.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085955_25085959(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085955-25085959.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085960_25085964(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085960-25085964.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085965_25085969(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085965-25085969.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085970_25085974(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085970-25085974.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085975_25085979(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085975-25085979.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085980_25085984(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085980-25085984.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085985_25085989(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085985-25085989.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085990_25085994(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085990-25085994.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25085995_25085999(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25085995-25085999.mainnet.prague.json.gz", testInfo);
  }

  @Test
  void block_25086000_25086004(TestInfo testInfo) {
    replay(MAINNET_TESTCONFIG(PRAGUE), "prague/25086000-25086004.mainnet.prague.json.gz", testInfo);
  }
}
