package linea.staterecover.plugin;

public class MyDataJ {
  public String description = "JavaPojo";
  public int someNumber = 20;
  public byte[] someBytes = new byte[]{0x1, 0x2};

  public MyDataJ() {
  }

  public MyDataJ(String name, int age, byte[] data) {
    this.description = name;
    this.someNumber = age;
    this.someBytes = data;
  }
}
