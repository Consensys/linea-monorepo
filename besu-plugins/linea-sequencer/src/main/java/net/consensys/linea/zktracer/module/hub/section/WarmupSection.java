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

package net.consensys.linea.zktracer.module.hub.section;

import java.util.List;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.StorageFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;

/**
 * A warmup section is generated if a transaction features pre-warmed addresses and/or keys. It
 * contains a succession of {@link AccountFragment } and {@link StorageFragment} representing the
 * pre-warmed addresses and eventual keys.
 */
public class WarmupSection extends TraceSection {
  public WarmupSection(Hub hub, List<TraceFragment> fragments) {
    this.addChunksWithoutStack(hub, fragments.toArray(new TraceFragment[0]));
  }
}
