import { hexlify, parseUnits, randomBytes } from "ethers";
import { ethers } from "hardhat";
import { BeaconBlockHeader, Validator, ValidatorContainer } from "./types";
import { SecretKey } from "@chainsafe/blst";
import { SSZMerkleTree } from "contracts/typechain-types";

export const randomInt = (max: number): number => Math.floor(Math.random() * max);

export const randomBytes32 = (): string => hexlify(randomBytes(32));

export const generateBeaconHeader = (stateRoot: string, slot?: number) => {
  return {
    slot: slot ?? randomInt(1743359),
    proposerIndex: randomInt(1337),
    parentRoot: randomBytes32(),
    stateRoot,
    bodyRoot: randomBytes32(),
  };
};

const ikm = Uint8Array.from(Buffer.from("test test test test test test test", "utf-8"));
const masterSecret = SecretKey.deriveMasterEip2333(ikm);
let secretIndex = 0;

export const generateValidator = (customWC?: string): Validator => {
  const secretKey = masterSecret.deriveChildEip2333(secretIndex++);

  return {
    blsPrivateKey: secretKey,
    container: {
      pubkey: secretKey.toPublicKey().toHex(true),
      withdrawalCredentials: customWC ?? hexlify(randomBytes32()),
      effectiveBalance: parseUnits(randomInt(32).toString(), "gwei"),
      slashed: false,
      activationEligibilityEpoch: BigInt(randomInt(343300)),
      activationEpoch: BigInt(randomInt(343300)),
      exitEpoch: BigInt(randomInt(343300)),
      withdrawableEpoch: BigInt(randomInt(343300)),
    },
  };
};

export const setBeaconBlockRoot = async (root: string) => {
  const systemAddress = "0xfffffffffffffffffffffffffffffffffffffffe";
  const initialBalance = 999999999999999999999999999n;
  await ethers.provider.send("hardhat_impersonateAccount", [systemAddress]);
  await ethers.provider.send("hardhat_setBalance", [systemAddress, ethers.toBeHex(initialBalance)]);
  const systemSigner = await ethers.getSigner(systemAddress);
  const BEACON_ROOTS = "0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02";
  const block = await systemSigner
    .sendTransaction({
      to: BEACON_ROOTS,
      value: 0,
      data: root,
    })
    .then((tx) => tx.getBlock());
  if (!block) throw new Error("invariant");
  return block.timestamp;
};

// Default mainnet values for validator state tree
export const prepareLocalMerkleTree = async (
  gIndex = "0x0000000000000000000000000000000000000000000000000096000000000028",
) => {
  const sszMerkleTree: SSZMerkleTree = await ethers.deployContract("SSZMerkleTree", [gIndex], {});
  const firstValidator = generateValidator();

  await sszMerkleTree.addValidatorLeaf(firstValidator.container);
  const validators: ValidatorContainer[] = [firstValidator.container];

  const firstValidatorLeafIndex = (await sszMerkleTree.leafCount()) - 1n;
  const gIFirstValidator = await sszMerkleTree.getGeneralizedIndex(firstValidatorLeafIndex);

  // compare GIndex.index()
  if (BigInt(gIFirstValidator) >> 8n !== BigInt(gIndex) >> 8n)
    throw new Error("Invariant: sszMerkleTree implementation is broken");

  const addValidator = async (validator: ValidatorContainer) => {
    await sszMerkleTree.addValidatorLeaf(validator);
    validators.push(validator);

    return {
      validatorIndex: validators.length - 1,
    };
  };

  const validatorAtIndex = (index: number) => {
    return validators[index];
  };

  const commitChangesToBeaconRoot = async (slot?: number) => {
    const beaconBlockHeader = generateBeaconHeader(await sszMerkleTree.getMerkleRoot(), slot);
    const beaconBlockHeaderHash = await sszMerkleTree.beaconBlockHeaderHashTreeRoot(beaconBlockHeader);
    return {
      childBlockTimestamp: await setBeaconBlockRoot(beaconBlockHeaderHash),
      beaconBlockHeader,
    };
  };

  const buildProof = async (validatorIndex: number, beaconBlockHeader: BeaconBlockHeader): Promise<string[]> => {
    const [validatorProof, stateProof, beaconBlockProof] = await Promise.all([
      sszMerkleTree.getValidatorPubkeyWCParentProof(validators[Number(validatorIndex)]).then((r) => r.proof),
      sszMerkleTree.getMerkleProof(BigInt(validatorIndex) + firstValidatorLeafIndex),
      sszMerkleTree.getBeaconBlockHeaderProof(beaconBlockHeader).then((r) => r.proof),
    ]);

    return [...validatorProof, ...stateProof, ...beaconBlockProof];
  };

  return {
    sszMerkleTree,
    gIFirstValidator,
    firstValidatorLeafIndex,
    get totalValidators(): number {
      return validators.length;
    },
    addValidator,
    validatorAtIndex,
    commitChangesToBeaconRoot,
    buildProof,
  };
};

