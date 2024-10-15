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
package net.consensys.linea;

import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.ConcurrentSkipListSet;

import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import lombok.AllArgsConstructor;
import lombok.Getter;

@Getter
@AllArgsConstructor
@JsonPropertyOrder({
  "failedCounter",
  "successCounter",
  "disabledCounter",
  "abortedCounter",
  "modules"
})
public class BlockchainReferenceTestOutcome {
  private final int failedCounter;
  private final int successCounter;
  private final int disabledCounter;
  private final int abortedCounter;
  private final ConcurrentMap<String, ConcurrentMap<String, ConcurrentSkipListSet<String>>>
      modulesToConstraintsToTests;
}
