package net.consensys.linea.zktracer.module.alu.mod;

import org.hyperledger.besu.evm.frame.MessageFrame;

import java.util.List;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

public class ModTracer implements ModuleTracer {
  private int stamp = 0;
  private final int MMEDIUM = 8;

  @Override
  public String jsonKey() {
    return "mod";
  }

  @Override
  public List<OpCode> supportedOpCodes() {
    return List.of(OpCode.DIV, OpCode.SDIV, OpCode.MOD, OpCode.SMOD);
  }

  @Override
  public Object trace(final MessageFrame frame) {

    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));

    final ModData data = new ModData(opCode, arg1, arg2);
    final ModTrace.Trace.Builder builder = ModTrace.Trace.Builder.newInstance();
    stamp++;
    for (int ct = 0; ct < maxCounter(data); ct++) {
      final int accLength = ct + 1;
      builder
          .appendModStamp(stamp)
          .appendOli(data.isOli())
          .appendCt(ct)
          .appendInst(UnsignedByte.of(opCode.value))
          .appendDecSigned(data.isSigned())
          .appendDecOutput(data.isDiv())
          .appendArg1Hi(data.getArg1().getHigh().toUnsignedBigInteger())
          .appendArg1Lo(data.getArg1().getLow().toUnsignedBigInteger())
          .appendArg2Hi(data.getArg2().getHigh().toUnsignedBigInteger())
          .appendArg2Lo(data.getArg2().getLow().toUnsignedBigInteger())
          .appendResHi(data.getResult().getHigh().toUnsignedBigInteger())
          .appendResLo(data.getResult().getLow().toUnsignedBigInteger())
          .appendAcc1_2(data.getArg1().getBytes32().slice(8, ct + 1).toUnsignedBigInteger())
          .appendAcc1_3(data.getArg1().getBytes32().slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc2_2(data.getArg2().getBytes32().slice(8, ct + 1).toUnsignedBigInteger())
          .appendAcc2_3(data.getArg2().getBytes32().slice(0, ct + 1).toUnsignedBigInteger())
          .appendAccB0(data.getBBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccB1(data.getBBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccB2(data.getBBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccB3(data.getBBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .appendAccR0(data.getRBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccR1(data.getRBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccR2(data.getRBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccR3(data.getRBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .appendAccQ0(data.getQBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccQ1(data.getQBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccQ2(data.getQBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccQ3(data.getQBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .appendAccDelta0(data.getDeltaBytes().get(0).slice(0, accLength).toUnsignedBigInteger())
          .appendAccDelta1(data.getDeltaBytes().get(1).slice(0, accLength).toUnsignedBigInteger())
          .appendAccDelta2(data.getDeltaBytes().get(2).slice(0, accLength).toUnsignedBigInteger())
          .appendAccDelta3(data.getDeltaBytes().get(3).slice(0, accLength).toUnsignedBigInteger())
          .appendByte2_2(UnsignedByte.of(data.getArg2().getByte(ct + 8)))
          .appendByte2_3(UnsignedByte.of(data.getArg2().getByte(ct)))
          .appendByte1_2(UnsignedByte.of(data.getArg1().getByte(ct + 8)))
          .appendByte1_3(UnsignedByte.of(data.getArg1().getByte(ct)))
          .appendByteB0(UnsignedByte.of(data.getBBytes().get(0).get(ct)))
          .appendByteB1(UnsignedByte.of(data.getBBytes().get(1).get(ct)))
          .appendByteB2(UnsignedByte.of(data.getBBytes().get(2).get(ct)))
          .appendByteB3(UnsignedByte.of(data.getBBytes().get(3).get(ct)))
          .appendByteR0(UnsignedByte.of(data.getRBytes().get(0).get(ct)))
          .appendByteR1(UnsignedByte.of(data.getRBytes().get(1).get(ct)))
          .appendByteR2(UnsignedByte.of(data.getRBytes().get(2).get(ct)))
          .appendByteR3(UnsignedByte.of(data.getRBytes().get(3).get(ct)))
          .appendByteQ0(UnsignedByte.of(data.getQBytes().get(0).get(ct)))
          .appendByteQ1(UnsignedByte.of(data.getQBytes().get(1).get(ct)))
          .appendByteQ2(UnsignedByte.of(data.getQBytes().get(2).get(ct)))
          .appendByteQ3(UnsignedByte.of(data.getQBytes().get(3).get(ct)))
          .appendByteDelta0(UnsignedByte.of(data.getDeltaBytes().get(0).get(ct)))
          .appendByteDelta1(UnsignedByte.of(data.getDeltaBytes().get(1).get(ct)))
          .appendByteDelta2(UnsignedByte.of(data.getDeltaBytes().get(2).get(ct)))
          .appendByteDelta3(UnsignedByte.of(data.getDeltaBytes().get(3).get(ct)))
          .appendByteH0(UnsignedByte.of(data.getHBytes().get(0).get(ct)))
          .appendByteH1(UnsignedByte.of(data.getHBytes().get(1).get(ct)))
          .appendByteH2(UnsignedByte.of(data.getHBytes().get(2).get(ct)))
          .appendAccH0(Bytes.wrap(data.getHBytes().get(0)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAccH1(Bytes.wrap(data.getHBytes().get(1)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAccH2(Bytes.wrap(data.getHBytes().get(2)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendCmp1(data.getCmp1()[ct])
          .appendCmp2(data.getCmp2()[ct])
          .appendMsb1(data.getMsb1()[ct])
          .appendMsb2(data.getMsb2()[ct]);
    }
    builder.setStamp(stamp);
    return builder.build();
  }

  private int maxCounter(ModData data) {
    if (data.isOli()) {
      return 1;
    } else {
      return MMEDIUM;
    }
  }
}
