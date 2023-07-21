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

package net.consensys.linea.zktracer.module.alu.add;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;

/**
 * Holds an execution trace and a module stamp for an addition/subtraction operation.
 *
 * @param trace a trace consisting of addition/subtraction related columns.
 * @param stamp a module stamp (counts calls to a given module).
 */
@JsonPropertyOrder({"Trace", "Stamp"})
public record AddTrace(@JsonProperty("Trace") Trace trace, @JsonProperty("Stamp") int stamp) {}
