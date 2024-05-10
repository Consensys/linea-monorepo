import Link from 'next/link';

import PackageJSON from '@/../package.json';

export default function Footer() {
  return (
    <footer className="wrapper container flex flex-col md:flex-row justify-between p-6 gap-3 md:items-center">
      <div className="text-xs uppercase space-x-2 text-white">
        <span>@{new Date().getFullYear()}</span>
        <Link href={'https://linea.build/'} passHref target={'_blank'}>
          LINEA • A Consensys Formation
        </Link>
        <span className="text-transparent">v{PackageJSON.version}</span>
      </div>
      <div className="grid grid-flow-col gap-5 text-xs uppercase">
        <Link
          href={'https://docs.linea.build/use-mainnet/bridges-of-linea'}
          passHref
          target={'_blank'}
          className="link link-hover text-white"
        >
          Tutorial
        </Link>
        <Link href={'https://docs.linea.build/'} passHref target={'_blank'} className="link link-hover text-white">
          Documentation
        </Link>
        <Link
          href={'https://linea.build/terms-of-service'}
          passHref
          target={'_blank'}
          className="link link-hover text-white"
        >
          Terms of service
        </Link>
      </div>
    </footer>
  );
}
