import { describe, it, expect } from "@jest/globals";
import { ethers } from "ethers";
import { isValidUrl, sanitizeAddress, sanitizeETHThreshold, sanitizeHexString, sanitizeUrl } from "../cli.js";

describe("CLI", () => {
  describe("sanitizeAddress", () => {
    it("should throw an error when the input is not a valid Ethereum address", () => {
      const invalidAddress = "0x0a0e";
      expect(() => sanitizeAddress("Address")(invalidAddress)).toThrow(`Address is not a valid Ethereum address.`);
    });

    it("should return the input when it is a valid Ethereum address", () => {
      const address = ethers.hexlify(ethers.randomBytes(20));
      expect(sanitizeAddress("Address")(address)).toStrictEqual(address);
    });
  });

  describe("isValidUrl", () => {
    it("should return false when the url is not valid", () => {
      const invalidUrl = "www.test.com";
      expect(isValidUrl(invalidUrl, ["http:", "https:"])).toBeFalsy();
    });

    it("should return false when the url protocol is not allowed", () => {
      const invalidUrl = "tcp://test.com";
      expect(isValidUrl(invalidUrl, ["http:", "https:"])).toBeFalsy();
    });

    it("should return true when the url valid", () => {
      const url = "http://test.com";
      expect(isValidUrl(url, ["http:", "https:"])).toBeTruthy();
    });
  });

  describe("sanitizeUrl", () => {
    it("should throw an error when the input is not a valid url", () => {
      const invalidUrl = "www.test.com";
      expect(() => sanitizeUrl("Url", ["http:", "https:"])(invalidUrl)).toThrow(
        `Url, with value: ${invalidUrl} is not a valid URL`,
      );
    });

    it("should return the input when it is a valid url", () => {
      const url = "http://test.com";
      expect(sanitizeUrl("Url", ["http:", "https:"])(url)).toStrictEqual(url);
    });
  });

  describe("sanitizeHexString", () => {
    it("should throw an error when the input is not a hex string", () => {
      const invalidHexString = "0a1f";
      const expectedLength = 2;
      expect(() => sanitizeHexString("HexString", expectedLength)(invalidHexString)).toThrow(
        `HexString must be hexadecimal string of length ${expectedLength}.`,
      );
    });

    it("should throw an error when the input length is not equal to the expected length", () => {
      const hexString = "0x0a1f";
      const expectedLength = 4;
      expect(() => sanitizeHexString("HexString", expectedLength)(hexString)).toThrow(
        `HexString must be hexadecimal string of length ${expectedLength}.`,
      );
    });

    it("should return the input when it is a hex string", () => {
      const hexString = "0x0a1f";
      const expectedLength = 2;
      expect(sanitizeHexString("HexString", expectedLength)(hexString)).toStrictEqual(hexString);
    });
  });

  describe("sanitizeETHThreshold", () => {
    it("should throw an error when the input threshold is less than 1 ETH", () => {
      const invalidThreshold = "0.5";
      expect(() => sanitizeETHThreshold()(invalidThreshold)).toThrow("Threshold must be higher than 1 ETH");
    });

    it("should return the input when it is higher than 1 ETH", () => {
      const threshold = "2";
      expect(sanitizeETHThreshold()(threshold)).toStrictEqual(threshold);
    });
  });
});
