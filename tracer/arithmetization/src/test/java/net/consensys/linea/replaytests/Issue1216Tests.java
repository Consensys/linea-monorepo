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
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(UnitTestWatcher.class)
public class Issue1216Tests {

  /**
   * When constructing the "before" {@link AccountSnapshot} of the recipient account (i.e. the heir
   * of the SELFDESTRUCT) we are using either
   *
   * <p>(1) {@link AccountSnapshot#canonical(Hub, Address)} of the (trimmed) recipient address or
   *
   * <p>(2) {@link AccountSnapshot#deepCopy()} of the "after" version of the self destructor.
   *
   * <p>The first method the right approach when the (trimmed) recipient address is different from
   * the address of the account undergoing SELFDESTRUCT. The second method has to be used when the
   * two coincide.
   *
   * <p>The issue in the SEPOLIA block 2392659 was that {@link AccountSnapshot#deepCopy()} was
   * blowing up with a NPE since the "after" version of the self destructor didn't exist yet in the
   * code. This was solved in issue #1216.
   */
  @Tag("nightly")
  @Tag("replay")
  @Test
  void issue_1216_sepolia_block_2392659() {
    replay(LINEA_SEPOLIA, "2392659.sepolia.json.gz");
  }
}
