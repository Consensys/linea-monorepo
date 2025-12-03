package linea.plugin.acc.test.rpc.linea;

import org.junit.jupiter.api.Test;

public class CreateAccountsTest extends AbstractSendBundleTest {
  @Test
  public void createAccounts() throws Exception {
    final var accountsPool = createAccounts(1, 1);
  }
}
