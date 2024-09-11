import BridgeLayout from "@/components/bridge/BridgeLayout";

export default async function Home() {
  return (
    <div className="min-w-min max-w-lg md:m-auto md:mt-32">
      <h1 className="mb-6 text-4xl font-bold md:hidden">Bridge</h1>
      <BridgeLayout />
    </div>
  );
}
