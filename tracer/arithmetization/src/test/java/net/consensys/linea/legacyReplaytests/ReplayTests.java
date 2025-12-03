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
import static net.consensys.linea.zktracer.ChainConfig.*;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;

@Tag("replay")
@Disabled
@ExtendWith(UnitTestWatcher.class)
public class ReplayTests extends TracerTestBase {

  @Test
  void fatMxp(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/2492975-2492977.mainnet.json.gz", testInfo);
  }

  /**
   * bulk-replay of multiple replay files specified by a directory. The conflated traces will be
   * moved to "conflated" directory once replayed. The replay files will be moved to "replayed"
   * directory once completed. Note: CORSET_VALIDATOR.validate() is disabled by default for
   * bulkReplay. Usage: bulkReplay("/path/to/your/directory", testInfo);
   */
  @Test
  void bulkReplay(TestInfo testInfo) {
    // bulkReplay("./src/test/resources/replays", testInfo);
    // bulkReplay(OLD_LINEA_MAINNET, "", testInfo);
  }

  @Test
  @Disabled
  void failingMmuModexp(TestInfo testInfo) {
    // OOB failure
    replay(OLD_MAINNET_TESTCONFIG, "legacy/5995162.mainnet.json.gz", testInfo);
  }

  @Test
  void failRlpAddress(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/5995097.mainnet.json.gz", testInfo);
  }

  @Test
  void rlprcptManyTopicsWoLogData(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/6569423.mainnet.json.gz", testInfo);
  }

  @Test
  void multipleFailingCallToEcrecover(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/5000544.mainnet.json.gz", testInfo);
  }

  @Test
  @Tag("nightly")
  void incident777zkGethMainnet(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/7461019-7461030.mainnet.json.gz", testInfo);
  }

  @Test
  void issue1006(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/6032696-6032699.mainnet.json.gz", testInfo);
  }

  @Test
  void issue1004(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/6020023-6020029.mainnet.json.gz", testInfo);
  }

  @Test
  void block_6110045(TestInfo testInfo) {
    // The purpose of this test is to check the mechanism for spotting divergence between the replay
    // tests and mainnet.  Specifically, this replay has transaction result information embedded
    // within it.
    replay(OLD_MAINNET_TESTCONFIG, "legacy/6110045.mainnet.json.gz", testInfo);
  }

  @Test
  void failingCreate2(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/2250197.mainnet.json.gz", testInfo);
  }

  @Disabled("Fails to create the ConflationSnapshot from the gson file")
  @Test
  void blockHash1(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/8718090.mainnet.json.gz", testInfo);
  }

  @Disabled("Fails to create the ConflationSnapshot from the gson file")
  @Test
  void blockHash2(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/8718330.mainnet.json.gz", testInfo);
  }

  // TODO: should be replaced by a unit test triggering AnyToRamWithPadding (mixed case) MMU
  // instruction
  @Test
  void negativeNumberOfMmioInstruction(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/6029454-6029459.mainnet.json.gz", testInfo);
  }

  @Test
  void simpleSelfDestruct(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/50020-50029.mainnet.json.gz", testInfo);
  }

  // TODO: should be replaced by a unit test triggering a failed CREATE2
  @Test
  void failedCreate2(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/41640-41649.mainnet.json.gz", testInfo);
  }

  @Test
  void largeInitCode(TestInfo testInfo) {
    replay(OLD_SEPOLIA_TESTCONFIG, "legacy/3318494.sepolia.json.gz", testInfo);
  }

  /**
   * TODO: should be replace by a unit test triggering a STATICCALL to a precompile, without enough
   * remaining gas if the precompile was considered COLD
   */
  @Test
  void hotOrColdPrecompile(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/2019510-2019519.mainnet.json.gz", testInfo);
  }

  // TODO: should be replace by a unit test triggering a CALLDATACOPY in a ROOT context of a
  // deployment transaction
  @Test
  void callDataCopyCnNotFound(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/67050-67059.mainnet.json.gz", testInfo);
  }

  /**
   * TODO: should be replace by a unit test triggering a RETURN during a deployment transaction,
   * where we run OOG when need to pay the gas cost of the code deposit
   */
  @Test
  void returnOogxForCodeDepositCost(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/1002387.mainnet.json.gz", testInfo);
  }

  @Test
  @Tag("nightly")
  void modexpTriggeringNonAlignedFirstLimbSingleSourceMmuModexp(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/3108622-3108633.mainnet.json.gz", testInfo);
  }

  /**
   * Not sure if we need to keep this replayTest. We were using a source offset instead of the dest
   * Offset to compute the memory expansion cost, thus creating a fake OOGX
   */
  @Test
  void mainnet1339346ContextRevertTwice(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/1339346.mainnet.json.gz", testInfo);
  }

  @Test
  void legacyTxWithoutChainID(TestInfo testInfo) {
    replay(OLD_SEPOLIA_TESTCONFIG, "legacy/254251.sepolia.json.gz", testInfo);
  }

  @Test
  void incorrectCreationCapture(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/4323985.mainnet.json.gz", testInfo);
  }

  @Disabled("Fails to create the ConflationSnapshot from the gson file")
  @Test
  void duplicateSubZero(TestInfo testInfo) {
    replay(MAINNET_LONDON_TESTCONFIG, "legacy/20197061-20197173.mainnet.json.gz", testInfo);
  }

  /**
   * The mainnet conflation below blew up the HUB <> state manager integration. The issue was
   * related to a <b>SSTOREX</b> exception happening in the wild. Recall that this is the exception
   * triggered when a <b>SSTORE</b> instruction is executed and the frame contains <b>≤ 2_300</b>
   * gas.
   */
  @Test
  void stateManagerIntegrationTest(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/SSTOREX_on_mainnet.json.gz", testInfo);
  }
}
