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

package net.consensys.linea.zktracer.module.mul;

import java.math.BigInteger;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
record MulTrace(@JsonProperty("Trace") Trace trace) {
  static final BigInteger EXP = new BigInteger("10");
  static final BigInteger MMEDIUM = new BigInteger("8");
  static final BigInteger MMEDIUMMO = new BigInteger("7");
  static final BigInteger MUL = new BigInteger("2");
  static final BigInteger ONETWOEIGHT = new BigInteger("128");
  static final BigInteger ONETWOSEVEN = new BigInteger("127");
  static final BigInteger THETA = new BigInteger("18446744073709551616");
  static final BigInteger THETA2 = new BigInteger("340282366920938463463374607431768211456");
}
