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
package net.consensys.linea.legacyReplaytests;

import static net.consensys.linea.ReplayTestTools.replay;
import static net.consensys.linea.zktracer.ChainConfig.OLD_MAINNET_TESTCONFIG;

import net.consensys.linea.UnitTestWatcher;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;

@Disabled
@ExtendWith(UnitTestWatcher.class)
public class UpdatedShomeyTests {

  @Test
  void split_range_19239480_19239481(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19239480-19239481.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19239502_19239502(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19239502-19239502.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19239569_19239573(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19239569-19239573.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19240615_19240637(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19240615-19240637.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19241185_19241187(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19241185-19241187.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19241495_19241523(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19241495-19241523.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19242040_19242058(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19242040-19242058.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19242738_19242740(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19242738-19242740.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19247347_19247355(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19247347-19247355.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19247389_19247391(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19247389-19247391.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19248865_19248877(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19248865-19248877.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19248878_19248882(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19248878-19248882.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19249100_19249128(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19249100-19249128.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19250714_19250739(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19250714-19250739.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19251076_19251097(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19251076-19251097.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19251417_19251433(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19251417-19251433.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19251688_19251695(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19251688-19251695.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19252291_19252328(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19252291-19252328.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19252392_19252424(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19252392-19252424.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19253122_19253160(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19253122-19253160.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19255567_19255570(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19255567-19255570.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19255638_19255648(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19255638-19255648.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19260916_19260920(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19260916-19260920.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19261984_19261986(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19261984-19261986.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19263140_19263143(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19263140-19263143.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19263589_19263617(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19263589-19263617.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19263697_19263706(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19263697-19263706.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19263746_19263759(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19263746-19263759.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19269273_19269283(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19269273-19269283.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19276590_19276593(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19276590-19276593.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19277153_19277155(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19277153-19277155.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19277205_19277209(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19277205-19277209.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19277210_19277210(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19277210-19277210.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19277227_19277227(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19277227-19277227.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19277362_19277363(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19277362-19277363.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19277661_19277662(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19277661-19277662.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19278085_19278095(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19278085-19278095.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19278117_19278139(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19278117-19278139.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19278361_19278377(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19278361-19278377.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19278457_19278460(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19278457-19278460.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19279568_19279581(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19279568-19279581.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19280065_19280068(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19280065-19280068.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19280843_19280861(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19280843-19280861.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19281442_19281478(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19281442-19281478.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19282860_19282878(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19282860-19282878.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19283417_19283451(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19283417-19283451.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19285053_19285075(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19285053-19285075.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19286476_19286504(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19286476-19286504.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19286910_19286917(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19286910-19286917.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19288775_19288813(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19288775-19288813.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19288947_19288980(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19288947-19288980.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19289613_19289614(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19289613-19289614.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19289615_19289616(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19289615-19289616.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19290441_19290479(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19290441-19290479.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19291103_19291136(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19291103-19291136.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19291773_19291811(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19291773-19291811.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19292980_19292983(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19292980-19292983.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19293000_19293004(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19293000-19293004.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19293359_19293395(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19293359-19293395.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19293534_19293548(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19293534-19293548.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19293971_19293980(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19293971-19293980.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19295632_19295635(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19295632-19295635.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19296085_19296117(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19296085-19296117.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19304297_19304313(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19304297-19304313.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19307602_19307640(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19307602-19307640.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19308734_19308766(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19308734-19308766.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19308851_19308874(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19308851-19308874.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19321451_19321489(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19321451-19321489.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19325021_19325059(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19325021-19325059.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19325718_19325721(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19325718-19325721.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19326229_19326250(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19326229-19326250.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19327194_19327209(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19327194-19327209.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19327846_19327876(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19327846-19327876.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19329972_19330002(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19329972-19330002.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19336921_19336941(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19336921-19336941.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19337400_19337419(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19337400-19337419.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19337440_19337460(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19337440-19337460.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19344473_19344507(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19344473-19344507.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19347697_19347735(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19347697-19347735.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19353891_19353891(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19353891-19353891.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19357441_19357479(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19357441-19357479.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19359596_19359610(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19359596-19359610.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19366075_19366077(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19366075-19366077.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19366078_19366079(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19366078-19366079.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19366461_19366463(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19366461-19366463.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19369046_19369075(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19369046-19369075.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19372571_19372577(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19372571-19372577.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19373666_19373700(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19373666-19373700.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19374039_19374077(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19374039-19374077.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19376303_19376332(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19376303-19376332.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19377828_19377866(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19377828-19377866.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19378155_19378193(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19378155-19378193.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19379192_19379193(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19379192-19379193.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19380366_19380387(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19380366-19380387.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19383544_19383563(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19383544-19383563.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19385106_19385144(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19385106-19385144.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19386801_19386832(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19386801-19386832.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19387636_19387636(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19387636-19387636.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19390295_19390333(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19390295-19390333.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19390851_19390860(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19390851-19390860.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19397327_19397330(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19397327-19397330.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19397361_19397365(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19397361-19397365.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19397483_19397488(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19397483-19397488.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19397965_19398001(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19397965-19398001.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19398290_19398293(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19398290-19398293.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19399253_19399257(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19399253-19399257.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19400123_19400133(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19400123-19400133.mainnet.json.gz", testInfo);
  }

  @Test
  void split_range_19400178_19400184(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/19400178-19400184.mainnet.json.gz", testInfo);
  }
}
