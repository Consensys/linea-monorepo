import Image from "next/image";
import styles from "./network-details.module.scss";

type Props = {
  name: string;
  image: string;
};

export default function NetworkDetails({ name, image }: Props) {
  return (
    <div className={styles.wrapper}>
      <Image className={styles.image} src={image} alt="" width={24} height={24} />
      <p className={styles.name}>{name}</p>
    </div>
  );
}
