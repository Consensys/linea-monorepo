package net.consensys.zktracer.module.alu;

import java.util.List;
import java.util.stream.Stream;
import net.consensys.zktracer.OpCode;
import net.consensys.zktracer.module.AbstractModuleTracerTest;
import net.consensys.zktracer.module.ModuleTracer;
import net.consensys.zktracer.module.alu.mod.ModTracer;

import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.extension.ExtendWith;

import org.junit.jupiter.params.provider.Arguments;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
class ModTracerTest extends AbstractModuleTracerTest {

  @Override
  protected Stream<Arguments> provideNonRandomArguments() {
    Bytes32 arg1 =
        Bytes32.fromHexString("0x09605496114a8cec5589a61b2d68a7a48ad7a06e09a8ebb9253536728c65498b");
    Bytes32 arg2 =
        Bytes32.fromHexString("0x8f4e3a60e003b11a6517a326d250dc23ea43e8b00b94f38079413de350f48c93");

    return Stream.of(Arguments.of(OpCode.MOD, arg1, arg2));
  }

  @Override
  protected ModuleTracer getModuleTracer() {
    return new ModTracer();
  }
}