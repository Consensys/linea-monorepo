package net.consensys.linea.zktracer.module.alu.mod;

import net.consensys.linea.zktracer.bytes.Bytes16;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.bytes.MutableBytes32;

public class BaseBytes {
  protected final int OFFSET = 8;
  private final int LOW_HIGH_SIZE = 16;
  protected MutableBytes32 bytes32;

  static BaseBytes fromBytes32(Bytes32 arg){
    return new BaseBytes(arg);
  }

  protected BaseBytes(final Bytes32 arg) {
    bytes32 = arg.mutableCopy();
  }

  public Bytes16 getHigh() {
    return Bytes16.wrap(bytes32.slice(0, LOW_HIGH_SIZE));
  }

  public Bytes16 getLow() {
    return Bytes16.wrap(bytes32.slice(LOW_HIGH_SIZE));
  }

  public byte getByte(int index){
    return bytes32.get(index);
  }

  public Bytes32 getBytes32(){
    return bytes32;
  }
}
