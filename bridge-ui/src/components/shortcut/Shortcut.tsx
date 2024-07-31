import Image from "next/image";
import React, { useRef, useState } from "react";
import { Swiper, SwiperRef, SwiperSlide } from "swiper/react";
import { motion } from "framer-motion";

import "swiper/css";
import classNames from "classnames";
import { Shortcut as ShortcutInterface } from "@/models/shortcut";
import { useShortcuts } from "../bridge/BridgeLayout";

const poweredBy = [
  {
    label: "Onthis.xyz",
    url: "https://onthis.xyz/",
    logo: "/images/logo/onthis.svg",
  },
  {
    label: "Across",
    url: "https://across.to/",
    logo: "/images/logo/across.svg",
  },
];

const Shortcut: React.FC = () => {
  const [progress, setProgress] = useState(0);
  const sliderRef = useRef<SwiperRef>(null);

  const shortcuts = useShortcuts();

  const handleClickNext = () => {
    sliderRef.current?.swiper.slideNext();
  };

  const handleClickPrev = () => {
    sliderRef.current?.swiper.slidePrev();
  };

  return (
    <div className="shortcut mt-32 w-full">
      <div className="flex items-center justify-between text-white">
        <h2 className="text-3xl md:text-[2.625rem]">Shortcuts</h2>
        <a
          href="https://docs.linea.build/developers/tooling/cross-chain/shortcuts"
          target="_blank"
          className="underline decoration-primary underline-offset-8"
        >
          How does it work?
        </a>
      </div>

      {shortcuts.length > 0 && (
        <Swiper
          slidesPerView={1}
          slidesPerGroup={1}
          spaceBetween={20}
          mousewheel={{
            forceToAxis: true,
          }}
          breakpoints={{
            640: {
              slidesPerView: 1,
            },
            768: {
              slidesPerView: 2,
            },
            1024: {
              slidesPerView: 3,
            },
            1440: {
              slidesPerView: 4,
            },
          }}
          ref={sliderRef}
          onSlideChange={(swiper) => {
            setProgress(swiper.progress);
          }}
          className="mt-6 size-full"
        >
          {shortcuts.map((shortcut, index) => (
            <SwiperSlide key={index}>
              <ShortcutCard item={shortcut} className="" index={index} />
            </SwiperSlide>
          ))}
        </Swiper>
      )}

      <div className="mt-8 flex flex-wrap items-center justify-center gap-8 text-white md:justify-between">
        <div className="flex items-center gap-2.5">
          <span>Powered by </span>
          {poweredBy.map((item, index) => (
            <a href={item.url} key={index} target="_blank" className="flex items-center gap-x-1.5 hover:underline">
              <Image
                src={item.logo}
                alt={item.label}
                className="size-6"
                width={0}
                height={0}
                style={{ width: "24px", height: "auto" }}
              />
              <span>{item.label}</span>
            </a>
          ))}
        </div>
        <div className="mr-1 flex gap-2">
          <button
            onClick={handleClickPrev}
            className="btn size-12 rounded bg-primary px-0 text-black hover:bg-primary disabled:bg-cardBg disabled:text-card"
            disabled={progress === 0}
          >
            <svg
              className="rotate-180"
              xmlns="http://www.w3.org/2000/svg"
              width="40"
              height="40"
              viewBox="0 0 40 40"
              fill="none"
            >
              <path
                d="M15.7815 30.2967L24.5276 21.5506C25.2006 20.8776 25.2006 19.7864 24.5276 19.1134L15.7815 10.3672"
                stroke="currentColor"
                strokeWidth="2"
              />
            </svg>
          </button>
          <button
            onClick={handleClickNext}
            className="btn size-12 rounded bg-primary px-0 text-black hover:bg-primary disabled:bg-cardBg disabled:text-card"
            disabled={progress === 1}
          >
            <svg className="" xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 40 40" fill="none">
              <path
                d="M15.7815 30.2967L24.5276 21.5506C25.2006 20.8776 25.2006 19.7864 24.5276 19.1134L15.7815 10.3672"
                stroke="currentColor"
                strokeWidth="2"
              />
            </svg>
          </button>
        </div>
      </div>

      <p className="my-14 max-w-xl">
        Note: *.onlinea.eth shortcuts are curated by Linea, but provided by Ecosystem Partners and Community. They are
        not canonical solutions and they include additional fees collected only by involved partners.
      </p>
    </div>
  );
};

const ShortcutCard: React.FC<{
  item: ShortcutInterface;
  className?: string;
  index?: number;
}> = ({ item, className, index = 0 }) => {
  const { title, description, logo, ens_name } = item;
  const [buttonClicked, setButtonClicked] = useState<boolean>(false);

  const onButtonClick = () => {
    setButtonClicked(true);
    window.navigator.clipboard.writeText(ens_name);
    setTimeout(() => setButtonClicked(false), 3000);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay: index * 0.1 }}
    >
      <div className={classNames("shortcut-card min-h-64 mt-2 mr-1", className)}>
        {logo && <Image src={logo} alt={title} width={120} height={32} className="mb-3 h-8" />}
        <div className="text-gray my-4 line-clamp-[10] text-lg md:mb-3.5">{description}</div>
        {ens_name && (
          <button
            onClick={onButtonClick}
            className={classNames(
              "btn w-fit btn-custom bg-white text-black rounded-full mt-auto hover:bg-primary text-sm md:text-[0.9375rem] font-medium uppercase",
              { "opacity-60 normal-case": buttonClicked },
            )}
          >
            {buttonClicked ? "Copied to Clipboard" : ens_name}
          </button>
        )}
      </div>
    </motion.div>
  );
};

export default Shortcut;
