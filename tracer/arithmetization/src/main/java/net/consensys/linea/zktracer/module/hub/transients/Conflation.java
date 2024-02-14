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

package net.consensys.linea.zktracer.module.hub.transients;

import java.util.ArrayList;
import java.util.List;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.runtime.LogInvocation;

/** Stores data relative to the conflation. */
@Accessors(fluent = true)
@Getter
public class Conflation {
  private int number = 0;
  private DeploymentInfo deploymentInfo;
  private final List<LogInvocation> logs = new ArrayList<>(100);

  public int log(LogInvocation logInvocation) {
    this.logs.add(logInvocation);
    return this.logs.size() - 1;
  }

  public int currentLogId() {
    return this.logs.size() - 1;
  }

  public void update() {
    this.number++;
    this.deploymentInfo = new DeploymentInfo();
  }
}
