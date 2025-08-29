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

package net.consensys.linea.zktracer.container.module;

import static com.google.common.base.Preconditions.checkState;

import java.util.List;

import lombok.Setter;
import net.consensys.linea.zktracer.Trace;

public abstract class EventDetectorModule implements Module {

  final String moduleKey;

  @Setter boolean eventDetected = false;

  protected EventDetectorModule(String moduleKey) {
    this.moduleKey = moduleKey;
  }

  @Override
  public void commitTransactionBundle() {
    checkState(
        !eventDetected, "Shouldn't commit transaction as an unprovable event has been detected.");
  }

  @Override
  public void popTransactionBundle() {
    eventDetected = false;
  }

  @Override
  public int lineCount() {
    return eventDetected ? Integer.MAX_VALUE : 0;
  }

  public void detectEvent() {
    eventDetected = true;
  }

  @Override
  public List<Trace.ColumnHeader> columnHeaders(Trace trace) {
    throw new IllegalStateException("should never be called");
  }

  @Override
  public void commit(Trace trace) {
    throw new IllegalStateException("should never be called");
  }

  @Override
  public int spillage(Trace trace) {
    return 0;
  }

  @Override
  public String moduleKey() {
    return moduleKey;
  }
}
