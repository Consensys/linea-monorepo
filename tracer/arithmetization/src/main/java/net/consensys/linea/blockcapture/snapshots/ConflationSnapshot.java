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

package net.consensys.linea.blockcapture.snapshots;

import java.util.List;

/**
 * Contain the minimal set of information to replay a conflation as a unit test without requiring
 * access to the whole state.
 *
 * @param blocks the blocks within the conflation
 * @param accounts the accounts whose state will be read during the conflation execution
 * @param storage storage cells that will be accessed during the conflation execution
 */
public record ConflationSnapshot(
    List<BlockSnapshot> blocks, List<AccountSnapshot> accounts, List<StorageSnapshot> storage) {}
