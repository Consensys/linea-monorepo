import { useWeb3Modal } from "@web3modal/wagmi/react";

export default function ConnectButton() {
  const { open } = useWeb3Modal();
  return (
    <button
      id="wallet-connect-btn"
      className="btn btn-primary rounded-full text-sm font-semibold uppercase md:text-[0.9375rem]"
      onClick={() => open()}
    >
      Connect Wallet
    </button>
  );
}
