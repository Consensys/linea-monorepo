package net.consensys.linea.consensus;

import com.google.auto.service.AutoService;
import io.vertx.core.Vertx;
import net.consensys.linea.LineaBesuEngineBlockTagUpdater;
import net.consensys.linea.LineaL1FinalizationUpdaterService;
import net.consensys.linea.PluginCliOptions;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.ServiceManager;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;

@AutoService(BesuPlugin.class)
public class L1FinalizationTagUpdaterPlugin implements BesuPlugin {
  private final Logger log = LogManager.getLogger(L1FinalizationTagUpdaterPlugin.class);
	private static final String CLI_OPTIONS_PREFIX = "linea";
	private PluginCliOptions cliOptions = new PluginCliOptions();
	private Vertx vertx = Vertx.vertx();
	private LineaL1FinalizationUpdaterService service;
	private BlockchainService blockchainService;

  @Override
  public void register(final ServiceManager serviceManager) {
    try {
      log.info("Registering Linea plugin {}", this.getClass().getName());
      final PicoCLIOptions cmdlineOptions = serviceManager.getService(PicoCLIOptions.class)
        .orElseThrow(() -> new IllegalStateException("Failed to obtain PicoCLI options from ServiceManager"));
      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, cliOptions);

      blockchainService = serviceManager.getService(BlockchainService.class)
        .orElseThrow(() -> new RuntimeException("Failed to obtain BlockchainService from ServiceManager"));
    } catch (Exception e) {
      log.error("Halting Besu startup: exception in plugin registration: ", e);
      e.printStackTrace();
      // System.exit will cause besu to exit
      System.exit(1);
    }
  }

	@Override
	public void start() {
		service = new LineaL1FinalizationUpdaterService(
      vertx,
      cliOptions.getConfig(),
      new LineaBesuEngineBlockTagUpdater(blockchainService)
    );
		service.start();
	}

	@Override
	public void stop() {
		service.stop();
		vertx.close();
	}
}
