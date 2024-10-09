import { JsonRpcProvider, Wallet } from "ethers";
import { Config } from "./type";
import {
  DummyContract,
  DummyContract__factory,
  L2MessageService,
  L2MessageService__factory,
  LineaRollup,
  LineaRollup__factory,
} from "../src/typechain";
import { AccountManager } from "./accounts/account-manager";

export default class TestSetup {
  constructor(private readonly config: Config) {}

  public getL1Provider(): JsonRpcProvider {
    return new JsonRpcProvider(this.config.L1.rpcUrl.toString());
  }

  public getL2Provider(): JsonRpcProvider {
    return new JsonRpcProvider(this.config.L2.rpcUrl.toString());
  }

  public getL1ChainId(): number {
    return this.config.L1.chainId;
  }

  public getL2ChainId(): number {
    return this.config.L2.chainId;
  }

  public getShomeiEndpoint(): URL | undefined {
    return this.config.L2.shomeiEndpoint;
  }

  public getShomeiFrontendEndpoint(): URL | undefined {
    return this.config.L2.shomeiFrontendEndpoint;
  }

  public getSequencerEndpoint(): URL | undefined {
    return this.config.L2.sequencerEndpoint;
  }

  public getLineaRollupContract(signer?: Wallet): LineaRollup {
    const lineaRollup: LineaRollup = LineaRollup__factory.connect(
      this.config.L1.lineaRollupAddress,
      this.getL1Provider(),
    );

    if (signer) {
      return lineaRollup.connect(signer);
    }

    return lineaRollup;
  }

  public getL2MessageServiceContract(signer?: Wallet): L2MessageService {
    const l2MessageService: L2MessageService = L2MessageService__factory.connect(
      this.config.L2.l2MessageServiceAddress,
      this.getL2Provider(),
    );

    if (signer) {
      return l2MessageService.connect(signer);
    }

    return l2MessageService;
  }

  public async getL1DummyContract(signer?: Wallet): Promise<DummyContract> {
    const dummyContract = DummyContract__factory.connect(this.config.L1.dummyContractAddress, this.getL1Provider());

    if (signer) {
      return dummyContract.connect(signer);
    }

    return dummyContract;
  }

  public async getL2DummyContract(signer?: Wallet): Promise<DummyContract> {
    const dummyContract = DummyContract__factory.connect(this.config.L2.dummyContractAddress, this.getL2Provider());

    if (signer) {
      return dummyContract.connect(signer);
    }

    return dummyContract;
  }

  public getL1AccountManager(): AccountManager {
    return this.config.L1.accountManager;
  }

  public getL2AccountManager(): AccountManager {
    return this.config.L2.accountManager;
  }
}
