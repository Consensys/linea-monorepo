package net.consensys.linea.zktracer.module.alu.mod;

import java.math.BigInteger;
import java.util.HashMap;
import java.util.function.Consumer;
import net.consensys.linea.zktracer.module.alu.mod.ModTrace.Trace.Builder;

public class ModBuilderMapper {

  private final HashMap<String, HashMap<Integer, Consumer<BigInteger>>> consumers
      = new HashMap<>();
  public static String ACC_B = "ACC_B";

  public ModBuilderMapper(Builder builder){
    HashMap<Integer, Consumer<BigInteger>> accB =  new HashMap<>();
    accB.put(0, builder::appendAcc_B_0);
    accB.put(1, builder::appendAcc_B_1);
    accB.put(2, builder::appendAcc_B_2);
    accB.put(3, builder::appendAcc_B_3);
    consumers.put(ACC_B, accB);
  }

  public HashMap<Integer, Consumer<BigInteger>> get(String name){
    return consumers.get(name);
  }
}
