<h1 align="center">
  <a href="https://t.me/pntmsurfbot">rule-set-ru</a>

</h1>

<p align="center">
  <img src="assets/images/shinobu.jpeg" alt="waifu" width="100%">
</p>

<p align="center">
  <b>Современный и обширный набор правил Mihomo, SingBox и Xray-Core</b>
</p>

## Подключение

### HAPP
```json
{
  "Name": "Phantom Surf Only Ru Routing",
  "GlobalProxy": "true",
  "RouteOrder": "block-proxy-direct",
  "RemoteDNSType": "DoH",
  "RemoteDNSDomain": "https://8.8.8.8/dns-query",
  "RemoteDNSIP": "8.8.8.8",
  "DomesticDNSType": "DoH",
  "DomesticDNSDomain": "https://77.88.8.8/dns-query",
  "DomesticDNSIP": "77.88.8.8",
  "Geoipurl": "https://github.com/pntmsurf/rule-set-ru/releases/download/latest/geoip.dat",
  "Geositeurl": "https://github.com/pntmsurf/rule-set-ru/releases/download/latest/geosite.dat",
  "LastUpdated": "1785305201",
  "DnsHosts": {
    "lkfl2.nalog.ru": "213.24.64.175",
    "lknpd.nalog.ru": "213.24.64.181"
  },
  "DirectSites": [],
  "DirectIp": [],
  "ProxySites": [
    "geosite:ru-ads",
    "geosite:ru-boards",
    "geosite:ru-dev",
    "geosite:ru-finance",
    "geosite:ru-food",
    "geosite:ru-forums",
    "geosite:ru-games",
    "geosite:ru-gov",
    "geosite:ru-job",
    "geosite:ru-map",
    "geosite:ru-markets",
    "geosite:ru-media",
    "geosite:ru-messengers",
    "geosite:ru-movies",
    "geosite:ru-music",
    "geosite:ru-neuro",
    "geosite:ru-porn",
    "geosite:ru-social",
    "geosite:ru-telecom",
    "geosite:ru-videos",
    "geosite:ru-yandex",
    "geosite:youtube",
    "geosite:gemini"
  ],
  "ProxyIp": [],
  "BlockSites": [],
  "BlockIp": [],
  "DomainStrategy": "IPIfNonMatch",
  "FakeDNS": "false",
  "UseChunkFiles": "false"
}
```

## Использование

### Xray
```json
{
 "routing": {
  "rules": [
   {
    "domain": [
     "geosite:ru-ads",
     "geosite:ru-boards",
     "geosite:ru-dev",
     "geosite:ru-finance",
     "geosite:ru-food",
     "geosite:ru-forums",
     "geosite:ru-games",
     "geosite:ru-gov",
     "geosite:ru-job",
     "geosite:ru-map",
     "geosite:ru-markets",
     "geosite:ru-media",
     "geosite:ru-messengers",
     "geosite:ru-movies",
     "geosite:ru-music",
     "geosite:ru-neuro",
     "geosite:ru-porn",
     "geosite:ru-social",
     "geosite:ru-telecom",
     "geosite:ru-videos",
     "geosite:ru-yandex"
    ],
    "outboundTag" : "direct"
   },
   {
    "balancerTag" : "youtube",
    "domain" : [
     "geosite:youtube"
    ]
   },
   {
    "balancerTag" : "gemini",
    "domain" : [
     "geosite:gemini"
    ]
   },
   {
    "balancerTag" : "primary",
    "network" : "tcp,udp"
   }
  ]
 }
}
```
