package net.consensys.linea.zktracer.module.alu.mod;

import java.math.BigInteger;
import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.Bytes16;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

public class ModData {
  private final OpCode opCode;
  private final boolean oli;
  private final Bytes32 arg1;
  private final Bytes32 arg2;
  private final Bytes16 arg1Hi;
  private final Bytes16 arg1Lo;
  private final Bytes16 arg2Hi;
  private final Bytes16 arg2Lo;

/*  private final Bytes32 arg1Bytes;
  private final Bytes32  arg2Bytes;*/
  private final Bytes32 resBytes;
  private final Bytes16 resHi;
  private final Bytes16 resLo;

/////////
  //
  private BaseTheta A_Bytes;
  private  BaseTheta B_Bytes;
  private  BaseTheta Q_Bytes;
  private  BaseTheta R_Bytes;

  private final boolean[] cmp1 = new boolean[4];
  private final boolean[]cmp2 = new boolean[4];

  public ModData(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
    this.opCode = opCode;
    this.oli = arg2.isZero();
    this.resBytes = getRes(opCode, arg1,arg2);

    this.resHi = Bytes16.wrap(resBytes.slice(0,16));
    this.resLo = Bytes16.wrap(resBytes.slice(16));

    this.arg1Hi = Bytes16.wrap(arg1.slice(0, 16));
    this.arg1Lo = Bytes16.wrap(arg1.slice(16));
    this.arg2Hi = Bytes16.wrap(arg2.slice(0, 16));
    this.arg2Lo = Bytes16.wrap(arg2.slice(16));

    this.arg1 = arg1;
    this.arg2 = arg2;

    if(!this.oli){
      BigInteger a = absoluteValueIfSignedInst(arg1);
      BigInteger b = absoluteValueIfSignedInst(arg2);
      BigInteger q = a.divide(b);
      BigInteger r = a.mod(b);
      this.A_Bytes = new BaseTheta(a);
      this.B_Bytes = new BaseTheta(b);
      this.Q_Bytes = new BaseTheta(q);
      this.R_Bytes = new BaseTheta(r);;
      this.setCmp12();
    }
  }

/*  private BigInteger a(int k) {
    checkElementIndex(k, 4);
    return new BigInteger(A_Bytes.getBytes()[k]);
  }*/

  private BigInteger b(int k) {
    checkElementIndex(k, 4);
    return new BigInteger(B_Bytes.get(k).toArray());
  }

/*  private BigInteger q(int k) {
    checkElementIndex(k, 4);
    return new BigInteger(Q_Bytes.getBytes()[k]);
  }*/

  private BigInteger r(int k) {
    checkElementIndex(k, 4);
    return new BigInteger(R_Bytes.get(k).toArray());
  }

/*  private BigInteger h(int k) {
    checkElementIndex(k, 3);
    return new BigInteger(H_Bytes.getBytes()[k]);
  }*/

  private void setCmp12() {
    for (int k = 0; k < 4; k++) {
      cmp1[k] = b(k).compareTo(r(k)) > 0;
      cmp2[k] = b(k).compareTo(r(k)) == 0;
    }
  }


  private BigInteger absoluteValueIfSignedInst (Bytes32 arg){
    if (isSigned()) {
      return arg.toUnsignedBigInteger().abs();
    }
    return arg.toUnsignedBigInteger();
  }

  public static Bytes32 getRes(OpCode op, Bytes32 arg1, Bytes32 arg2) {
    BigInteger res;
    switch (op) {
      case DIV -> res = arg1.toUnsignedBigInteger().divide(arg2.toUnsignedBigInteger());
      case SDIV ->res = arg1.toBigInteger().divide(arg2.toBigInteger());
      case MOD -> res =arg1.toUnsignedBigInteger().mod(arg2.toUnsignedBigInteger());
      case SMOD -> res = arg1.toBigInteger().mod(arg2.toBigInteger());
      default ->
        throw new RuntimeException("Modular arithmetic was given wrong opcode");
    };
    return Bytes32.leftPad(Bytes.of(res.toByteArray()));
  }



  public OpCode getOpCode() {
    return opCode;
  }

  public boolean isOli() {
    return oli;
  }

  public Bytes32 getArg1() {
    return arg1;
  }

  public Bytes32 getArg2() {
    return arg2;
  }

  public Bytes16 getArg1Hi() {
    return arg1Hi;
  }

  public Bytes16 getArg1Lo() {
    return arg1Lo;
  }

  public Bytes16 getArg2Hi() {
    return arg2Hi;
  }

  public Bytes16 getArg2Lo() {
    return arg2Lo;
  }


  public Bytes32 getResBytes() {
    return resBytes;
  }

  public Bytes16 getResHi() {
    return resHi;
  }

  public Bytes16 getResLo() {
    return resLo;
  }

  public BaseTheta getA_Bytes() {
    return A_Bytes;
  }

  public BaseTheta getB_Bytes() {
    return B_Bytes;
  }

  public BaseTheta getQ_Bytes() {
    return Q_Bytes;
  }

  public BaseTheta getR_Bytes() {
    return R_Bytes;
  }

/*  public BaseTheta getH_Bytes() {
    return H_Bytes;
  }*/

/*  public BaseTheta getDELTA_Bytes() {
    return DELTA_Bytes;
  }

  public boolean[] getMsb1() {
    return msb1;
  }

  public boolean[] getMsb2() {
    return msb2;
  }*/

  public boolean[] getCmp1() {
    return cmp1;
  }

  public boolean[] getCmp2() {
    return cmp2;
  }

  public boolean isSigned() {
    return this.opCode == OpCode.SDIV || this.opCode == OpCode.SMOD;
  }

  public boolean isDiv() {
    return this.opCode == OpCode.DIV || this.opCode == OpCode.SDIV;
  }

  static void checkElementIndex(int index, int size) {
    if (index < 0 || index >= size) {
      throw new IndexOutOfBoundsException("index is out of bounds");
    }
  }
}
