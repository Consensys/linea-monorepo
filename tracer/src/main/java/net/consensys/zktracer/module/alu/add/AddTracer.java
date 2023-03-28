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
package net.consensys.zktracer.module.alu.add;

import net.consensys.zktracer.OpCode;
import net.consensys.zktracer.bytes.Bytes16;
import net.consensys.zktracer.bytes.UnsignedByte;
import net.consensys.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.evm.frame.MessageFrame;

import java.math.BigInteger;
import java.util.List;

public class AddTracer implements ModuleTracer {


    private int stamp = 0;
    @Override
    public String jsonKey() {
        return "add";
    }

    @Override
    public List<OpCode> supportedOpCodes() {
         return List.of(OpCode.ADD, OpCode.SUB);
    }

    @SuppressWarnings({"UnusedVariable"})
    @Override
    public Object trace(MessageFrame frame) {
        // TODO duplicated code
        final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
        final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));

        // TODO duplicated code
        final Bytes16 arg1Hi = Bytes16.wrap(arg1.slice(0, 16));
        final Bytes16 arg1Lo = Bytes16.wrap(arg1.slice(16));
        final Bytes16 arg2Hi = Bytes16.wrap(arg2.slice(0, 16));
        final Bytes16 arg2Lo = Bytes16.wrap(arg2.slice(16));

        boolean overflowHi = false;
        boolean overflowLo;

        final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
        final Res res = Res.create(opCode, arg1, arg2);

        final Bytes16 resHi = res.getResHi();
        final Bytes16 resLo = res.getResLo();

        final AddTrace.Trace.Builder builder = AddTrace.Trace.Builder.newInstance();

        stamp++;
        for (int i = 0; i < 16; i++) {

            UInt256 arg1Int = UInt256.fromBytes(arg1);
            UInt256 arg2Int = UInt256.fromBytes(arg2);
            BigInteger arg1BigInt = arg1Int.toUnsignedBigInteger();
            BigInteger arg2BigInt = arg2Int.toUnsignedBigInteger();
            BigInteger resultBigInt;

            if (opCode == OpCode.ADD) {
                resultBigInt = arg1BigInt.add(arg2BigInt);
                if (resultBigInt.compareTo(UInt256.MAX_VALUE.toBigInteger()) > 0) {
                    overflowHi = true;
                }
            } else if (opCode == OpCode.SUB) {
                if (UInt256.ZERO.toBigInteger().add(arg2BigInt).compareTo(UInt256.MAX_VALUE.toBigInteger()) > 0) {
                    overflowHi = true;
                }
            }

            // check if the result is greater than 2^128
            final BigInteger twoToThe128 = BigInteger.ONE.shiftLeft(128);
            if (opCode == OpCode.ADD) {
                BigInteger addResult = arg1Lo.toUnsignedBigInteger().add(arg2Lo.toUnsignedBigInteger());
                overflowLo = (addResult.compareTo(twoToThe128) >= 0);
            } else {
                BigInteger addResult = resLo.toUnsignedBigInteger().add(arg2Lo.toUnsignedBigInteger());
                overflowLo = (addResult.compareTo(twoToThe128) >= 0);
            }


            builder
                    .appendAcc1(resHi.slice(0, 1 + i).toUnsignedBigInteger())
                    .appendAcc2(resLo.slice(0, 1 + i).toUnsignedBigInteger())
                    .appendArg1Hi(arg1Hi.toUnsignedBigInteger())
                    .appendArg1Lo(arg1Lo.toUnsignedBigInteger())
                    .appendArg2Hi(arg2Hi.toUnsignedBigInteger())
                    .appendArg2Lo(arg2Lo.toUnsignedBigInteger());

            builder
                    .appendByte1(UnsignedByte.of(resHi.get(i)))
                    .appendByte2(UnsignedByte.of(resLo.get(i)));


            builder.appendCounter(i);

            builder
                    .appendInst(UnsignedByte.of(opCode.value));

            boolean overflow = overflowBit(i, overflowHi, overflowLo);
            builder.appendOverflow(overflow);

            // res HiLo
            builder
                    .appendResHi(resHi.toUnsignedBigInteger())
                    .appendResLo(resLo.toUnsignedBigInteger());
            builder.appendStamp(stamp);
        }
        builder.setStamp(stamp);
        return builder.build();
    }

    private boolean overflowBit(final int counter, final boolean overflowHi, final boolean overflowLo) {
        if (counter == 14) return overflowHi;
        if (counter == 15) return overflowLo;
        return false; // default bool value in go
    }
}
