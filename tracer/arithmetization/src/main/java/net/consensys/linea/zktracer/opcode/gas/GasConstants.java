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
import net.consensys.linea.zktracer.module.constants.GlobalConstants;

// TODO we shouldn't use this, but takes value from constans/constants.lisp

/** All the classes of gas prices per instruction used in the EVM. */
@RequiredArgsConstructor
public enum GasConstants {
  G_ZERO(GlobalConstants.GAS_CONST_G_ZERO),
  G_JUMP_DEST(GlobalConstants.GAS_CONST_G_JUMPDEST),
  G_BASE(GlobalConstants.GAS_CONST_G_BASE),
  G_VERY_LOW(GlobalConstants.GAS_CONST_G_VERY_LOW),
  G_LOW(GlobalConstants.GAS_CONST_G_LOW),
  G_MID(GlobalConstants.GAS_CONST_G_MID),
  G_HIGH(GlobalConstants.GAS_CONST_G_HIGH),
  G_WARM_ACCESS(GlobalConstants.GAS_CONST_G_WARM_ACCESS),
  G_ACCESS_LIST_ADDRESS(GlobalConstants.GAS_CONST_G_ACCESS_LIST_ADRESS),
  G_ACCESS_LIST_STORAGE(GlobalConstants.GAS_CONST_G_ACCESS_LIST_STORAGE),
  G_COLD_ACCOUNT_ACCESS(GlobalConstants.GAS_CONST_G_COLD_ACCOUNT_ACCESS),
  G_COLD_S_LOAD(GlobalConstants.GAS_CONST_G_COLD_SLOAD),
  G_S_SET(GlobalConstants.GAS_CONST_G_SSET),
  G_S_RESET(GlobalConstants.GAS_CONST_G_SRESET),
  R_S_CLEAR(GlobalConstants.REFUND_CONST_R_SCLEAR),
  G_SELF_DESTRUCT(GlobalConstants.GAS_CONST_G_SELFDESTRUCT),
  G_CREATE(GlobalConstants.GAS_CONST_G_CREATE),
  G_CODE_DEPOSIT(GlobalConstants.GAS_CONST_G_CODE_DEPOSIT),
  G_CALL_VALUE(GlobalConstants.GAS_CONST_G_CALL_VALUE),
  G_CALL_STIPEND(GlobalConstants.GAS_CONST_G_CALL_STIPEND),
  G_NEW_ACCOUNT(GlobalConstants.GAS_CONST_G_NEW_ACCOUNT),
  G_EXP(GlobalConstants.GAS_CONST_G_EXP),
  G_EXP_BYTE(GlobalConstants.GAS_CONST_G_EXP_BYTE),
  G_MEMORY(GlobalConstants.GAS_CONST_G_MEMORY),
  G_TX_CREATE(GlobalConstants.GAS_CONST_G_TX_CREATE),
  G_TX_DATA_ZERO(GlobalConstants.GAS_CONST_G_TX_DATA_ZERO),
  G_TX_DATA_NON_ZERO(GlobalConstants.GAS_CONST_G_TX_DATA_NONZERO),
  G_TRANSACTION(GlobalConstants.GAS_CONST_G_TRANSACTION),
  G_LOG_0(GlobalConstants.GAS_CONST_G_LOG),
  G_LOG_1(GlobalConstants.GAS_CONST_G_LOG + GlobalConstants.GAS_CONST_G_LOG_TOPIC),
  G_LOG_2(GlobalConstants.GAS_CONST_G_LOG + 2 * GlobalConstants.GAS_CONST_G_LOG_TOPIC),
  G_LOG_3(GlobalConstants.GAS_CONST_G_LOG + 3 * GlobalConstants.GAS_CONST_G_LOG_TOPIC),
  G_LOG_4(GlobalConstants.GAS_CONST_G_LOG + 4 * GlobalConstants.GAS_CONST_G_LOG_TOPIC),
  G_LOG_DATA(GlobalConstants.GAS_CONST_G_LOG_DATA),
  G_LOG_TOPIC(GlobalConstants.GAS_CONST_G_LOG_TOPIC),
  G_KECCAK_256(GlobalConstants.GAS_CONST_G_KECCAK_256),
  G_KECCAK_256_WORD(GlobalConstants.GAS_CONST_G_KECCAK_256_WORD),
  G_COPY(GlobalConstants.GAS_CONST_G_COPY),
  G_BLOCK_HASH(GlobalConstants.GAS_CONST_G_BLOCKHASH),
  // below are markers for gas that is computed in other modules
  // that is: hub, memory expansion, stipend, precompile info
  S_MXP(0),
  S_CALL(0),
  // computing the cost of a CALL requires HUB data (warmth, account existence, ...), MXP
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
