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

package net.consensys.linea.zktracer.opcode.gas;

import lombok.RequiredArgsConstructor;

// TODO we shouldn't use this, but takes value from constans/constants.lisp

/** All the classes of gas prices per instruction used in the EVM. */
@RequiredArgsConstructor
public enum GasConstants {
  G_ZERO(0),
  G_JUMP_DEST(0),
  G_BASE(2),
  G_VERY_LOW(3),
  G_LOW(5),
  G_MID(8),
  G_HIGH(10),
  G_WARM_ACCESS(100),
  G_ACCESS_LIST_ADDRESS(2400),
  G_ACCESS_LIST_STORAGE(1900),
  G_COLD_ACCOUNT_ACCESS(2600),
  G_COLD_S_LOAD(2100),
  G_S_SET(20000),
  G_S_RESET(2900),
  R_S_CLEAR(15000),
  R_SELF_DESTRUCT(24000),
  G_SELF_DESTRUCT(5000),
  G_CREATE(32000),
  G_CODE_DEPOSIT(200),
  G_CALL_VALUE(9000),
  G_CALL_STIPEND(2300),
  G_NEW_ACCOUNT(25000),
  G_EXP(10),
  G_EXP_BYTE(50),
  G_MEMORY(3),
  G_TX_CREATE(32000),
  G_TX_DATA_ZERO(4),
  G_TX_DATA_NON_ZERO(16),
  G_TRANSACTION(21000),
  G_LOG_0(Constants.LOG),
  G_LOG_1(Constants.LOG + Constants.LOG_TOPIC),
  G_LOG_2(Constants.LOG + 2 * Constants.LOG_TOPIC),
  G_LOG_3(Constants.LOG + 3 * Constants.LOG_TOPIC),
  G_LOG_4(Constants.LOG + 4 * Constants.LOG_TOPIC),
  G_LOG_DATA(8),
  G_LOG_TOPIC(375),
  G_KECCAK_256(30),
  G_KECCAK_256_WORD(6),
  G_COPY(3),
  G_BLOCK_HASH(20),
  // below are markers for gas that is computed in other modules
  // that is: hub, memory expansion, stipend, precompile info
  S_MXP(0),
  S_CALL(0), // computing the cost of a CALL requires HUB data (warmth, account existence, ...), MXP
  // data for memory expansion, STP data for gas stipend <- made it its own type
  S_HUB(0),
  S_STP(0),
  S_PREC_INFO(0);

  /** The gas price of the instruction family. */
  private final int cost;

  public int cost() {
    return this.cost;
  }

  /** Constants required to compute some instruction families base price. */
  private static class Constants {
    /** Base price for a LOGx call. */
    private static final int LOG = 375;
    /** Additional price per topic for a LOGx call. */
    private static final int LOG_TOPIC = 375;
  }
}
