package net.consensys.linea.zktracer.bytes;

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

import static org.assertj.core.api.Assertions.assertThat;

import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.Test;

public class BaseThetaTest {

  @Test
  public void baseThetaTest() {
    Bytes firstByte = Bytes.fromHexString("0x000000000000000a");
    Bytes secondByte = Bytes.fromHexString("0x000000000000000b");
    Bytes thirdByte = Bytes.fromHexString("0x000000000000000c");
    Bytes fourthByte = Bytes.fromHexString("0x000000000000000d");
    Bytes32 bytes32 = Bytes32.wrap(Bytes.concatenate(firstByte, secondByte, thirdByte, fourthByte));

    BaseTheta baseTheta = BaseTheta.fromBytes32(bytes32);
    Bytes32 expectedBytes32 =
        Bytes32.wrap(Bytes.concatenate(fourthByte, thirdByte, secondByte, firstByte));

    assertThat(baseTheta.getBytes32()).isEqualTo(expectedBytes32);

    Bytes16 expectedLow = Bytes16.wrap(Bytes.concatenate(thirdByte, fourthByte));
    Bytes16 expectedHigh = Bytes16.wrap(Bytes.concatenate(firstByte, secondByte));

    assertThat(baseTheta.getLow()).isEqualTo(expectedLow);
    assertThat(baseTheta.getHigh()).isEqualTo(expectedHigh);

    assertThat(baseTheta.get(0)).isEqualTo(fourthByte);
    assertThat(baseTheta.get(1)).isEqualTo(thirdByte);
    assertThat(baseTheta.get(2)).isEqualTo(secondByte);
    assertThat(baseTheta.get(3)).isEqualTo(firstByte);
  }

  @Test
  public void baseBytesTest() {
    Bytes32 bytes32 =
        Bytes32.fromHexString("0x000000000000000a000000000000000b000000000000000c000000000000000d");
    BaseBytes baseBytes = BaseBytes.fromBytes32(bytes32);
    Bytes16 expectedHigh = Bytes16.fromHexString("0x000000000000000a000000000000000b");
    Bytes16 expectedLow = Bytes16.fromHexString("0x000000000000000c000000000000000d");

    assertThat(baseBytes.getBytes32()).isEqualTo(bytes32);
    assertThat(baseBytes.getLow()).isEqualTo(expectedLow);
    assertThat(baseBytes.getHigh()).isEqualTo(expectedHigh);
  }

  @Test
  public void getRangeTest() {
    Bytes firstByte = Bytes.fromHexString("0x000000000000000a");
    Bytes secondByte = Bytes.fromHexString("0x000000000000000b");
    Bytes thirdByte = Bytes.fromHexString("0x000000000000000c");
    Bytes fourthByte = Bytes.fromHexString("0x000000000000000d");
    Bytes32 bytes32 = Bytes32.wrap(Bytes.concatenate(firstByte, secondByte, thirdByte, fourthByte));

    BaseTheta baseTheta = BaseTheta.fromBytes32(bytes32);

    assertThat(baseTheta.getRange(0, 6, 7)).isEqualTo(Bytes.fromHexString("0x000a"));
    assertThat(baseTheta.getRange(2, 5, 7)).isEqualTo(Bytes.fromHexString("0x00000c"));
    assertThat(baseTheta.getRange(2, 0, 1)).isEqualTo(Bytes.fromHexString("0x0000"));
  }

  @Test
  public void getTest() {
    Bytes firstByte = Bytes.fromHexString("0x000000000000000a");
    Bytes secondByte = Bytes.fromHexString("0x000000000000000b");
    Bytes thirdByte = Bytes.fromHexString("0x000000000000000c");
    Bytes fourthByte = Bytes.fromHexString("0x000000000000000d");
    Bytes32 bytes32 = Bytes32.wrap(Bytes.concatenate(firstByte, secondByte, thirdByte, fourthByte));

    BaseTheta baseTheta = BaseTheta.fromBytes32(bytes32);

    assertThat(baseTheta.get(0, 7)).isEqualTo(Bytes.fromHexString("0x0a"));
    assertThat(baseTheta.get(2, 7)).isEqualTo(Bytes.fromHexString("0x0c"));
  }

  @Test
  public void setTest() {
    Bytes firstByte = Bytes.fromHexString("0x000000000000000a");
    Bytes secondByte = Bytes.fromHexString("0x000000000000000b");
    Bytes thirdByte = Bytes.fromHexString("0x000000000000000c");
    Bytes fourthByte = Bytes.fromHexString("0x000000000000000d");
    Bytes32 bytes32 = Bytes32.wrap(Bytes.concatenate(firstByte, secondByte, thirdByte, fourthByte));

    // equal before
    BaseTheta baseTheta = BaseTheta.fromBytes32(bytes32);
    assertThat(baseTheta.getBytes32()).isEqualTo(bytes32);

    Bytes32 modifiedBytes32 =
        Bytes32.fromHexString("0x0000000000000009000000000000000b000000000000000c080000000000000d");

    baseTheta.set(0, 7, (byte) '9');
    baseTheta.set(3, 0, (byte) '8');

    // equal after modifications
    assertThat(baseTheta.getBytes32()).isEqualTo(modifiedBytes32);
  }
}
