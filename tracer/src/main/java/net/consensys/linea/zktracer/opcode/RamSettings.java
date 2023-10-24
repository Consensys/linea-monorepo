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

package net.consensys.linea.zktracer.opcode;

import java.util.Objects;

public record RamSettings(DataLocation source, DataLocation target) {
  public static final RamSettings DEFAULT = new RamSettings(DataLocation.NONE, DataLocation.NONE);

  public RamSettings(DataLocation source, DataLocation target) {
    this.source = Objects.requireNonNullElse(source, DataLocation.NONE);
    this.target = Objects.requireNonNullElse(target, DataLocation.NONE);
  }

  public boolean enabled() {
    return this.source != DataLocation.NONE || this.target != DataLocation.NONE;
  }
}
