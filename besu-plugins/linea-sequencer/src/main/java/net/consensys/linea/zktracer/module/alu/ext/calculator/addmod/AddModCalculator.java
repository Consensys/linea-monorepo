package net.consensys.linea.zktracer.module.alu.ext.calculator.addmod;

import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import net.consensys.linea.zktracer.bytestheta.BytesArray;
import net.consensys.linea.zktracer.module.alu.ext.calculator.AbstractExtCalculator;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

public class AddModCalculator extends AbstractExtCalculator {

  @Override
  public UInt256 computeResult(final Bytes32 arg1, final Bytes32 arg2, final Bytes32 arg3) {
     throw new RuntimeException("Not Implemented") ;
  }

  @Override
  public boolean[] computeOverflowRes(
      final BaseTheta aBytes,
      final BaseTheta bBytes,
      final BytesArray hBytes,
      final UInt256 alpha,
      final UInt256 beta) {
     throw new RuntimeException("Not Implemented") ;
  }

  @Override
  public BytesArray computeJs(final Bytes32 arg1, final Bytes32 arg2) {
     throw new RuntimeException("Not Implemented") ;
  }

  @Override
  public BytesArray computeQs(final Bytes32 arg1, final Bytes32 arg2, final Bytes32 arg3) {
    throw new RuntimeException("Not Implemented");
  }
}
