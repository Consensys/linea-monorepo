import { getSettings } from "@layerswap/widget";
import { LayerswapClientWrapper } from "./LayerswapClientWrapper";

export async function Widget() {
  const settings = await getSettings(
    "NDBxG+aon6WlbgIA2LfwmcbLU52qUL9qTnztTuTRPNSohf/VnxXpRaJlA5uLSQVqP8YGIiy/0mz+mMeZhLY4/Q",
  );

  return <LayerswapClientWrapper settings={settings} />;
}
