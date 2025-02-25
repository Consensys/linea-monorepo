import { useState } from "react";
import ConnectButton from "@/components/v2/connect-button";
import Link from "next/link";
import { useAccount } from "wagmi";
import Image from "next/image";
import CopyIcon from "@/assets/icons/copy.svg";
import CheckIcon from "@/assets/icons/check.svg";
import { formatAddress } from "@/utils/format";
import styles from "./header-connect.module.scss";
import { DynamicWidget, useDynamicContext } from "@dynamic-labs/sdk-react-core";

export default function Connect() {
  const [copied, setCopied] = useState<boolean>(false);
  const { handleLogOut } = useDynamicContext();

  const { address, isConnected } = useAccount();

  const handleCopy = async () => {
    if (!address) return;

    try {
      await navigator.clipboard.writeText(address);
      setCopied(true);

      setTimeout(() => {
        setCopied(false);
      }, 3000);
    } catch (err) {
      console.error("Failed to copy: ", err);
    }
  };

  // if (isConnected) {
  //   return (
  //     <div className={styles["wrapper"]}>
  //       <div className={styles["avatar"]}>
  //         <Image src="/images/logo/temp-user.svg" width={32} height={32} alt="user" />
  //       </div>
  //       <ul className={styles.submenu}>
  //         <li className={styles.submenuItem}>
  //           <div>
  //             <span>{formatAddress(address)}</span>
  //             {copied ? (
  //               <CheckIcon className={styles["check-icon"]} />
  //             ) : (
  //               <CopyIcon className={styles["copy-icon"]} onClick={handleCopy} />
  //             )}
  //           </div>
  //         </li>
  //         <li className={styles.submenuItem}>
  //           <div>
  //             <Link href="https://etherscan.io/" target="_blank" rel="noopenner noreferrer">
  //               Explorer
  //               <svg className={styles.newTab}>
  //                 <use href="#icon-new-tab" />
  //               </svg>
  //             </Link>
  //           </div>
  //         </li>
  //         <li className={styles.submenuItem}>
  //           <button onClick={() => handleLogOut()}>Disconnect wallet</button>
  //         </li>
  //       </ul>
  //     </div>
  //   );
  // }

  return <DynamicWidget innerButtonComponent={<>Connect</>} />;
}
