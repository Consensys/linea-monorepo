"use client";

import SearchIcon from "@/assets/icons/search.svg";
import Modal from "@/components/v2/modal";
import { useDevice } from "@/hooks/useDevice";
import { useState } from "react";
import NetworkDetails from "./network-details";
import styles from "./select-network.module.scss";

interface Props {
  isModalOpen: boolean;
  onCloseModal: () => void;
}

const ListNetwork = [
  {
    name: "Aleph Zero",
  },
  {
    name: "Arbitrum One",
  },
  {
    name: "Base",
  },
  {
    name: "Blast",
  },
  {
    name: "Ethereum Mainnet",
  },
  {
    name: "Ink",
  },
  {
    name: "Linea",
  },
  {
    name: "Lisk",
  },
  {
    name: "Mode",
  },
  {
    name: "Optimism",
  },
  {
    name: "Polygon Network",
  },
  {
    name: "Redstone",
  },
  {
    name: "Scroll",
  },
  {
    name: "World Chain",
  },
  {
    name: "Zora",
  },
  {
    name: "Soneium",
  },
  {
    name: "ZkSync Era",
  },
];

export default function SelectNetwork({ isModalOpen, onCloseModal }: Props) {
  const [filteredNetworks, setFilteredNetworks] = useState(ListNetwork);
  const { isMobile } = useDevice();

  const [searchQuery, setSearchQuery] = useState("");

  return (
    <Modal title="Select a token" isOpen={isModalOpen} onClose={onCloseModal} isDrawer={isMobile}>
      <div className={styles["modal-inner"]}>
        <div className={styles["input-wrapper"]}>
          <SearchIcon />
          <input
            type="text"
            placeholder="Search by name"
            onChange={({ target: { value } }) => setSearchQuery(value)}
            value={searchQuery}
          />
        </div>
        <div className={styles["list-network"]}>
          {filteredNetworks.length > 0 ? (
            filteredNetworks.map((network, index: number) => {
              return <NetworkDetails key={index} name={network.name} image="/images/logo/ethereum-rounded.svg" />;
            })
          ) : (
            <p>No networks found</p>
          )}
        </div>
      </div>
    </Modal>
  );
}
