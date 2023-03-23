package net.consensys.zktracer.module.alu.add;

import net.consensys.zktracer.OpCode;
import net.consensys.zktracer.bytes.Bytes16;
import net.consensys.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.plugin.evm.frame.MessageFrame;

import java.util.List;

public class AddTracer implements ModuleTracer {

    @Override
    public String jsonKey() {
        return "add";
    }

    @Override
    public List<OpCode> supportedOpCodes() {
         return List.of(OpCode.ADD, OpCode.ADDMOD);
    }

    @Override
    public Object trace(final MessageFrame frame) {
        final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
        final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));

        final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
        return null;
    }
}
