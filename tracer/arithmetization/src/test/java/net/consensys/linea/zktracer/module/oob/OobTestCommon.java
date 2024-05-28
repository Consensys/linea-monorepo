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

package net.consensys.linea.zktracer.module.oob;

class OobTestCommon {
  // TODO: all these methods are deprecated since oobEvent has been removed, update them once oob is
  // done
  // Support methods to assert that the oob events are set correctly
  static boolean getOobEvent1AtRow(final Oob oob, int i) {
    // return oob.getChunks().get(i).isOobEvent1();
    return false;
  }

  static boolean getOobEvent2AtRow(final Oob oob, int i) {
    // return oob.getChunks().get(i).isOobEvent2();
    return false;
  }

  /* Note that the methods below refer to the values of oobEvent1 and oobEvent2 in the chunks,
   * note the trace. One chunk may correspond to more than one row in the trace.
   */
  static void assertOobEvents(final Oob oob, boolean[] oobEvent1, boolean[] oobEvent2) {
    if (oobEvent1.length != oobEvent2.length) {
      throw new IllegalArgumentException("oobEvent1 and oobEvent2 do not have the same length");
    }
    assert (oobEvent1.length == oob.getChunks().size());
    for (int i = 0; i < oobEvent1.length; i++) {
      assert (getOobEvent1AtRow(oob, i) == oobEvent1[i]);
      assert (getOobEvent2AtRow(oob, i) == oobEvent2[i]);
    }
  }

  static void assertNumberOfOnesInOobEvent1(final Oob oob, int numberOfOnesInOobEvent1) {
    /*
    int actualNumberOfOnesInOobEvent1 = 0;
    for (OobOperation oobOperation : oob.getOperations()) {
      actualNumberOfOnesInOobEvent1 += oobOperation.isOobEvent1() ? 1 : 0;
    }
    assert (actualNumberOfOnesInOobEvent1 == numberOfOnesInOobEvent1);
    */
  }
}
