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

package net.consensys.linea.zktracer.container;

/**
 * A stacked container must behave as the container it emulates, all the while being able to enter
 * nested modification contexts, that can be transparently reverted.
 */
public interface StackedContainer {
  /** Enter a new modification context. */
  void enter();

  /** Erase the modifications brought while in the latest modification context. */
  void pop();
}
