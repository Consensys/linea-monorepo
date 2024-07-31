import BridgeLayout from "@/components/bridge/BridgeLayout";
import { Shortcut } from "@/models/shortcut";
import matter from "gray-matter";

async function getShortcuts() {
  const { data } = matter.read("src/data/shortcuts.md");
  return data as Shortcut[];
}

export default async function Home() {
  const shortcuts = await getShortcuts();

  return <BridgeLayout shortcuts={shortcuts} />;
}
