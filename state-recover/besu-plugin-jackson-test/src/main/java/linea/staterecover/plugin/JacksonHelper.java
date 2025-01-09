package linea.staterecover.plugin;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.module.kotlin.KotlinModule;

public class JacksonHelper {
  static {
    // tru to force loading
    try {
      var json1 = serializeToJson(false, new MyDataJ());
      System.out.println("jackson_java + javaClass: " + json1);
    } catch (Exception e) {
      System.out.println("jackson_java + javaClass error " + e.getMessage());
    }

    try {
      var json1 = serializeToJson(true, new MyDataJ());
      System.out.println("jackson_kotlin + javaClass: " + json1);
    } catch (Exception e) {
      System.out.println("jackson_kotlin + javaClass error " + e.getMessage());
    }

    try {
      var json1 = serializeToJson(false, new MyDataK());
      System.out.println("jackson_java + kotlinClass: " + json1);
    } catch (Exception e) {
      System.out.println("jackson_java + kotlinClass error " + e.getMessage());
    }

    try {
      var json1 = serializeToJson(true, new MyDataK());
      System.out.println("jackson_kotlin + kotlinClass: " + json1);
    } catch (Exception e) {
      System.out.println("jackson_kotlin + kotlinClass error " + e.getMessage());
    }
  }

  public static String someValue = "Hello :)";

  public static ObjectMapper buildObjectMapper(boolean registerKotlinModule) {
    ObjectMapper mapper = new ObjectMapper();
    if (registerKotlinModule) {
      mapper.registerModule(new KotlinModule.Builder().build());
    }
    return mapper;
  }

  public static String serializeToJson(boolean registerKotlinModule, Object obj) throws Exception {
    return buildObjectMapper(registerKotlinModule).writeValueAsString(obj);
  }
}
