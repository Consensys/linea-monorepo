package net.consensys.linea.zktracer.module.alu.ext.calculator.addmod;

import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import net.consensys.linea.zktracer.bytestheta.BytesArray;
import net.consensys.linea.zktracer.module.alu.ext.calculator.AbstractExtCalculator;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

public class AddModCalculator extends AbstractExtCalculator {

  @Override
  public UInt256 computeResult(final Bytes32 arg1, final Bytes32 arg2, final Bytes32 arg3) {
    return UInt256.fromBytes(arg1).addMod(UInt256.fromBytes(arg2), UInt256.fromBytes(arg3));
  }

  @Override
  public BytesArray computeJs(final Bytes32 arg1, final Bytes32 arg2) {
    return AddModBytesJCalculator.computeJs(arg1, arg2);
  }

  @Override
  public BytesArray computeQs(final Bytes32 arg1, final Bytes32 arg2, final Bytes32 arg3) {
    return AddModBytesQCalculator.computeQs(arg1, arg2, arg3);
  }

  /**
   * Computes the overflow result for the given arguments.
   *
   * @param arg1 the arg1 value.
   * @param arg2 the arg2 value.
   * @param aBytes the aBytes value.
   * @param bBytes the bBytes value.
   * @param hBytes the hBytes value.
   * @param alpha the alpha value.
   * @param beta the beta value.
   * @return the overflow result.
   */
  @Override
  public boolean[] computeOverflowRes(
      final BaseBytes arg1,
      final BaseBytes arg2,
      final BytesArray aBytes,
      final BytesArray bBytes,
      final BytesArray hBytes,
      final UInt256 alpha,
      final UInt256 beta) {
    return AddModOverflowResCalculator.calculateOverflow(arg1, arg2);
  }
}
