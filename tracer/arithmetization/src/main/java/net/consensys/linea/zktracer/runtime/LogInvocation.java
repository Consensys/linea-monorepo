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

package net.consensys.linea.zktracer.runtime;

import java.util.ArrayList;
import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import org.apache.tuweni.bytes.Bytes;

@RequiredArgsConstructor
public class LogInvocation {
  private final CallStack callStack;
  private final Bytes payload;
  private final List<Bytes> topics;
  private final int callFrameId;

  public static int forOpcode(final Hub hub) {
    final List<Bytes> topics = new ArrayList<>(4);
    final Bytes payload = hub.transients().op().logData();
    switch (hub.opCode()) {
      case LOG0 -> {}
      case LOG1 -> topics.add(hub.messageFrame().getStackItem(2));
      case LOG2 -> {
        topics.add(hub.messageFrame().getStackItem(2));
        topics.add(hub.messageFrame().getStackItem(3));
      }
      case LOG3 -> {
        topics.add(hub.messageFrame().getStackItem(2));
        topics.add(hub.messageFrame().getStackItem(3));
        topics.add(hub.messageFrame().getStackItem(4));
      }
      case LOG4 -> {
        topics.add(hub.messageFrame().getStackItem(2));
        topics.add(hub.messageFrame().getStackItem(3));
        topics.add(hub.messageFrame().getStackItem(4));
        topics.add(hub.messageFrame().getStackItem(5));
      }
      default -> throw new IllegalStateException("not a LOG operation");
    }

    return hub.transients()
        .conflation()
        .log(new LogInvocation(hub.callStack(), payload, topics, hub.currentFrame().id()));
  }

  public boolean reverted() {
    return this.callStack.getById(this.callFrameId).hasReverted();
  }
}
