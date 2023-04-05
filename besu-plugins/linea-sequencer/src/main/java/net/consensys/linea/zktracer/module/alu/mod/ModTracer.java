package net.consensys.linea.zktracer.module.alu.mod;

import java.util.List;
import net.consensys.linea.zktracer.OpCode;

import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class ModTracer implements ModuleTracer {
  private int stamp = 0;
  private final int	MMEDIUM = 8;

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
      builder
          .appendModStamp(stamp)
          .appendOli(data.isOli())
          .appendCt(ct)
          .appendInst(UnsignedByte.of(opCode.value))
          .appendDecSigned(data.isSigned())
          .appendDecOutput(data.isDiv())

          .appendArg1Hi(data.getArg1Hi().toUnsignedBigInteger())
          .appendArg1Lo(data.getArg1Lo().toUnsignedBigInteger())
          .appendArg2Hi(data.getArg2Hi().toUnsignedBigInteger())
          .appendArg2Lo(data.getArg2Lo().toUnsignedBigInteger())

          .appendResHi(data.getResHi().toUnsignedBigInteger())
          .appendResLo(data.getResHi().toUnsignedBigInteger())

          .appendByte_1_3(UnsignedByte.of(data.getArg1().get(ct)))
          .appendByte_1_2(UnsignedByte.of(data.getArg1().get(ct + 8)))

          .appendAcc_1_2(data.getArg1().slice(8, ct+1).toUnsignedBigInteger())
          .appendAcc_1_3(data.getArg1().slice(0, ct+1).toUnsignedBigInteger())

          .appendByte_2_3(UnsignedByte.of(data.getArg2().get(ct)))
          .appendByte_2_2(UnsignedByte.of(data.getArg2().get(ct + 8)))


          .appendAcc_2_2(data.getArg2().slice(8, ct+1).toUnsignedBigInteger())
          .appendAcc_2_3(data.getArg2().slice(0, ct+1).toUnsignedBigInteger())

          .appendAcc_B_0(Bytes.wrap(data.getB_Bytes().get(0)).slice(0, ct + 1).toUnsignedBigInteger());


/*
	self.Trace.PushBytes(ACC_B_3.Name(), ad.B_Bytes[3][:ct+1])
	self.Trace.PushBytes(ACC_B_2.Name(), ad.B_Bytes[2][:ct+1])
	self.Trace.PushBytes(ACC_B_1.Name(), ad.B_Bytes[1][:ct+1])
	self.Trace.PushBytes(ACC_B_0.Name(), ad.B_Bytes[0][:ct+1])

/*
          .appendByte_B_3(UnsignedByte.of(data.getB_Bytes().getBytes()[3][ct]))
          .appendByte_B_2(UnsignedByte.of(data.getB_Bytes().getBytes()[2][ct]))
          .appendByte_B_1(UnsignedByte.of(data.getB_Bytes().getBytes()[1][ct]))
          .appendByte_B_0(UnsignedByte.of(data.getB_Bytes().getBytes()[0][ct]))
*/

/*          .appendAcc_B_3(
              Bytes.wrap(data.getB_Bytes().getBytes()[0]).slice(0, ct).toUnsignedBigInteger());*/

    }
    return builder.build();


    /* self.Trace.PushBytes(ACC_B_3.Name(), ad.B_Bytes[3][:ct+1])
    self.Trace.PushBytes(ACC_B_2.Name(), ad.B_Bytes[2][:ct+1])
    self.Trace.PushBytes(ACC_B_1.Name(), ad.B_Bytes[1][:ct+1])
    self.Trace.PushBytes(ACC_B_0.Name(), ad.B_Bytes[0][:ct+1])
    //
    self.Trace.PushByte(BYTE_Q_3.Name(), ad.Q_Bytes[3][ct])
    self.Trace.PushByte(BYTE_Q_2.Name(), ad.Q_Bytes[2][ct])
    self.Trace.PushByte(BYTE_Q_1.Name(), ad.Q_Bytes[1][ct])
    self.Trace.PushByte(BYTE_Q_0.Name(), ad.Q_Bytes[0][ct])
    self.Trace.PushBytes(ACC_Q_3.Name(), ad.Q_Bytes[3][:ct+1])
    self.Trace.PushBytes(ACC_Q_2.Name(), ad.Q_Bytes[2][:ct+1])
    self.Trace.PushBytes(ACC_Q_1.Name(), ad.Q_Bytes[1][:ct+1])
    self.Trace.PushBytes(ACC_Q_0.Name(), ad.Q_Bytes[0][:ct+1])
    //
    self.Trace.PushByte(BYTE_R_3.Name(), ad.R_Bytes[3][ct])
    self.Trace.PushByte(BYTE_R_2.Name(), ad.R_Bytes[2][ct])
    self.Trace.PushByte(BYTE_R_1.Name(), ad.R_Bytes[1][ct])
    self.Trace.PushByte(BYTE_R_0.Name(), ad.R_Bytes[0][ct])
    self.Trace.PushBytes(ACC_R_3.Name(), ad.R_Bytes[3][:ct+1])
    self.Trace.PushBytes(ACC_R_2.Name(), ad.R_Bytes[2][:ct+1])
    self.Trace.PushBytes(ACC_R_1.Name(), ad.R_Bytes[1][:ct+1])
    self.Trace.PushBytes(ACC_R_0.Name(), ad.R_Bytes[0][:ct+1])
    //
    self.Trace.PushByte(BYTE_DELTA_3.Name(), ad.DELTA_Bytes[3][ct])
    self.Trace.PushByte(BYTE_DELTA_2.Name(), ad.DELTA_Bytes[2][ct])
    self.Trace.PushByte(BYTE_DELTA_1.Name(), ad.DELTA_Bytes[1][ct])
    self.Trace.PushByte(BYTE_DELTA_0.Name(), ad.DELTA_Bytes[0][ct])
    self.Trace.PushBytes(ACC_DELTA_3.Name(), ad.DELTA_Bytes[3][:ct+1])
    self.Trace.PushBytes(ACC_DELTA_2.Name(), ad.DELTA_Bytes[2][:ct+1])
    self.Trace.PushBytes(ACC_DELTA_1.Name(), ad.DELTA_Bytes[1][:ct+1])
    self.Trace.PushBytes(ACC_DELTA_0.Name(), ad.DELTA_Bytes[0][:ct+1])
    //
    self.Trace.PushByte(BYTE_H_2.Name(), ad.H_Bytes[2][ct])
    self.Trace.PushByte(BYTE_H_1.Name(), ad.H_Bytes[1][ct])
    self.Trace.PushByte(BYTE_H_0.Name(), ad.H_Bytes[0][ct])
    self.Trace.PushBytes(ACC_H_2.Name(), ad.H_Bytes[2][:ct+1])
    self.Trace.PushBytes(ACC_H_1.Name(), ad.H_Bytes[1][:ct+1])
    self.Trace.PushBytes(ACC_H_0.Name(), ad.H_Bytes[0][:ct+1])
    //
    self.Trace.PushBool(CMP_1.Name(), ad.cmp1[ct])
    self.Trace.PushBool(CMP_2.Name(), ad.cmp2[ct])
    self.Trace.PushBool(MSB_1.Name(), ad.msb1[ct])
    self.Trace.PushBool(MSB_2.Name(), ad.msb2[ct])*/
  }

  private int maxCounter(ModData data) {
    if (data.isOli()) {
      return 1;
    } else {
      return MMEDIUM;
    }
  }
}
