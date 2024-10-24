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
package net.consensys.linea.metrics;

import java.util.Optional;

import org.hyperledger.besu.plugin.services.metrics.MetricCategory;

public enum LineaMetricCategory implements MetricCategory {
  /** Profitability metric category */
  PROFITABILITY("profitability");

  private static final Optional<String> APPLICATION_PREFIX = Optional.of("linea_");
  private final String name;

  LineaMetricCategory(final String name) {
    this.name = name;
  }

  @Override
  public String getName() {
    return name;
  }

  @Override
  public Optional<String> getApplicationPrefix() {
    return APPLICATION_PREFIX;
  }
}
