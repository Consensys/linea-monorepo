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

package net.consensys.linea.zktracer.module.mul;

import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

import java.math.BigInteger;

import net.consensys.linea.zktracer.bytes.Bytes16;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import net.consensys.linea.zktracer.module.Util;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

public class MulUtilsTest {
  @BeforeAll
  static void beforeAll() {
    OpCodes.load();
  }

  @Test
  void isTiny() {
    // tiny means zero or one
    assertThat(MulData.isTiny(BigInteger.ZERO)).isTrue();
    assertThat(MulData.isTiny(BigInteger.ONE)).isTrue();
    assertThat(MulData.isTiny(BigInteger.TWO)).isFalse();
    assertThat(MulData.isTiny(BigInteger.TEN)).isFalse();
  }

  @Test
  void twoAdicity() {
    assertThat(MulData.twoAdicity(UInt256.MIN_VALUE)).isEqualTo(256);
    // TODO no idea what these should be
    //    assertThat(MulData.twoAdicity(UInt256.MAX_VALUE)).isEqualTo(0);
    //    assertThat(MulData.twoAdicity(UInt256.valueOf(1))).isEqualTo(0);
  }

  @Test
  void multiplyByZero() {
    Bytes32 arg1 = Bytes32.random();
    OpCodeData mul = OpCode.MUL.getData();
    MulData oxo = new MulData(mul, arg1, Bytes32.ZERO);

    assertThat(oxo.getArg2Hi().isZero()).isTrue();
    assertThat(oxo.getArg2Lo()).isEqualTo(Bytes16.ZERO);
    assertThat(oxo.getArg2Hi()).isEqualTo(Bytes16.ZERO);
    assertThat(oxo.getOpCode().getData()).isEqualTo(mul);
    assertThat(oxo.isTinyExponent()).isTrue();
    assertThat(oxo.isOneLineInstruction()).isTrue();
    assertThat(oxo.bits[0]).isFalse();
  }

  @Test
  void zeroExp() {
    Bytes32 arg1 = Bytes32.random();
    OpCodeData exp = OpCode.EXP.getData();
    MulData oxo = new MulData(exp, arg1, Bytes32.ZERO);

    assertThat(oxo.getArg2Hi().isZero()).isTrue();
    assertThat(oxo.getArg2Lo()).isEqualTo(Bytes16.ZERO);
    assertThat(oxo.getArg2Hi()).isEqualTo(Bytes16.ZERO);
    assertThat(oxo.getOpCode().getData()).isEqualTo(exp);
    assertThat(oxo.isTinyExponent()).isTrue();
    assertThat(oxo.isOneLineInstruction()).isTrue();
    assertThat(oxo.bits[0]).isFalse();
  }

  @Test
  void testByteBits_ofZero() {
    Boolean[] booleans = Util.byteBits(UnsignedByte.of(0));

    assertThat(booleans.length).isEqualTo(8);
    assertThat(booleans[0]).isNotNull();
  }

  @Test
  void hBytesAllZeros() {
    Bytes32 arg1 = Bytes32.ZERO;
    Bytes32 arg2 = Bytes32.ZERO;
    MulData mulData = new MulData(OpCode.EXP, arg1, arg2);
    mulData.setHsAndBitsFromBaseThetas(
        BaseTheta.fromBytes32(UInt256.ZERO), BaseTheta.fromBytes32(UInt256.ZERO));

    assertThat(mulData.hBytes.get(0).isZero()).isTrue();
    assertThat(mulData.hBytes.get(1).isZero()).isTrue();
    assertThat(mulData.hBytes.get(2).isZero()).isTrue();
    assertThat(mulData.hBytes.get(3).isZero()).isTrue();

    assertThat(mulData.hBytes.get(0).shiftLeft(64)).isEqualTo(mulData.hBytes.get(1)); // ZERO
  }

  @Test
  void hBytesWhereOneArgIsZero() {
    Bytes32 arg1 = Bytes32.ZERO;
    Bytes32 arg2 = Bytes32.ZERO;
    MulData mulData = new MulData(OpCode.EXP, arg1, arg2);
    mulData.setHsAndBitsFromBaseThetas(
        BaseTheta.fromBytes32(UInt256.ZERO), BaseTheta.fromBytes32(UInt256.valueOf(1)));

    assertThat(mulData.hBytes.get(0).isZero()).isTrue();
    assertThat(mulData.hBytes.get(1).isZero()).isTrue();
    assertThat(mulData.hBytes.get(2).isZero()).isTrue();
    assertThat(mulData.hBytes.get(3).isZero()).isTrue();

    assertThat(mulData.hBytes.get(0).shiftLeft(64)).isEqualTo(mulData.hBytes.get(1)); // ZERO
  }

  @Test
  void hBytes_5_5() {
    Bytes32 arg1 = Bytes32.fromHexString("0x05");
    Bytes32 arg2 = Bytes32.fromHexString("0x05");
    MulData mulData = new MulData(OpCode.EXP, arg1, arg2);
    mulData.setHsAndBitsFromBaseThetas(BaseTheta.fromBytes32(arg1), BaseTheta.fromBytes32(arg2));

    assertThat(mulData.hBytes.get(0).isZero()).isTrue();
    assertThat(mulData.hBytes.get(1).isZero()).isTrue();
    assertThat(mulData.hBytes.get(2).isZero()).isTrue();
    assertThat(mulData.hBytes.get(3).isZero()).isTrue();
  }

