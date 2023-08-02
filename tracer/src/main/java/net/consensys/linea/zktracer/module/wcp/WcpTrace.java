/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.module.wcp;

import java.math.BigInteger;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
record WcpTrace(@JsonProperty("Trace") Trace trace) {
  static final BigInteger EQ_ = new BigInteger("20");
  static final BigInteger GT = new BigInteger("17");
  static final BigInteger ISZERO = new BigInteger("21");
  static final BigInteger LIMB_SIZE = new BigInteger("16");
  static final BigInteger LIMB_SIZE_MINUS_ONE = new BigInteger("15");
  static final BigInteger LT = new BigInteger("16");
  static final BigInteger SGT = new BigInteger("19");
  static final BigInteger SLT = new BigInteger("18");
}
