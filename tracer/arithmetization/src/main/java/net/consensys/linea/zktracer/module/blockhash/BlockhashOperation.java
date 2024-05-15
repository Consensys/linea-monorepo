/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.module.blockhash;

import static net.consensys.linea.zktracer.module.blockhash.Trace.LLARGE;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes32;

@Accessors(fluent = true)
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
@RequiredArgsConstructor
public class BlockhashOperation extends ModuleOperation {
  @Getter @EqualsAndHashCode.Include private final short relativeBlock;
  @Getter @EqualsAndHashCode.Include private final Bytes32 opcodeArgument;
  private final long absoluteBlockNumber;
  private final boolean lowerBound;
  private final boolean upperBound;
  @Getter private final Bytes32 result;

  @Override
  protected int computeLineCount() {
    return 1;
  }

  public void trace(Trace trace, final Bytes32 hash) {
    trace
        .iomf(true)
        .blockNumberHi(opcodeArgument.slice(0, LLARGE))
        .blockNumberLo(opcodeArgument.slice(LLARGE, LLARGE))
        .resHi(result.slice(0, LLARGE))
        .resLo(result.slice(LLARGE, LLARGE))
        .relBlock(relativeBlock)
        .absBlock(absoluteBlockNumber)
        .lowerBoundCheck(lowerBound)
        .upperBoundCheck(upperBound)
        .inRange(lowerBound && upperBound)
        .blockHashHi(hash.slice(0, LLARGE))
        .blockHashLo(hash.slice(LLARGE, LLARGE))
        .byteHi0(UnsignedByte.of(hash.get(0)))
        .byteHi1(UnsignedByte.of(hash.get(1)))
        .byteHi2(UnsignedByte.of(hash.get(2)))
        .byteHi3(UnsignedByte.of(hash.get(3)))
        .byteHi4(UnsignedByte.of(hash.get(4)))
        .byteHi5(UnsignedByte.of(hash.get(5)))
        .byteHi6(UnsignedByte.of(hash.get(6)))
        .byteHi7(UnsignedByte.of(hash.get(7)))
        .byteHi8(UnsignedByte.of(hash.get(8)))
        .byteHi9(UnsignedByte.of(hash.get(9)))
        .byteHi10(UnsignedByte.of(hash.get(10)))
        .byteHi11(UnsignedByte.of(hash.get(11)))
        .byteHi12(UnsignedByte.of(hash.get(12)))
        .byteHi13(UnsignedByte.of(hash.get(13)))
        .byteHi14(UnsignedByte.of(hash.get(14)))
        .byteHi15(UnsignedByte.of(hash.get(15)))
        .byteLo0(UnsignedByte.of(hash.get(LLARGE + 0)))
        .byteLo1(UnsignedByte.of(hash.get(LLARGE + 1)))
        .byteLo2(UnsignedByte.of(hash.get(LLARGE + 2)))
        .byteLo3(UnsignedByte.of(hash.get(LLARGE + 3)))
        .byteLo4(UnsignedByte.of(hash.get(LLARGE + 4)))
        .byteLo5(UnsignedByte.of(hash.get(LLARGE + 5)))
        .byteLo6(UnsignedByte.of(hash.get(LLARGE + 6)))
        .byteLo7(UnsignedByte.of(hash.get(LLARGE + 7)))
        .byteLo8(UnsignedByte.of(hash.get(LLARGE + 8)))
        .byteLo9(UnsignedByte.of(hash.get(LLARGE + 9)))
        .byteLo10(UnsignedByte.of(hash.get(LLARGE + 10)))
        .byteLo11(UnsignedByte.of(hash.get(LLARGE + 11)))
        .byteLo12(UnsignedByte.of(hash.get(LLARGE + 12)))
        .byteLo13(UnsignedByte.of(hash.get(LLARGE + 13)))
        .byteLo14(UnsignedByte.of(hash.get(LLARGE + 14)))
        .byteLo15(UnsignedByte.of(hash.get(LLARGE + 15)))
        .validateRow();
  }
}
