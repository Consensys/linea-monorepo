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

import java.util.Collections;
import java.util.Map;
import java.util.Objects;
import java.util.Set;
import java.util.stream.Collectors;

import com.fasterxml.jackson.annotation.JsonIgnore;

public record ModuleToConstraints(String moduleName, Map<String, Set<String>> constraints) {

  @Override
  public boolean equals(Object o) {
    if (this == o) {
      return true;
    }
    if (o == null || moduleName.getClass() != o.getClass()) {
      return false;
    }
    return moduleName.equals(o);
  }

  @Override
  public int hashCode() {
    return Objects.hash(moduleName);
  }

  @JsonIgnore
  public Set<String> getFailedTests() {
    return constraints.values().stream().flatMap(Set::stream).collect(Collectors.toSet());
  }

  @JsonIgnore
  public Set<String> getFailedTests(String failedConstraint) {
    Set<String> failedTests = constraints.get(failedConstraint);
    if (failedTests == null) {
      return Collections.emptySet();
    }
    return failedTests;
  }
}
