import BridgeIcon from "@/assets/icons/bridge.svg";
import TransactionsIcon from "@/assets/icons/transaction.svg";
import DocsIcon from "@/assets/icons/docs.svg";
import FaqIcon from "@/assets/icons/faq.svg";

export const MENU_ITEMS = [
  {
    title: "Bridge",
    href: "/",
    external: false,
    Icon: BridgeIcon,
  },
  {
    title: "Transactions",
    href: "/transactions",
    external: false,
    Icon: TransactionsIcon,
  },
  {
    title: "Docs",
    href: "https://docs.linea.build/",
    external: true,
    Icon: DocsIcon,
  },
  {
    title: "FAQ",
    href: "https://support.linea.build/hc/en-us/categories/13281330249371-FAQs",
    external: true,
    Icon: FaqIcon,
  },
];

export const NETWORK_ID_TO_NAME: Record<number, string> = {
  59144: "Linea",
  59141: "Linea Sepolia",
  1: "Ethereum",
  11155111: "Sepolia",
};
