import { ethers } from "ethers";
import { encodeData } from "./encoding";

export const generateKeccak256 = (types: string[], values: unknown[], opts: { encodePacked?: boolean }) =>
  ethers.keccak256(encodeData(types, values, opts.encodePacked));
