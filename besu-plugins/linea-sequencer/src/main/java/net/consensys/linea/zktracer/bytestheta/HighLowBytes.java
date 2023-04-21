package net.consensys.linea.zktracer.bytestheta;

import org.apache.tuweni.bytes.Bytes;

public interface HighLowBytes {
  Bytes getHigh();

  Bytes getLow();
}
