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
import { deployContract } from "../src/common/deployments";
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
    const l1DummyContractAddress = this.config.L1.dummyContractAddress;
    if (l1DummyContractAddress) {
      if (signer) {
        return DummyContract__factory.connect(l1DummyContractAddress, signer);
      }
      return DummyContract__factory.connect(l1DummyContractAddress, this.getL1Provider());
    }

    if (!signer) {
      throw new Error("signer is required to deploy a contract");
    }

    const dummyContract = (await deployContract(new DummyContract__factory(), signer)) as unknown as DummyContract;
    this.config.L1.dummyContractAddress = await dummyContract.getAddress();
    return dummyContract;
  }

  public async getL2DummyContract(signer?: Wallet): Promise<DummyContract> {
    const l2DummyContractAddress = this.config.L2.dummyContractAddress;
    if (l2DummyContractAddress) {
      if (signer) {
        return DummyContract__factory.connect(l2DummyContractAddress, signer);
      }
      return DummyContract__factory.connect(l2DummyContractAddress, this.getL2Provider());
    }

    if (!signer) {
      throw new Error("signer is required to deploy a contract");
    }

    const dummyContract = (await deployContract(new DummyContract__factory(), signer)) as unknown as DummyContract;
    this.config.L2.dummyContractAddress = await dummyContract.getAddress();
    return dummyContract;
  }

  public getL1AccountManager(): AccountManager {
    return this.config.L1.accountManager;
  }

  public getL2AccountManager(): AccountManager {
    return this.config.L2.accountManager;
  }
}
