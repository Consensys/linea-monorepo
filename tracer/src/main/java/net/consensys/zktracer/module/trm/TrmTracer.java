package net.consensys.zktracer.module.trm;

import net.consensys.zktracer.OpCode;
import net.consensys.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

import java.util.List;

public class TrmTracer implements ModuleTracer {

    @Override
    public String jsonKey() {
        return "trm";
    }

    @Override
    public List<OpCode> supportedOpCodes() {
        return List.of(OpCode.BALANCE,
                OpCode.EXTCODESIZE,
                OpCode.EXTCODECOPY,
                OpCode.EXTCODEHASH,
                OpCode.CALL,
                OpCode.CALLCODE,
                OpCode.DELEGATECALL,
                OpCode.STATICCALL,
                OpCode.SELFDESTRUCT);
    }

    @Override
    public Object trace(MessageFrame frame) {
        return null;
    }

    @Override
    public int lineCount(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
        // TODO CT column counts from 0 to 15 unless the TRM[=] is zero, in which case it hovers at zero
        return 0;
    }
}
