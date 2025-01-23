export type AssetType = ContentfulFields<{
  description?: string;
  file: {
    url: string;
    details: {
      size?: number;
      image: {
        width: number;
        height: number;
      };
    };
    contentType?: string;
    fileName?: string;
  };
  title: string;
}>;

export type CtaProps = ContentfulFields<{
  customClass?: string;
  type: "button" | "link";
  icon?: AssetType;
  label: string;
  openNewTab?: boolean;
  file?: AssetType;
  url?: string;
  backgroundColor?: string;
  variant?: "contained" | "outlined" | "underlined";
  hubspotForm?: HubspotFormType;
  heading?: string;
}>;

export type LinkBlock = {
  __id: string;
  name: string;
  label: string;
  url: string;
  external: boolean;
  icon: AssetType;
  submenusLeft: LinkBlock[];
  submenusRight: LinkBlock;
};

export type FooterProps = {
  logo: AssetType;
  copyright: string;
  slogan?: string;
  privacyLinks: CtaProps[];
  socialLinks: CtaProps[];
  menus?: LinkBlock[];
};

export type ImagesCollection = ContentfulFields<{
  images: AssetType[];
}>;

export type SEOProps = {
  title: string;
  description: string;
  image: AssetType;
  canonical: string;
  keywords: string[];
  noIndex: boolean;
  noFollow: boolean;
};

export type BannerProps = {
  title: string;
  content: string;
};

export type PopupProps = ContentfulFields<{
  name: string;
  content: string;
  cta: CtaProps;
}>;

export type PageLayoutProps = {
  name: string;
  seo: SEOProps;
  banner: BannerProps;
  header: HeaderProps;
  hero: HeroProps;
  sections: ModuleType[];
  footer: FooterProps;
  theme?: "v1" | "v2";
};

export type ModuleType = SectionType | FeatureType | PopupProps;

export type FeatureType = ContentfulFields<{
  moduleId: string;
  name: string;
  title: string;
  description: string;
  eyebrow: string;
  overlayLeft: AssetType;
  overlayRight: AssetType;
  featureImage: AssetType;
  images: AssetType[];
  contentCol: number;
  reverse: boolean;
  descriptionAsFeature: boolean;
  component: string;
  backgroundColor: string;
  backgroundImage: AssetType;
  ctas: CtaProps[];
  items: ContentfulFields<CardType>[];
  fullWidth: boolean;
  itemsCenter: boolean;
}>;

export type ContentfulFields<T> = T & {
  __id?: string;
  __typename?: string;
};

export type SpectrumType = ContentfulFields<{
  title: string;
  color: string;
  items: CardType[];
}>;

export type ItemType =
  | CardType
  | VideoType
  | WaveCardType
  | ModuleListDataType
  | BrandAssetsProps
  | SpectrumType
  | ImagesCollection
  | HubspotFormType;

export type SectionType<T = ItemType> = ContentfulFields<{
  name: string;
  title: string;
  description: string;
  eyebrow: string;
  overlayLeft: AssetType;
  overlayRight: AssetType;
  background: AssetType;
  layout: string;
  component: string;
  relative: boolean;
  backgroundColor: string;
  typography: "h1" | "h2";
  ctas: CtaProps[];
  items: ContentfulFields<T>[];
  col: number;
  colMd: number;
  colLg: number;
  colXl: number;
  centerOnMobile: boolean;
  slideOnMobile: boolean;
  customData: Record<string, any>;
  idList?: string[];
}>;

export type ItemWrapperProps<T = ItemType> = {
  item: T;
  backgroundColor: string;
  className?: string;
  index: number;
};

export type WaveCardType = ContentfulFields<{
  name: string;
  title: string;
  description: string;
  url: string;
  partners: PartnerTradingCardType[];
  thumbnail: AssetType;
}>;

export type PartnerTradingCardType = ContentfulFields<{
  media: AssetType;
  url: string;
}>;

export type LegalType = {
  frontMatter: {
    canonical: string;
    date: string;
    slug: string;
    title: string;
  };
  content: string;
};

export type VideoType = ContentfulFields<{
  name: string;
  vimeoUrl: string;
  youtubeUrl: string;
  file: AssetType;
  placeholder: AssetType;
  options: Record<string, any>;
}>;

export type HubspotFormType = ContentfulFields<{
  targetId: string;
  formId: string;
  region?: string;
  portalId: string;
  isDarkTheme?: boolean;
  fullWidth?: boolean;
  v2?: boolean;
}>;

export type CardCategory = "Bridges and Onramps" | "DeFi" | "Gaming" | "NFTs" | "Infra & Tools";

export enum Theme {
  "default" = "default",
  "navy" = "navy",
  "cyan" = "cyan",
  "indigo" = "indigo",
  "tangerine" = "tangerine",
}
