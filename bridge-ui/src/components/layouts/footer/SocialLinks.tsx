import Image from "next/image";
import Link from "next/link";

export const SocialLinks = () => (
  <div className="flex items-center gap-3 py-2">
    <Link href={"https://linea.mirror.xyz/"} passHref target="_blank" rel="noopener noreferrer">
      <Image src={"/images/logo/sidebar/linea-mirror.svg"} alt="Linea logo" width={19} height={18} />
    </Link>
    <Link href={"https://x.com/LineaBuild"} passHref target="_blank" rel="noopener noreferrer">
      <Image src={"/images/logo/sidebar/twitter.svg"} alt="Linea logo" width={20} height={16} />
    </Link>
    <Link href={"https://www.youtube.com/@LineaBuild"} passHref target="_blank" rel="noopener noreferrer">
      <Image src={"/images/logo/sidebar/youtube.svg"} alt="Linea logo" width={24} height={16} />
    </Link>
    <Link href={"https://discord.gg/linea"} passHref target="_blank" rel="noopener noreferrer">
      <Image src={"/images/logo/sidebar/discord.svg"} alt="Linea logo" width={22} height={16} />
    </Link>
  </div>
);