export const ACTIVE_0X01_VALIDATOR = {
  blockRoot: "0xeb961eae87c614e11a7959a529a59db3c9d825d284dc30e0d12e43ba6daf4cca",
  gIFirstValidator: "0x0000000000000000000000000000000000000000000000000096000000000028",
  beaconBlockHeader: {
    slot: 22140000,
    proposerIndex: "1337",
    parentRoot: "0x8576d3eb5ef5b3a460b85e5493d0d0510b7d1a2943a4e51add5227c9d3bffa0f",
    stateRoot: "0xa802c5f4f818564a2774a19937fdfafc0241f475d7f28312c3609c6e5995d980",
    bodyRoot: "0xca4f98890bc98a59f015d06375a5e00546b8f2ac1e88d31b1774ea28d4b3e7d1",
  },
  witness: {
    validatorIndex: 129n,
    // Check this
    beaconBlockTimestamp: 42,
    validator: {
      pubkey: "0xaffd606a767b69df617169824f10baf992c407b059e18d8db2cf4f8cc7432dfe36477e94c9a14e87e1d13e5b22202b92",
      withdrawalCredentials: "0x010000000000000000000000b3e29c46ee1745724417c0c51eb2351a1c01cf36",
      effectiveBalance: 32000000000n,
      activationEligibilityEpoch: 10n,
      activationEpoch: 16n,
      exitEpoch: 18446744073709551615n,
      withdrawableEpoch: 18446744073709551615n,
      slashed: false,
    },
    // Should be proof for the Validator Container
    proof: [
      "0x0edd708eb0bcfc6f5c6c86579867e00b50938ae4656b566c1525385cf0e17d99",
      "0x4598d399eb5a13129ec7df15bf7f67ce62894f8dde56e20e01576d4b1c85d4c1",
      "0x250ab3879dec53f13d60b1fcba79ce887f2626863c4e1e678bbd0d0f1d1e9beb",
      "0xb71de3abf0dcb360fa49c4f512c5bfc5d513ecbd1dbc76c57ac29de173a65c23",
      "0x4eca0ad10870d33a08f0d4608d7ac506fa484876c884ed10887d46b5e1e694ca",
      "0xbed6b05d39560f0ebfd42fcaa33ddbdd9daf2efc3e0002fb327beae8088a4dad",
      "0xed185c1976880109a80043b032ff1bbedf74c501dc6e9a2785dabb438513a7dc",
      "0xc37a16340d7620558ce6048ce5913bfdfb4122b13185e99b6643a64d8f7e033c",
      "0xcbfafa05858b8aa3c8ff0312e723037a38985efea6528963b583aa7927907633",
      "0x73e9bb21041240d959b1fc6951a11b78cc5bf2955801663bf49cc397dfeef4dd",
      "0xe563a45ae8fd94663ad9b3cb0ac3e25c69827aa4b7ee39e5390b3bd1143e04bb",
      "0xbf617e7d2bd6314bb2fe5f525b95f42e629ee88a4d5075ed4841fe1165e2e633",
      "0x552485bed4a23db34514b36b78eca98f5933d4c874845042db9420c20e25bb19",
      "0xe72547a4304ee482db2d7fc7d8663b139889b1fc17d15cd29fd41fc8da7e4405",
      "0xc549ff35940423c68e7241a05ca1aa3af69fc4765232c04e575205fa57b8c3b1",
      "0x071f50189feb9a59c48c8d3a5b22bc42576ca404065aad4c3c34e662143fb35e",
      "0xfd1bd6dd9b73b8d2dfeebd1a0b45eb8903da8db371b72e3042f668a13ba0c43e",
      "0x4d8f1bb6f91e4161c180f9ecaad2f15c66cb6ec7f2cdedc3bcb93c274aad5fa2",
      "0x3b6350121535a5c9588561919e9772a84b549602680679766f6de26de339926e",
      "0xbb3b90338675e2e6bbb5057f462829234b29a9ab6ae17f1bb1e64c76bce64970",
      "0x3072ffd933e00085269e510dea720c2ac88a7853fdb3231ee5fb27e1e9a934cc",
      "0x8a8d7fe3af8caa085a7639a832001457dfb9128a8061142ad0335629ff23ff9c",
      "0xfeb3c337d7a51a6fbf00b9e34c52e1c9195c969bd4e7a0bfd51d5c5bed9c1167",
      "0xe71f0aa83cc32edfbefa9f4d3e0174ca85182eec9f3a09f6a6c0df6377a510d7",
      "0x31206fa80a50bb6abe29085058f16212212a60eec8f049fecb92d8c8e0a84bc0",
      "0x21352bfecbeddde993839f614c3dac0a3ee37543f9b412b16199dc158e23b544",
      "0x619e312724bb6d7c3153ed9de791d764a366b389af13c58bf8a8d90481a46765",
      "0x7cdd2986268250628d0c10e385c58c6191e6fbe05191bcc04f133f2cea72c1c4",
      "0x848930bd7ba8cac54661072113fb278869e07bb8587f91392933374d017bcbe1",
      "0x8869ff2c22b28cc10510d9853292803328be4fb0e80495e8bb8d271f5b889636",
      "0xb5fe28e79f1b850f8658246ce9b6a1e7b49fc06db7143e8fe0b4f2b0c5523a5c",
      "0x985e929f70af28d0bdd1a90a808f977f597c7c778c489e98d3bd8910d31ac0f7",
      "0xc6f67e02e6e4e1bdefb994c6098953f34636ba2b6ca20a4721d2b26a886722ff",
      "0x1c9a7e5ff1cf48b4ad1582d3f4e4a1004f3b20d8c5a2b71387a4254ad933ebc5",
      "0x2f075ae229646b6f6aed19a5e372cf295081401eb893ff599b3f9acc0c0d3e7d",
      "0x328921deb59612076801e8cd61592107b5c67c79b846595cc6320c395b46362c",
      "0xbfb909fdb236ad2411b4e4883810a074b840464689986c3f8a8091827e17c327",
      "0x55d8fb3687ba3ba49f342c77f5a1f89bec83d811446e1a467139213d640b6a74",
      "0xf7210d4f8e7e1039790e7bf4efa207555a10a6db1dd4b95da313aaa88b88fe76",
      "0xad21b516cbc645ffe34ab5de1c8aef8cd4e7f8d2b51e8e1456adc7563cda206f",
      "0x908a1e0000000000000000000000000000000000000000000000000000000000",
      "0xc6341f0000000000000000000000000000000000000000000000000000000000",
      "0x41e1bba24366f5cf6502295fea29d06396dd5d0031241811264f5212de2feb00",
      "0xc965aa7691807a7649ae6690e1a1d4871fc9ac6c3d96625fde856e152ea236d1",
      "0x85d66de5e59bf58a58e8e7334299a7fedcd87d23f97cf1579eae74e9fc0f0eaa",
      "0x5c76c5ff78ad80ef4bf12d6adb8df3b15f9c23798bda1aa1e110041254e76cce",
      "0x1c4a401fbd320fd7b848c9fc6118444da8797c2e41c525eb475ff76dbf44500b",
    ],
  },
};
