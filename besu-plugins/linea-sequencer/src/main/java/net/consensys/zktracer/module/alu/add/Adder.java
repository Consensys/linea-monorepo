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

  public static Bytes32 addSub(final OpCode opCode, final Bytes32 value, final Bytes32 value2) {
    LOG.info("adding " + value + " " + opCode.name() + " " + value2);
    return switch (opCode) {
      case ADD -> Bytes32.leftPad(Bytes.of(value.toBigInteger().add(value2.toBigInteger()).toByteArray()));
      case SUB -> Bytes32.leftPad(Bytes.of(value.toBigInteger().subtract(value2.toBigInteger()).toByteArray()));
      default -> Bytes32.ZERO; // TODO what should happen here
    };
  }

}