  @Test
  void hBytes_largeEnoughArgsToGetNonZeros() {
    // these args aren't used directly in the calculation of hs and bits
    Bytes32 arg1 = Bytes32.fromHexString("0x05");
    Bytes32 arg2 = Bytes32.fromHexString("0x05");

    BaseTheta aBaseTheta = BaseTheta.fromBytes32(UInt256.valueOf(43532)); // aa 0c // 170 12

    // squaring is one way to get a sufficiently big number
    BigInteger b = BigInteger.valueOf(82494664664768L);
    BigInteger b2 = b.multiply(b); // 6805369698152522037820493824
    UInt256 b2uint = UInt256.valueOf(b2);

    BaseTheta bBaseTheta = BaseTheta.fromBytes32(b2uint);

    MulData mulData = new MulData(OpCode.EXP, arg1, arg2);
    mulData.setHsAndBitsFromBaseThetas(aBaseTheta, bBaseTheta);

    Bytes h0 = Bytes.fromHexString("0x00000e9b37bfd908");

    assertThat(mulData.hBytes.get(0)).isEqualTo(h0);
    assertThat(mulData.hBytes.get(1).isZero()).isTrue();
    assertThat(mulData.hBytes.get(2).isZero()).isTrue();
    assertThat(mulData.hBytes.get(3).isZero()).isTrue();
  }

  @Test
  void hBytes_aReallyBigNumber() {
    // these args aren't used directly in the calculation of hs and bits
    Bytes32 arg1 = Bytes32.fromHexString("0x05");
    Bytes32 arg2 = Bytes32.fromHexString("0x05");

    BigInteger b = new BigInteger("296251353699975589350401737146368");
    UInt256 b4uint = UInt256.valueOf(b);

    BaseTheta aBaseTheta = BaseTheta.fromBytes32(b4uint);
    BaseTheta bBaseTheta = BaseTheta.fromBytes32(b4uint);

    MulData mulData = new MulData(OpCode.EXP, arg1, arg2);
    mulData.setHsAndBitsFromBaseThetas(aBaseTheta, bBaseTheta);

    Bytes h0 = Bytes.fromHexString("0xe9e4064d86460000");
    Bytes h1 = Bytes.fromHexString("0x00000779079ae9e3");

    assertThat(mulData.hBytes.get(0)).isEqualTo(h0);
    assertThat(mulData.hBytes.get(1)).isEqualTo(h1);
    assertThat(mulData.hBytes.get(2).isZero()).isTrue();
    assertThat(mulData.hBytes.get(3).isZero()).isTrue();
  }

  @Test
  void hBytesValue_and_generatesATrueBit() {
    // these args aren't used directly in the calculation of hs and bits
    Bytes32 arg1 = Bytes32.fromHexString("0x05");
    Bytes32 arg2 = Bytes32.fromHexString("0x05");

    BigInteger b = new BigInteger("3672491014949214151879080813");
    UInt256 b4uint = UInt256.valueOf(b);

    BaseTheta aBaseTheta = BaseTheta.fromBytes32(b4uint);
    BaseTheta bBaseTheta = BaseTheta.fromBytes32(b4uint);

    MulData mulData = new MulData(OpCode.EXP, arg1, arg2);
    mulData.setHsAndBitsFromBaseThetas(aBaseTheta, bBaseTheta);

    Bytes h0 = Bytes.fromHexString("0x8d222c056a351fb0");
    Bytes h1 = Bytes.fromHexString("0x0000000013fc1cf0");

    assertThat(mulData.hBytes.get(2).isZero()).isTrue();
    assertThat(mulData.hBytes.get(3).isZero()).isTrue();
    assertThat(mulData.hBytes.get(0)).isEqualTo(h0);
    assertThat(mulData.hBytes.get(1)).isEqualTo(h1);

    // bits
    // expected value obtained from go implementation debug output
    Boolean[] expectedBools = {false, false, false, false, false, true, false, false};
    assertThat(mulData.bits).isEqualTo(expectedBools);
  }

  @Test
  void hBytes_twoReallyBigNumbers_generatesADifferentTrueBit() {

    // these args aren't used directly in the calculation of hs and bits
    Bytes32 arg1 = Bytes32.fromHexString("0x05");
    Bytes32 arg2 = Bytes32.fromHexString("0x05");

    BigInteger b1 =
        new BigInteger(
            "22469423347992668196557015132986860313508181747976369840918221307635594854917");
    UInt256 b1uint = UInt256.valueOf(b1);
    BigInteger b2 =
        new BigInteger(
            "24978870742348927442211038043709369780953629326985996012322553405261539030097");
    UInt256 b2uint = UInt256.valueOf(b2);

    BaseTheta aBaseTheta = BaseTheta.fromBytes32(b1uint);
    BaseTheta bBaseTheta = BaseTheta.fromBytes32(b2uint);

    final MulData mulData = new MulData(OpCode.EXP, arg1, arg2);
    mulData.setHsAndBitsFromBaseThetas(aBaseTheta, bBaseTheta);

    BigInteger sum010 = new BigInteger("375860551383434850958895718584879559103");
    assertThat(Util.getOverflow(UInt256.valueOf(sum010), 3, "mu OOB")).isEqualTo(1);

    // bits
    // expected value obtained from go implementation debug output
    Boolean[] expectedBools = {false, false, false, false, false, false, true, false};
    assertThat(mulData.bits).isEqualTo(expectedBools);
  }
}
