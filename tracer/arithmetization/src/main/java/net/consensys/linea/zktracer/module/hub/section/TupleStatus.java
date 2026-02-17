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

package net.consensys.linea.zktracer.module.hub.section;

public enum TupleStatus {
  VALID,
  CHAIN_ID_IS_NEITHER_EQUAL_TO_ZERO_NOR_NETWORK_CHAIN_ID,
  NONCE_IS_GREATER_THAN_MAX_NONCE,
  S_IS_GREATER_THAN_HALF_CURVE_ORDER,
  EC_RECOVER_FAILS,
  AUTHORITY_ACCOUNT_CODE_NOT_EMPTY_AND_NOT_DELEGATED,
  AUTHORITY_NONCE_IS_NOT_EQUAL_TO_NONCE,
  UNDEFINED;

  boolean isInvalidDueToPreliminaryChecks() {
    return this == CHAIN_ID_IS_NEITHER_EQUAL_TO_ZERO_NOR_NETWORK_CHAIN_ID
        || this == NONCE_IS_GREATER_THAN_MAX_NONCE
        || this == S_IS_GREATER_THAN_HALF_CURVE_ORDER;
  }

  boolean isInvalid() {
    return this != VALID && this != UNDEFINED;
  }

  boolean isUndefined() {
    return this == UNDEFINED;
  }
}
