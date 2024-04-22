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

package net.consensys.linea.zktracer.module.rom;

import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.romLex.RomChunk;
import net.consensys.linea.zktracer.module.romLex.RomLex;

@RequiredArgsConstructor
public class Rom implements Module {
  private final RomLex romLex;

  @Override
  public String moduleKey() {
    return "ROM";
  }

  @Override
  public void enterTransaction() {}

  @Override
  public void popTransaction() {}

  @Override
  public int lineCount() {
    return this.romLex.chunks().lineCount();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int cfi = 0;
    final int cfiInfty = this.romLex.sortedChunks().size();
    for (RomChunk chunk : this.romLex.sortedChunks()) {
      cfi += 1;
      chunk.trace(trace, cfi, cfiInfty);
    }
  }
}
