package net.consensys.linea.zktracer.bytestheta;

import org.apache.tuweni.bytes.Bytes;

public interface HighLowBytes {
  /**
   * Returns the high part of the bytes object.
   *
   * @return the high part of the bytes object
   */
  Bytes getHigh();
  /**
   * Returns the low part of the bytes object.
   *
   * @return the low part of the bytes object
   */
  Bytes getLow();
}
