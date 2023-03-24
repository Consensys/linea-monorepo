package net.consensys.zktracer.module.alu.add;

import net.consensys.zktracer.OpCode;
import net.consensys.zktracer.bytes.Bytes16;
import net.consensys.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

import java.util.List;

public class AddTracer implements ModuleTracer {


    private int stamp = 0;
    @Override
    public String jsonKey() {
        return "add";
    }

    @Override
    public List<OpCode> supportedOpCodes() {
         return List.of(OpCode.ADD, OpCode.ADDMOD);
    }

    @SuppressWarnings({"UnusedVariable"})
    @Override
    public Object trace(MessageFrame frame) {
        // TODO duplicated code
        final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
        final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));

        // TODO duplicated code
        final Bytes16 arg1Hi = Bytes16.wrap(arg1.slice(0, 16));
        final Bytes16 arg1Lo = Bytes16.wrap(arg1.slice(16));
        final Bytes16 arg2Hi = Bytes16.wrap(arg2.slice(0, 16));
        final Bytes16 arg2Lo = Bytes16.wrap(arg2.slice(16));

        final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());


        final AddTrace.Trace.Builder builder = AddTrace.Trace.Builder.newInstance();

        stamp++;

        builder.setStamp(stamp);
        return builder.build();
    }
}
