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

package net.consensys.linea.zktracer.module.mmu;

import java.util.List;
import java.util.Map;

import lombok.Getter;
import lombok.experimental.Accessors;

@Accessors(fluent = true)
class PreComputations {
  @Getter private final Type1PreComputation type1;
  @Getter private final Type2PreComputation type2;
  @Getter private final Type3PreComputation type3;
  @Getter private final Type4PreComputation type4;
  @Getter private final Type5PreComputation type5;
  @Getter private final Map<Integer, MmuPreComputation> typeMap;
  @Getter private final List<MmuPreComputation> types;

  PreComputations() {
    this.type1 = new Type1PreComputation();
    this.type2 = new Type2PreComputation();
    this.type3 = new Type3PreComputation();
    this.type4 = new Type4PreComputation();
    this.type5 = new Type5PreComputation();
    this.types = List.of(type1, type2, type3, type4, type5);
    this.typeMap =
        Map.of(
            MmuTrace.type1, type1,
            MmuTrace.type2, type2,
            MmuTrace.type3, type3,
            MmuTrace.type4CC, type4,
            MmuTrace.type4CD, type4,
            MmuTrace.type4RD, type4,
            MmuTrace.type5, type5);
  }
}
