package net.consensys.linea.zktracer.module.trm;

import org.hyperledger.besu.evm.frame.MessageFrame;

import java.util.List;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.module.ModuleTracer;

public class TrmTracer implements ModuleTracer {

  @Override
  public String jsonKey() {
    return "trm";
  }

  @Override
  public List<OpCode> supportedOpCodes() {
    return List.of(
        OpCode.BALANCE,
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
}
