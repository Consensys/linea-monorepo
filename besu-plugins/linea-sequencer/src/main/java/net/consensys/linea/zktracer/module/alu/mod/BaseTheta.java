package net.consensys.linea.zktracer.module.alu.mod;

import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

public class BaseTheta  extends BaseBytes{

  private BaseTheta(final Bytes32 arg) {
    super(arg);
    bytes32 = arg.mutableCopy();
    for (int k = 0; k < 4; k++) {
      Bytes bytes = arg.slice(OFFSET * k, OFFSET);
      setBytes( OFFSET * (3-k) , bytes);
    }
  }

  static BaseTheta fromBytes32(Bytes32 arg){
    return new BaseTheta(arg);
  }
  public void setBytes(int index, Bytes bytes){
    bytes32.set(index, bytes);
  }
  public Bytes getBytes(int index) {
    return bytes32.slice(OFFSET * index, OFFSET);
  }
  public Bytes slice (int i, int length){
    return bytes32.slice(i, length);
  }
}
