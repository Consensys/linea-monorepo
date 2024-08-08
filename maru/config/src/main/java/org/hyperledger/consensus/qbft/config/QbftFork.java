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
package org.hyperledger.consensus.qbft.config;

import java.lang.reflect.Field;
import java.util.Arrays;

import com.fasterxml.jackson.databind.node.ObjectNode;

/** The Qbft fork. */
public class QbftFork extends org.hyperledger.besu.config.QbftFork {
  /**
   * Instantiates a new Qbft fork.
   *
   * @param forkConfigRoot the fork config root
   */
  public QbftFork(ObjectNode forkConfigRoot) {
    super(forkConfigRoot);
  }

  // TODO: Purge it with fire
  public static QbftFork fromBesuQbftFork(org.hyperledger.besu.config.QbftFork other) {
    try {
      Field forkConfigRootField =
          Arrays.stream(other.getClass().getSuperclass().getDeclaredFields())
              .filter(field -> field.getName().equals("forkConfigRoot"))
              .findFirst()
              .get();
      forkConfigRootField.setAccessible(true);
      ObjectNode forkConfigRoot = (ObjectNode) forkConfigRootField.get(other);
      return new QbftFork(forkConfigRoot);
    } catch (IllegalAccessException e) {
      throw new RuntimeException(e);
    }
  }
}
