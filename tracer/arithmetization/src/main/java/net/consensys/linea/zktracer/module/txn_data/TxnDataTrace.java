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

package net.consensys.linea.zktracer.module.txn_data;

import com.fasterxml.jackson.annotation.JsonProperty;
import net.consensys.linea.zktracer.module.ModuleTrace;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
record TxnDataTrace(@JsonProperty("Trace") Trace trace) implements ModuleTrace {
  static final int CREATE2_SHIFT = 255;
  static final int G_TXDATA_NONZERO = 16;
  static final int G_TXDATA_ZERO = 4;
  static final int G_accesslistaddress = 2400;
  static final int G_accessliststorage = 1900;
  static final int G_transaction = 21000;
  static final int G_txcreate = 32000;
  static final int INT_LONG = 183;
  static final int INT_SHORT = 128;
  static final int LIST_LONG = 247;
  static final int LIST_SHORT = 192;
  static final int LLARGE = 16;
  static final int LLARGEMO = 15;
  static final int LT = 16;
  static final int RLPADDR_CONST_RECIPE_1 = 1;
  static final int RLPADDR_CONST_RECIPE_2 = 2;
  static final int RLPRECEIPT_SUBPHASE_ID_ADDR = 53;
  static final int RLPRECEIPT_SUBPHASE_ID_CUMUL_GAS = 3;
  static final int RLPRECEIPT_SUBPHASE_ID_DATA_LIMB = 77;
  static final int RLPRECEIPT_SUBPHASE_ID_DATA_SIZE = 83;
  static final int RLPRECEIPT_SUBPHASE_ID_NO_LOG_ENTRY = 11;
  static final int RLPRECEIPT_SUBPHASE_ID_STATUS_CODE = 2;
  static final int RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE = 65;
  static final int RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA = 96;
  static final int RLPRECEIPT_SUBPHASE_ID_TYPE = 7;
  static final int common_rlp_txn_phase_number_0 = 0;
  static final int common_rlp_txn_phase_number_1 = 7;
  static final int common_rlp_txn_phase_number_2 = 2;
  static final int common_rlp_txn_phase_number_3 = 8;
  static final int common_rlp_txn_phase_number_4 = 9;
  static final int common_rlp_txn_phase_number_5 = 6;
  static final int nROWS0 = 6;
  static final int nROWS1 = 7;
  static final int nROWS2 = 7;
  static final int type_0_rlp_txn_phase_number_6 = 3;
  static final int type_1_rlp_txn_phase_number_6 = 3;
  static final int type_1_rlp_txn_phase_number_7 = 10;
  static final int type_2_rlp_txn_phase_number_6 = 5;
  static final int type_2_rlp_txn_phase_number_7 = 10;

  @Override
  public int length() {
    return this.trace.size();
  }
}
