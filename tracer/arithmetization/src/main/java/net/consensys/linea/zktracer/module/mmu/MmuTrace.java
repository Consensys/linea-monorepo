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

package net.consensys.linea.zktracer.module.mmu;

import com.fasterxml.jackson.annotation.JsonProperty;
import net.consensys.linea.zktracer.module.ModuleTrace;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
record MmuTrace(@JsonProperty("Trace") Trace trace) implements ModuleTrace {
  static final int CALLDATACOPY = 55;
  static final int CALLDATALOAD = 53;
  static final int CODECOPY = 57;
  static final int EXTCODECOPY = 60;
  static final int ExceptionalRamToStack3To2FullFast = 607;
  static final int Exceptional_RamToStack_3To2Full = 627;
  static final int ExoToRam = 602;
  static final int ExoToRamSlideChunk = 616;
  static final int ExoToRamSlideOverlappingChunk = 618;
  static final int FirstFastSecondPadded = 625;
  static final int FirstPaddedSecondZero = 626;
  static final int FullExoFromTwo = 621;
  static final int FullStackToRam = 623;
  static final int KillingOne = 604;
  static final int LIMB_SIZE = 16;
  static final int LIMB_SIZE_MINUS_ONE = 15;
  static final int LsbFromStackToRAM = 624;
  static final int NA_RamToStack_1To1PaddedAndZero = 633;
  static final int NA_RamToStack_2To1FullAndZero = 631;
  static final int NA_RamToStack_2To1PaddedAndZero = 632;
  static final int NA_RamToStack_2To2Padded = 630;
  static final int NA_RamToStack_3To2Full = 628;
  static final int NA_RamToStack_3To2Padded = 629;
  static final int PaddedExoFromOne = 619;
  static final int PaddedExoFromTwo = 620;
  static final int PushOneRamToStack = 606;
  static final int PushTwoRamToStack = 605;
  static final int PushTwoStackToRam = 608;
  static final int RETURNDATACOPY = 62;
  static final int RamIsExo = 603;
  static final int RamLimbExcision = 613;
  static final int RamToRam = 601;
  static final int RamToRamSlideChunk = 614;
  static final int RamToRamSlideOverlappingChunk = 615;
  static final int SMALL_LIMB_SIZE = 4;
  static final int SMALL_LIMB_SIZE_MINUS_ONE = 3;
  static final int StoreXInAThreeRequired = 609;
  static final int StoreXInB = 610;
  static final int StoreXInC = 611;
  static final int tern0 = 0;
  static final int tern1 = 1;
  static final int tern2 = 2;
  static final int type1 = 100;
  static final int type2 = 200;
  static final int type3 = 300;
  static final int type4CC = 401;
  static final int type4CD = 402;
  static final int type4RD = 403;
  static final int type5 = 500;

  @Override
  public int length() {
    return this.trace.size();
  }
}
