package net.consensys.linea.zktracer.module.alu.mod;

import static org.assertj.core.api.Assertions.assertThat;

import net.consensys.linea.zktracer.bytes.Bytes16;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.Test;

public class BaseThetaTest {

  @Test
  public void shouldCreateBaseThetaCorrectly(){
    Bytes firstByte = Bytes.fromHexString("0x000000000000000a");
    Bytes secondByte = Bytes.fromHexString("0x000000000000000b");
    Bytes thirdByte = Bytes.fromHexString("0x000000000000000c");
    Bytes fourthByte = Bytes.fromHexString("0x000000000000000d");
    Bytes32 bytes32 = Bytes32.wrap(Bytes.concatenate(firstByte, secondByte, thirdByte, fourthByte));

    BaseTheta baseTheta =  BaseTheta.fromBytes32(bytes32);
    Bytes32 expectedBytes32 = Bytes32.wrap(Bytes.concatenate( fourthByte, thirdByte, secondByte, firstByte));

    assertThat(baseTheta.getBytes32()).isEqualTo(expectedBytes32);

    Bytes16 expectedLow =    Bytes16.wrap(Bytes.concatenate(secondByte, firstByte));
    Bytes16 expectedHigh =  Bytes16.wrap(Bytes.concatenate(fourthByte, thirdByte));

    assertThat(baseTheta.getLow()).isEqualTo(expectedLow);
    assertThat(baseTheta.getHigh()).isEqualTo(expectedHigh);
    assertThat(baseTheta.getBytes(0)).isEqualTo(fourthByte);
    assertThat(baseTheta.getBytes(1)).isEqualTo(thirdByte);
    assertThat(baseTheta.getBytes(2)).isEqualTo(secondByte);
    assertThat(baseTheta.getBytes(3)).isEqualTo(firstByte);
  }

  @Test
  public void shouldCreateBaseBytesCorrectly(){
    Bytes32 bytes32 = Bytes32.fromHexString("0x000000000000000a000000000000000b000000000000000c000000000000000d");
    BaseBytes baseBytes =  BaseBytes.fromBytes32(bytes32);
    Bytes16 expectedHigh = Bytes16.fromHexString("0x000000000000000a000000000000000b");
    Bytes16 expectedLow =  Bytes16.fromHexString("0x000000000000000c000000000000000d");

    assertThat(baseBytes.getBytes32()).isEqualTo(bytes32);
    assertThat(baseBytes.getLow()).isEqualTo(expectedLow);
    assertThat(baseBytes.getHigh()).isEqualTo(expectedHigh);
  }
}
