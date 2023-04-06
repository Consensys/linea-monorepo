package net.consensys.linea.zktracer.module.alu.mod;

import static net.consensys.linea.zktracer.module.alu.mod.ModBuilderMapper.ACC_B;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.function.BiConsumer;
import java.util.function.BiFunction;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.function.Supplier;
import kotlin.reflect.KFunction;
import net.consensys.linea.zktracer.OpCode;

import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.ModuleTracer;
import net.consensys.linea.zktracer.module.alu.mod.ModTrace.Trace.Builder;
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

          .appendArg1Hi(data.getArg1().getHigh().toUnsignedBigInteger())
          .appendArg1Lo(data.getArg1().getLow().toUnsignedBigInteger())
          .appendArg2Hi(data.getArg2().getHigh().toUnsignedBigInteger())
          .appendArg2Lo(data.getArg2().getLow().toUnsignedBigInteger())

          .appendResHi(data.getResult().getHigh().toUnsignedBigInteger())
          .appendResLo(data.getResult().getLow().toUnsignedBigInteger())

          .appendByte_1_2(UnsignedByte.of(data.getArg1().getByte(ct + 8)))
          .appendByte_1_3(UnsignedByte.of(data.getArg1().getByte(ct)))
          .appendAcc_1_2(data.getArg1().getBytes32().slice(8, ct+1).toUnsignedBigInteger())
          .appendAcc_1_3(data.getArg1().getBytes32().slice(0, ct+1).toUnsignedBigInteger())

          .appendByte_2_2(UnsignedByte.of(data.getArg2().getByte(ct + 8)))
          .appendByte_2_3(UnsignedByte.of(data.getArg2().getByte(ct)))
          .appendAcc_2_2(data.getArg2().getBytes32().slice(8, ct+1).toUnsignedBigInteger())
          .appendAcc_2_3(data.getArg2().getBytes32().slice(0, ct+1).toUnsignedBigInteger())

          .appendAcc_B_0(Bytes.wrap(data.getB_Bytes().getBytes(0)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc_B_1(Bytes.wrap(data.getB_Bytes().getBytes(1)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc_B_2(Bytes.wrap(data.getB_Bytes().getBytes(2)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc_B_3(Bytes.wrap(data.getB_Bytes().getBytes(3)).slice(0, ct + 1).toUnsignedBigInteger())

          .appendByte_B_0(UnsignedByte.of(data.getB_Bytes().getBytes(0).get(ct)))
          .appendByte_B_1(UnsignedByte.of(data.getB_Bytes().getBytes(1).get(ct)))
          .appendByte_B_2(UnsignedByte.of(data.getB_Bytes().getBytes(2).get(ct)))
          .appendByte_B_3(UnsignedByte.of(data.getB_Bytes().getBytes(3).get(ct)))

          .appendByte_R_0(UnsignedByte.of(data.getR_Bytes().getBytes(0).get(ct)))
          .appendByte_R_1(UnsignedByte.of(data.getR_Bytes().getBytes(1).get(ct)))
          .appendByte_R_2(UnsignedByte.of(data.getR_Bytes().getBytes(2).get(ct)))
          .appendByte_R_3(UnsignedByte.of(data.getR_Bytes().getBytes(3).get(ct)))

          .appendAcc_R_0(Bytes.wrap(data.getR_Bytes().getBytes(0)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc_R_1(Bytes.wrap(data.getR_Bytes().getBytes(1)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc_R_2(Bytes.wrap(data.getR_Bytes().getBytes(2)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc_R_3(Bytes.wrap(data.getR_Bytes().getBytes(3)).slice(0, ct + 1).toUnsignedBigInteger())

          .appendByte_Q_0(UnsignedByte.of(data.getQ_Bytes().getBytes(0).get(ct)))
          .appendByte_Q_1(UnsignedByte.of(data.getQ_Bytes().getBytes(1).get(ct)))
          .appendByte_Q_2(UnsignedByte.of(data.getQ_Bytes().getBytes(2).get(ct)))
          .appendByte_Q_3(UnsignedByte.of(data.getQ_Bytes().getBytes(3).get(ct)))

          .appendAcc_Q_0(Bytes.wrap(data.getQ_Bytes().getBytes(0)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc_Q_1(Bytes.wrap(data.getQ_Bytes().getBytes(1)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc_Q_2(Bytes.wrap(data.getQ_Bytes().getBytes(2)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc_Q_3(Bytes.wrap(data.getQ_Bytes().getBytes(3)).slice(0, ct + 1).toUnsignedBigInteger())

          .appendByteDelta_0(UnsignedByte.of(data.getDeltaBytes().getBytes(0).get(ct)))
          .appendByteDelta_1(UnsignedByte.of(data.getDeltaBytes().getBytes(1).get(ct)))
          .appendByteDelta_2(UnsignedByte.of(data.getDeltaBytes().getBytes(2).get(ct)))
          .appendByteDelta_3(UnsignedByte.of(data.getDeltaBytes().getBytes(3).get(ct)))

          .appendAccDelta_0(Bytes.wrap(data.getDeltaBytes().getBytes(0)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAccDelta_1(Bytes.wrap(data.getDeltaBytes().getBytes(1)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAccDelta_2(Bytes.wrap(data.getDeltaBytes().getBytes(2)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAccDelta_3(Bytes.wrap(data.getDeltaBytes().getBytes(3)).slice(0, ct + 1).toUnsignedBigInteger())

          .appendByte_H_0(UnsignedByte.of(data.getH_Bytes().getBytes(0).get(ct)))
          .appendByte_H_1(UnsignedByte.of(data.getH_Bytes().getBytes(1).get(ct)))
          .appendByte_H_2 (UnsignedByte.of(data.getH_Bytes().getBytes(2).get(ct)))

          .appendAcc_H_0(Bytes.wrap(data.getH_Bytes().getBytes(0)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc_H_1(Bytes.wrap(data.getH_Bytes().getBytes(1)).slice(0, ct + 1).toUnsignedBigInteger())
          .appendAcc_H_2(Bytes.wrap(data.getH_Bytes().getBytes(2)).slice(0, ct + 1).toUnsignedBigInteger())

          .appendCmp1(data.getCmp1()[ct])
          .appendCmp2(data.getCmp2()[ct])

          .appendMsb1(data.getMsb1()[ct])
          .appendMsb2(data.getMsb2()[ct]);

      ModBuilderMapper mapper =  new ModBuilderMapper(builder);
      appendAcc(mapper, ACC_B, data.getB_Bytes().getBytes32(), ct);
    }
    return builder.build();
  }

  private void appendAcc(ModBuilderMapper mapper, String accName, Bytes32 data, int ct){
    Function<Integer, BigInteger> function =
        index -> data.slice( index *  8, + ct + 1)
            .toUnsignedBigInteger();

    mapper.get(accName)
         .forEach((integer, consumer) -> consumer.accept(
             function.apply(integer)
         )
  );
  }

  private int maxCounter(ModData data) {
    if (data.isOli()) {
      return 1;
    } else {
      return MMEDIUM;
    }
  }
}
