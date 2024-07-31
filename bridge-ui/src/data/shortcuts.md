---
- title: 'Bridge and ReStake'
  description: 'Bridge ETH from Mainnet to Linea.'
  logo: '/images/logo/linea-logo.svg'
  ens_name: 'bridge.onlinea.eth'

- title: 'Bridge and ReStake'
  description: 'Bridge ETH from Mainnet, keep 0.005 in ETH and swap the remaining for wstETH.'
  logo: '/images/logo/lido.svg'
  ens_name: 'lido.onlinea.eth'

- title: 'Bridge and ReStake'
  description: 'Bridge ETH from Mainnet, keep 0.005 in ETH and restake the remaining in ezETH.'
  logo: '/images/logo/renzo.svg'
  ens_name: 'renzo.onlinea.eth'

# - title: 'Bridge and ReStake'
#   description: 'Low gas fees and low latency with high throughput backed by the security of Ethereum.'
#   logo: '/images/logo/renzo.svg'
#   ens_name: 'onrenzo.linea.eth'
---

Each shortcut is a block with the following structure:

```
- title: ...
  description: ...
  logo: ...
  ens_name: ...
```

Copy and paste the block and change the values to create a new shortcut.

> **_NOTE:_** If you use internal image for `logo`, please make sure to add the image to this repository at `bridge-ui/public/images/...`
