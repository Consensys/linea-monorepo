import React, { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import classNames from 'classnames';

const STORAGE_KEY = 'firstTime';
export default function TermsModal() {
  const termsModalRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const [videoEnabled, setVideoEnabled] = useState(false);

  const isFirstTime = () => {
    if (localStorage.getItem(STORAGE_KEY)) return false;
    return true;
  };

  const agreeToTerms = () => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(true));
    setOpen(false);
    setVideoEnabled(false);
  };

  useEffect(() => {
    if (window && isFirstTime()) {
      setTimeout(() => {
        setOpen(true);
        setVideoEnabled(true);
      }, 1000);
    }
  }, []);

  return (
    <div
      ref={termsModalRef}
      id="terms_modal"
      className={classNames(
        'p-4 fixed right-2 left-2 md:left-auto md:right-5 md:max-w-[20rem] bg-white rounded text-black z-50 transition-all duration-500',
        !open ? 'invisible -bottom-full' : 'visible bottom-2 md:bottom-16',
      )}
    >
      {videoEnabled && (
        <>
          <div className="font-medium text-lg">Tutorial</div>
          <div className="pb-2">
            <iframe
              className="aspect-video w-full"
              src="https://www.youtube.com/embed/V4DflPkxqE8"
              title="YouTube video player"
              frameBorder="0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
              allowFullScreen
            ></iframe>
          </div>
        </>
      )}
      <div className="font-medium text-lg">Terms of Use</div>
      <div className="leading-relaxed text-xs mb-2">
        Linea Mainnet is in Alpha — click{' '}
        <Link href="https://docs.linea.build/risk-disclosures" target="_blank" className="font-bold mr-1">
          here
        </Link>
        to learn more about the risks and security measures. I agree and accept that my use of the services is subject
        to the Linea Terms of Use, which contains an arbitration provision and class action waiver{' '}
        <Link href="https://linea.build/terms-of-service" target="_blank" className="font-bold ml-1">
          (Terms of Service | Linea )
        </Link>{' '}
      </div>
      <button
        id="agree-terms-btn"
        onClick={agreeToTerms}
        type="button"
        className="btn font-medium w-full rounded-full uppercase btn-sm btn-primary mt-3"
      >
        Got It
      </button>
    </div>
  );
}
