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

import lombok.Setter;
import net.consensys.linea.zktracer.module.ModuleName;

public class IncrementAndDetectModule extends IncrementingModule {

  public static final String ERROR_MESSAGE_TRIED_TO_COMMIT_UNPROVABLE_TX =
      "Shouldn't commit transaction as an unprovable event has been detected.";

  @Setter boolean eventDetected = false;

  public IncrementAndDetectModule(ModuleName moduleKey) {
    super(moduleKey);
  }

  @Override
  public void commitTransactionBundle() {
    checkState(!eventDetected, ERROR_MESSAGE_TRIED_TO_COMMIT_UNPROVABLE_TX);
    super.commitTransactionBundle();
  }

  @Override
  public void popTransactionBundle() {
    eventDetected = false;
    super.popTransactionBundle();
  }

  @Override
  public int lineCount() {
    return eventDetected ? Integer.MAX_VALUE : counts.lineCount();
  }

  public void detectEvent() {
    eventDetected = true;
  }
}
