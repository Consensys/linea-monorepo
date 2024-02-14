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

package net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob;

import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.module.oob.OobDataChannel;
import org.apache.tuweni.bytes.Bytes;

/** This interface defines the API required to execute a call to the OOB module. */
public interface OobCall extends TraceSubFragment {
  /**
   * Given a data channel number, returns the data that should be sent to the OOB through this
   * channel.
   *
   * @param i the channel number
   * @return the data to send to the OOB through the channel DATA_i
   */
  Bytes data(OobDataChannel i);

  /** The instruction to trigger in the OOB for this call. */
  int oobInstruction();

  @Override
  default Trace trace(Trace trace) {
    return trace
        .pMiscellaneousOobFlag(true)
        .pMiscellaneousOobData1(this.data(OobDataChannel.of(0)))
        .pMiscellaneousOobData2(this.data(OobDataChannel.of(1)))
        .pMiscellaneousOobData3(this.data(OobDataChannel.of(2)))
        .pMiscellaneousOobData4(this.data(OobDataChannel.of(3)))
        .pMiscellaneousOobData5(this.data(OobDataChannel.of(4)))
        .pMiscellaneousOobData6(this.data(OobDataChannel.of(5)))
        .pMiscellaneousOobData7(this.data(OobDataChannel.of(6)))
        .pMiscellaneousOobData8(this.data(OobDataChannel.of(7)))
        .pMiscellaneousOobInst(this.oobInstruction());
  }
}
