import Image from "next/image";

type HeaderLogoProps = {
  pathname: string;
};

function formatPath(pathname: string): string {
  switch (pathname) {
    case "/":
    case "":
      return "Bridge";
    case "/transactions":
      return "Transactions";
    case "/faq":
      return "FAQ";
    default:
      return "";
  }
}

export const HeaderLogo: React.FC<HeaderLogoProps> = ({ pathname }) => (
  <div className="flex-1">
    <Image
      src={"/images/logo/linea.svg"}
      alt="Linea logo"
      width={0}
      height={0}
      className="md:hidden"
      priority
      style={{ width: "auto", height: "auto" }}
    />
    <h1 className="hidden text-4xl md:flex">{formatPath(pathname)}</h1>
  </div>
);
