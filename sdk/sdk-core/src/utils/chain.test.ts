import { isLineaMainnet, isLineaSepolia, isMainnet, isSepolia } from "./chain";

describe("chain utils", () => {
  describe("isLineaSepolia", () => {
    it("returns true only for Linea Sepolia chain id", () => {
      expect(isLineaSepolia(59141)).toBe(true);
      expect(isLineaSepolia(59144)).toBe(false);
      expect(isLineaSepolia(1)).toBe(false);
      expect(isLineaSepolia(11155111)).toBe(false);
    });
  });

  describe("isLineaMainnet", () => {
    it("returns true only for Linea Mainnet chain id", () => {
      expect(isLineaMainnet(59144)).toBe(true);
      expect(isLineaMainnet(59141)).toBe(false);
      expect(isLineaMainnet(1)).toBe(false);
      expect(isLineaMainnet(11155111)).toBe(false);
    });
  });

  describe("isMainnet", () => {
    it("returns true only for Ethereum Mainnet chain id", () => {
      expect(isMainnet(1)).toBe(true);
      expect(isMainnet(59141)).toBe(false);
      expect(isMainnet(59144)).toBe(false);
      expect(isMainnet(11155111)).toBe(false);
    });
  });

  describe("isSepolia", () => {
    it("returns true only for Ethereum Sepolia chain id", () => {
      expect(isSepolia(11155111)).toBe(true);
      expect(isSepolia(1)).toBe(false);
      expect(isSepolia(59141)).toBe(false);
      expect(isSepolia(59144)).toBe(false);
    });
  });
});
