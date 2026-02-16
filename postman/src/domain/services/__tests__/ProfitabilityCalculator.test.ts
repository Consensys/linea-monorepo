import { ProfitabilityCalculator, ProfitabilityConfig } from "../ProfitabilityCalculator";

describe("ProfitabilityCalculator", () => {
  const defaultConfig: ProfitabilityConfig = {
    profitMargin: 1.0,
    maxClaimGasLimit: 500_000n,
    isPostmanSponsorshipEnabled: false,
    maxPostmanSponsorGasLimit: 250_000n,
  };

  let calculator: ProfitabilityCalculator;

  beforeEach(() => {
    calculator = new ProfitabilityCalculator(defaultConfig);
  });

  describe("hasZeroFee", () => {
    it("should return true when fee is 0 and profitMargin is nonzero", () => {
      expect(calculator.hasZeroFee(0n)).toBe(true);
    });

    it("should return false when fee is nonzero", () => {
      expect(calculator.hasZeroFee(1000n)).toBe(false);
    });

    it("should return false when profitMargin is 0 (disabled)", () => {
      const zeroMarginCalc = new ProfitabilityCalculator({ ...defaultConfig, profitMargin: 0 });
      expect(zeroMarginCalc.hasZeroFee(0n)).toBe(false);
    });
  });

  describe("calculateGasEstimationThreshold", () => {
    it("should return fee divided by gasLimit as a number", () => {
      expect(calculator.calculateGasEstimationThreshold(100_000n, 50n)).toBe(2000);
    });

    it("should truncate to integer for large values", () => {
      const result = calculator.calculateGasEstimationThreshold(2_000_000_000_000_000n, 1_000_000n);
      expect(result).toBe(2000000000);
    });
  });

  describe("getGasLimit", () => {
    it("should return gasLimit when within bounds", () => {
      expect(calculator.getGasLimit(400_000n)).toBe(400_000n);
    });

    it("should return null when gasLimit exceeds max", () => {
      expect(calculator.getGasLimit(600_000n)).toBeNull();
    });

    it("should return gasLimit when exactly at max", () => {
      expect(calculator.getGasLimit(500_000n)).toBe(500_000n);
    });
  });

  describe("isForSponsorship", () => {
    it("should return false when sponsorship is disabled", () => {
      expect(calculator.isForSponsorship(100_000n, true, false)).toBe(false);
    });

    it("should return false when gasLimit exceeds sponsor threshold", () => {
      const sponsorCalc = new ProfitabilityCalculator({
        ...defaultConfig,
        isPostmanSponsorshipEnabled: true,
      });
      expect(sponsorCalc.isForSponsorship(300_000n, true, false)).toBe(false);
    });

    it("should return true when has zero fee and gasLimit is within sponsor limit", () => {
      const sponsorCalc = new ProfitabilityCalculator({
        ...defaultConfig,
        isPostmanSponsorshipEnabled: true,
      });
      expect(sponsorCalc.isForSponsorship(200_000n, true, false)).toBe(true);
    });

    it("should return true when is underpriced and gasLimit is within sponsor limit", () => {
      const sponsorCalc = new ProfitabilityCalculator({
        ...defaultConfig,
        isPostmanSponsorshipEnabled: true,
      });
      expect(sponsorCalc.isForSponsorship(200_000n, false, true)).toBe(true);
    });

    it("should return false when message would be claimed regardless", () => {
      const sponsorCalc = new ProfitabilityCalculator({
        ...defaultConfig,
        isPostmanSponsorshipEnabled: true,
      });
      expect(sponsorCalc.isForSponsorship(200_000n, false, false)).toBe(false);
    });
  });

  describe("isL1UnderPriced", () => {
    it("should return true when actual cost exceeds message fee", () => {
      const result = calculator.isL1UnderPriced({
        gasLimit: 50_000n,
        messageFee: 1n,
        maxFeePerGas: 1_000_000_000n,
      });
      expect(result).toBe(true);
    });

    it("should return false when message fee covers the cost", () => {
      const result = calculator.isL1UnderPriced({
        gasLimit: 50_000n,
        messageFee: 2_000_000_000_000_000n,
        maxFeePerGas: 1_000_000_000n,
      });
      expect(result).toBe(false);
    });
  });

  describe("isL2UnderPriced", () => {
    it("should return true when message fee is insufficient for L2 cost", () => {
      const result = calculator.isL2UnderPriced({
        gasLimit: 50_000n,
        messageFee: 1n,
        compressedTransactionSize: 1000,
        blockExtraData: { version: 1, fixedCost: 1000, variableCost: 500, ethGasPrice: 1000000 },
      });
      expect(result).toBe(true);
    });

    it("should return false when message fee covers L2 cost", () => {
      const result = calculator.isL2UnderPriced({
        gasLimit: 50_000n,
        messageFee: 2_000_000_000_000_000n,
        compressedTransactionSize: 100,
        blockExtraData: { version: 1, fixedCost: 100, variableCost: 50, ethGasPrice: 1000000 },
      });
      expect(result).toBe(false);
    });
  });
});
