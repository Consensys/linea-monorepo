import ToChainDropdown from "./dropdowns/ToChainDropdown";

export function ToChain() {
  return (
    <div className="mb-4 flex items-center justify-between">
      <span className="text-white">To this network</span>
      <ToChainDropdown />
    </div>
  );
}
