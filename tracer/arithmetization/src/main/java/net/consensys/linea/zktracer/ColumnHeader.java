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

package net.consensys.linea.zktracer;

import static com.google.common.base.Preconditions.*;

public record ColumnHeader(String name, int bytesPerElement, int length) {
  public int dataSize() {
    return this.length() * this.bytesPerElement();
  }

  public int headerSize() {
    return 2
        + // i16: name size
        this.name.length() // [u8]: name bytes
        + 1 // bytes per element
        + 4; // element count
  }

  public long cumulatedSize() {
    return this.headerSize() + this.dataSize();
  }

  public static ColumnHeader make(String name, int bytesPerElement, int length) {
    checkArgument(name.length() < Short.MAX_VALUE, "column name is too long");
    return new ColumnHeader(name, bytesPerElement, length);
  }
}
