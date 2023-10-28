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

package net.consensys.linea.zktracer.module.rlp_txn;

import java.math.BigInteger;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
record TxnrlpTrace(@JsonProperty("Trace") Trace trace) {
  static final BigInteger G_txdatanonzero = new BigInteger("16");
  static final BigInteger G_txdatazero = new BigInteger("4");
  static final BigInteger LLARGE = new BigInteger("16");
  static final BigInteger LLARGEMO = new BigInteger("15");
  static final BigInteger int_long = new BigInteger("183");
  static final BigInteger int_short = new BigInteger("128");
  static final BigInteger list_long = new BigInteger("247");
  static final BigInteger list_short = new BigInteger("192");
}
