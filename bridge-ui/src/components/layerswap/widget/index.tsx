import "@layerswap/widget/index.css";
import { Swap, LayerswapProvider, GetSettings } from '@layerswap/widget'
import CustomHooks from "./custom-hooks";

export async function Widget() {
    const settings = await GetSettings()
    return (
        <LayerswapProvider
            integrator='linea'
            themeData={themeData}
            settings={settings}
        >
            <CustomHooks>
                <Swap
                    featuredNetwork={{
                        initialDirection: 'from',
                        network: 'ETHEREUM_MAINNET',
                        oppositeDirectionOverrides: 'onlyExchanges',
                    }}
                />
            </CustomHooks>
        </LayerswapProvider>
    );
}

const themeData = {
    placeholderText: '134, 134, 134',
    actionButtonText: '255, 255, 255',
    buttonTextColor: '17, 17, 17',
    logo: '255, 0, 147',
    footerLogo: 'none',
    primary: {
        DEFAULT: '97, 26, 239',
        '50': '248, 200, 220',
        '100': '246, 182, 209',
        '200': '241, 146, 186',
        '300': '237, 110, 163',
        '400': '232, 73, 140',
        '500': '97, 26, 239',
        '600': '166, 51, 94',
        '700': '136, 17, 67',
        '800': '147, 8, 99',
        '900': '181, 144, 255',
        'text': '18, 18, 18',
        'textMuted': '86, 97, 123',
    },
    secondary: {
        DEFAULT: '248, 247, 241',
        '50': '49, 60, 155',
        '100': '46, 59, 147',
        '200': '134, 134, 134',
        '300': '139, 139, 139',
        '400': '220, 219, 214',
        '500': '228, 227, 219',
        '600': '240, 240, 235',
        '700': '248, 247, 241',
        '800': '243, 244, 246',
        '900': '255, 255, 255',
        '950': '255, 255, 255',
        'text': '82, 82, 82',
    }
}