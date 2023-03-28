package net.consensys.zktracer.module.alu.add;
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
import net.consensys.zktracer.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.math.BigInteger;

public class Adder {
  private static final Logger LOG = LoggerFactory.getLogger(Adder.class);

  public static Bytes32 addSub(final OpCode opCode, final Bytes32 arg1, final Bytes32 arg2) {
    LOG.info("adding " + arg1 + " " + opCode.name() + " " + arg2);
    final BigInteger res = x(opCode, arg1, arg2);
    // ensure result is correct length
    final Bytes resBytes = Bytes.of(res.toByteArray());
    if (resBytes.size() > 32 ) {
      return Bytes32.wrap(resBytes, resBytes.size() - 32);
    }
    return Bytes32.leftPad(Bytes.of(res.toByteArray()));
  }

  private static BigInteger x(final OpCode opCode, final Bytes32 value, final Bytes32 value2) {
    {
      return switch (opCode) {
        case ADD -> value.toUnsignedBigInteger().add(value2.toUnsignedBigInteger());
        case SUB -> value.toUnsignedBigInteger().subtract(value2.toUnsignedBigInteger());
        default -> BigInteger.ZERO; // TODO what should happen here
      };
    }
  }

}
