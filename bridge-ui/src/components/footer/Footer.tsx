import Link from "next/link";

import PackageJSON from "@/../package.json";

export default function Footer() {
  return (
    <footer className="wrapper container flex flex-col justify-between gap-3 p-6 md:flex-row md:items-center">
      <div className="space-x-2 text-xs uppercase text-white">
        <span>@{new Date().getFullYear()}</span>
        <Link href={"https://linea.build/"} passHref target={"_blank"}>
          LINEA • A Consensys Formation
        </Link>
        <span className="text-transparent">v{PackageJSON.version}</span>
      </div>
      <div className="grid grid-flow-col gap-5 text-xs uppercase">
        <Link
          href={"https://docs.linea.build/use-mainnet/bridges-of-linea"}
          passHref
          target={"_blank"}
          className="link-hover link text-white"
        >
          Tutorial
        </Link>
        <Link href={"https://docs.linea.build/"} passHref target={"_blank"} className="link-hover link text-white">
          Documentation
        </Link>
        <Link
          href={"https://linea.build/terms-of-service"}
          passHref
          target={"_blank"}
          className="link-hover link text-white"
        >
          Terms of service
        </Link>
      </div>
    </footer>
  );
}
