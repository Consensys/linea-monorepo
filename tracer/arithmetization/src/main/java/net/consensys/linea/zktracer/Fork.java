/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer;

public enum Fork {
  LONDON,
  PARIS,
  SHANGHAI,
  CANCUN,
  PRAGUE;

  public static String toString(Fork fork) {
    return switch (fork) {
      case LONDON -> "london";
      case PARIS -> "paris";
      case SHANGHAI -> "shanghai";
      case CANCUN -> "cancun";
      case PRAGUE -> "prague";
    };
  }

  /**
   * Construct a fork instance from the name of a fork (e.g. "London", "Shanghai", etc). Observe
   * that case does not matter here. Hence, "LONDON", "London", "london", "lonDon" are all suitable
   * aliases for the LONDON instance.
   *
   * @param fork
   * @return
   */
  public static Fork fromString(String fork) {
    return Fork.valueOf(fork.toUpperCase());
  }

  public static boolean isPostShanghai(Fork fork) {
    return fork.compareTo(SHANGHAI) >= 0;
  }
}
