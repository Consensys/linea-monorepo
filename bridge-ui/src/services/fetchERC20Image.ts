import axios, { AxiosResponse } from "axios";
import log from "loglevel";

interface CoinGeckoToken {
  id: string;
  symbol: string;
  name: string;
}

interface CoinGeckoTokenDetail {
  image: {
    small: string;
  };
}

async function fetchERC20Image(name: string) {
  try {
    if (!name) {
      throw new Error("Name is required");
    }

    const coinsResponse: AxiosResponse<CoinGeckoToken[]> = await axios.get(
      "https://api.coingecko.com/api/v3/coins/list",
    );
    const coin = coinsResponse.data.find((coin: CoinGeckoToken) => coin.name === name);

    if (!coin) {
      throw new Error("Coin not found");
    }

    const coinId = coin.id;
    const coinDataResponse: AxiosResponse<CoinGeckoTokenDetail> = await axios.get(
      `https://api.coingecko.com/api/v3/coins/${coinId}`,
    );

    if (!coinDataResponse.data.image.small) {
      throw new Error("Image not found");
    }

    const image = coinDataResponse.data.image.small;
    return image.split("?")[0];
  } catch (error) {
    log.warn(error);
    return "/images/logo/noTokenLogo.svg";
  }
}

export default fetchERC20Image;
