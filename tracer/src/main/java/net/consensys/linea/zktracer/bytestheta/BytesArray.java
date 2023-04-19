package net.consensys.linea.zktracer.bytestheta;

import static com.google.common.base.Preconditions.checkArgument;

import org.apache.tuweni.bytes.Bytes;

public class BytesArray {

  public BytesArray(final byte[][] bytes) {
    this.bytesArray = bytes;
  }

  final byte[][] bytesArray;

  public BytesArray(int size) {
    bytesArray = new byte[size][8];
  }

  public Bytes get(int index) {
    return Bytes.wrap(bytesArray[index]);
  }

  public void set(int index, Bytes bytes) {
    checkArgument(bytes.size() == 8);
    this.bytesArray[index] = bytes.toArray();
  }

  public Bytes[] getBytesRange(final int start, final int end) {
    int rangeSize = end - start + 1;
    Bytes[] bytes = new Bytes[rangeSize];
    for (int i = 0; i < rangeSize; i++) {
      bytes[i] = get(start + i);
    }
    return bytes;
  }
}
