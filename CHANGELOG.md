# Changelog

## [2.0.1](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/compare/v2.0.0...v2.0.1) (2026-02-27)


### Bug Fixes

* set chartRepo.CachePath to use configured cache directory ([#38](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/38)) ([7adecac](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/7adecacc1f153fe9ca6dfacd704ee753d792404e))

## [2.0.0](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/compare/v1.0.1...v2.0.0) (2026-02-26)


### ⚠ BREAKING CHANGES

* CLI flags renamed
    - --mode → --transport
    - --addr → --listen
    - --timeout → --helm-timeout
    - --max-output → --max-output-size
    - SSE transport removed

### Code Refactoring

* rename tools, add get_notes, remove dead code, improve quality ([#16](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/16)) ([7482240](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/7482240ae5d65cc13cdd762058ac261eab79145d))

## [1.0.1](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/compare/v1.0.0...v1.0.1) (2025-12-17)


### Bug Fixes

* GetChartContents missing Raw/Templates + add Fly.io deployment ([#11](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/11)) ([ee93d23](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ee93d2375fe115af5d3905310c2a9cebeaa1fd95))

## [1.0.0](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/compare/v1.0.0...v1.0.0) (2025-12-16)


### ⚠ BREAKING CHANGES

* **deps:** Update module helm.sh/helm/v3 to v4.0.1 ([#70](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/70))
* **deps:** Update module helm.sh/helm/v3 to v4.0.1 ([#62](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/62))
* **github-action:** Update actions/checkout action to v6 ([#66](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/66))
* **github-action:** Update actions/setup-go action to v6 ([#44](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/44))
* **github-action:** Update jdx/mise-action action to v3 ([#38](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/38))
* **github-action:** Update actions/checkout action to v5 ([#33](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/33))
* **deps:** Update module gopkg.in/yaml.v2 to v3.0.1

### Features

* add an ability to extract full contents of the chart including dependencies ([cf95450](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/cf954508415779b96f05594322acf3d4a561de24))
* add support of streamable HTTP transport ([4757a4e](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/4757a4ed0935b5ca2117e1618e2bbcd0c5ed706f))
* add tool to extract list of dependency charts ([6351de5](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/6351de5aab5251d779e44d0d9ba7f4012cabe29f))
* add tool to get values for chart ([cfccab4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/cfccab4ed5e6ce555c492bc80ba73e1469756f18))
* add tool to retrieve full chart contents ([243fe92](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/243fe92e28b0b78dcb49cd05842cd8149be50d5b))
* **deps:** update golangci-lint to 2.2.1 ([30e07f4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/30e07f43cb72454bd1580da58782cf1e3701af79))
* **deps:** update golangci-lint to 2.2.1 ([74a7c64](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/74a7c64f36090e28d59f67cc9088725f59acd07c))
* **deps:** update golangci-lint to 2.3.0 ([b00615b](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/b00615b5ba95f991bb3af011be6c2bbc23d98ad6))
* **deps:** update golangci-lint to 2.3.0 ([8eb7cd6](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/8eb7cd60a4344d9aecf591b044f30d6211b4ad39))
* **deps:** update golangci-lint to 2.4.0 ([#36](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/36)) ([9c89eaa](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/9c89eaaab71a9413fc449869b81348d35510b442))
* **deps:** update golangci-lint to 2.5.0 ([#50](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/50)) ([ad17d10](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ad17d10cbbba6357e3e3414598c9ac0fcd4e3a51))
* **deps:** update golangci-lint to 2.6.0 ([#58](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/58)) ([7accfff](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/7accfff14d1f4811e7d136c1e014e483570bfeaf))
* **deps:** update golangci-lint to 2.7.0 ([#72](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/72)) ([fc9280b](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/fc9280b53e9cfa3e8dc1fd4c11111062e377f5a7))
* **deps:** update module github.com/mark3labs/mcp-go to v0.31.0 ([482ccca](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/482cccac5fd74e3e3efbe8ab9422d66f9562e6cb))
* **deps:** update module github.com/mark3labs/mcp-go to v0.31.0 ([eaeca24](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/eaeca24636d9e73d520e1e2a136b2ad1ea99e836))
* **deps:** update module github.com/mark3labs/mcp-go to v0.32.0 ([e440c65](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/e440c65d7e9a7a8094bfa7b51ccfc109c186d622))
* **deps:** update module github.com/mark3labs/mcp-go to v0.32.0 ([da512d4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/da512d46acc3b42bcbf2a8a5c714646ef5aa82ec))
* **deps:** update module github.com/mark3labs/mcp-go to v0.33.0 ([fc7397f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/fc7397f6134e85f0824b58f46d0ca8b4b07b9c74))
* **deps:** update module github.com/mark3labs/mcp-go to v0.33.0 ([b38b9b7](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/b38b9b73a225fd53a81ded7d218883281a53f1ea))
* **deps:** update module github.com/mark3labs/mcp-go to v0.34.0 ([d1362e9](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d1362e928cb9012946ea7830b3664f00b648a2cc))
* **deps:** update module github.com/mark3labs/mcp-go to v0.34.0 ([8c770da](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/8c770daa3b090aeb0626e2fe7b73672f110731a5))
* **deps:** update module github.com/mark3labs/mcp-go to v0.35.0 ([aa6f350](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/aa6f350cd34a9f1f35158a5cde4a430dd866b6cc))
* **deps:** update module github.com/mark3labs/mcp-go to v0.35.0 ([8f6da6e](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/8f6da6efa47d2f03f47e5c3766e0d4b737650333))
* **deps:** update module github.com/mark3labs/mcp-go to v0.36.0 ([4d8cbbc](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/4d8cbbcdb6451b5353c58fbf1a83cfb862a5b31a))
* **deps:** update module github.com/mark3labs/mcp-go to v0.36.0 ([17cc752](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/17cc75204935224d6fcecd566c354265e472b774))
* **deps:** update module github.com/mark3labs/mcp-go to v0.37.0 ([#31](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/31)) ([d580d7f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d580d7fc21d1b5e9fa0d246121c6c8cb382a6275))
* **deps:** update module github.com/mark3labs/mcp-go to v0.38.0 ([#40](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/40)) ([050aec4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/050aec450bddf3e0e0d5a56df9fda296adb7724d))
* **deps:** update module github.com/mark3labs/mcp-go to v0.39.1 ([#42](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/42)) ([ff09cac](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ff09cac1bc462e2e297a9bb36778bbad429d0343))
* **deps:** update module github.com/mark3labs/mcp-go to v0.40.0 ([#49](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/49)) ([4658ec0](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/4658ec0cb8f60c15caf4dd6cbddbfb4e1f8d0555))
* **deps:** update module github.com/mark3labs/mcp-go to v0.41.0 ([#51](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/51)) ([26bb73f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/26bb73f6c47fc02a1a1a5efa525b7301f7c4d6ec))
* **deps:** update module github.com/mark3labs/mcp-go to v0.42.0 ([#56](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/56)) ([d4371fb](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d4371fbf2c721e1f4a77d377a45ea2dd53d5e04d))
* **deps:** update module github.com/mark3labs/mcp-go to v0.43.0 ([#59](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/59)) ([eadcd4f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/eadcd4f1384d323c6a33cf68f2d7546f2bc5cd58))
* **deps:** Update module gopkg.in/yaml.v2 to v3.0.1 ([fba4aa1](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/fba4aa140ed870ab97f34e9c2129162d40e42cf1))
* **deps:** update module helm.sh/helm/v3 to v3.19.0 ([#46](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/46)) ([7e8b00b](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/7e8b00b0c717201bf320ee2b067300eb2f7d916c))
* **deps:** Update module helm.sh/helm/v3 to v4.0.1 ([#62](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/62)) ([18e2e1c](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/18e2e1c00f692b123cf7abf60eea863abfe3b321))
* **deps:** Update module helm.sh/helm/v3 to v4.0.1 ([#70](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/70)) ([540d53f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/540d53f25dc6a3c22f5722bc396ab4d413270ee1))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.0 ([3cd753c](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/3cd753ca1992c3dee7e260e0f230d320908f92d9))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.0 ([e8013b9](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/e8013b93df3e6e2bdc8bee1ef98b19601afaa3d2))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.15.0 ([701c6af](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/701c6af424f428004bc6098bb02e751b2c57fab4))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.15.0 ([6877dc6](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/6877dc61c5a5f55fdf9cbbe5892ef01598daf6cd))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.0 ([deb13b5](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/deb13b58ceed580969a7c621042115870e4b2187))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.0 ([ee7a9f0](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ee7a9f0384da1ae8e82260862299f9d4b0d3d27a))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.17.0 ([#54](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/54)) ([b73619b](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/b73619b5e2a300d050841f051b9f9a048a04a801))
* **deps:** update ubi:air-verse/air to 1.62.0 ([479341e](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/479341e7b42cad3d3a4fc6058b95fdd37ea8f34b))
* **deps:** update ubi:air-verse/air to 1.62.0 ([c48af9c](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/c48af9cd7a9d6983c5a2aea964e24738194326ed))
* **deps:** update ubi:air-verse/air to 1.63.0 ([#45](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/45)) ([16c8597](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/16c85974d191f2989cc3c7e13f9a5c202c61fd96))
* **docker-image:** update alpine docker tag to v3.23.0 ([#73](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/73)) ([89b6226](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/89b6226e7b577edf5725b4c7f81257edceb8e6b9))
* migrate to official MCP SDK with cobra CLI and Streamable HTTP ([#3](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/3)) ([e134c69](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/e134c69805e7128f5f476fe714e9b579d69fe6ef))


### Bug Fixes

* **deps:** update act to 0.2.78 ([cf0651f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/cf0651f9a35af787fa5fc5845885094915a708a1))
* **deps:** update act to 0.2.78 ([af3da53](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/af3da537aa73d37dfdc4929004049336a5e5b8fa))
* **deps:** update act to 0.2.79 ([c58f189](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/c58f1890d360b07351f1a836966fe97d94b3f7f8))
* **deps:** update act to 0.2.79 ([ae26e6c](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ae26e6cefb19ca13b82524adf1f5f5925ec582f3))
* **deps:** update act to 0.2.80 ([f1faa38](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/f1faa3881fc2d1f8256ccb4373a6bc5072269aae))
* **deps:** update act to 0.2.80 ([86433ce](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/86433ce6f92dc3d0fcb8e57254f1af719d0fdf44))
* **deps:** update act to 0.2.81 ([#41](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/41)) ([ec7366c](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ec7366c622abc262fba0ac6bc8c7112a101fe51f))
* **deps:** update act to 0.2.82 ([#53](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/53)) ([5b2d2f3](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/5b2d2f336959bfb97be2ce1150a3aaf1f09f7db8))
* **deps:** update act to 0.2.83 ([#69](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/69)) ([5ece481](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/5ece481248aa92f85adea69c3374b0f51eb0ffcc))
* **deps:** update golangci-lint to 2.2.2 ([8678ae5](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/8678ae54e599b916ec96c2d89926c586da0e6f18))
* **deps:** update golangci-lint to 2.2.2 ([7720a48](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/7720a489133e88eb74423cfaf72354c35b9aad6c))
* **deps:** update golangci-lint to 2.3.1 ([1cc5b3a](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/1cc5b3a3386a19363464ad281c67c37c3362bb6a))
* **deps:** update golangci-lint to 2.6.1 ([#60](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/60)) ([6eca19e](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/6eca19e185d1a31acb44c1b24e599c6573167f84))
* **deps:** update golangci-lint to 2.6.2 ([#63](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/63)) ([57e7c94](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/57e7c94c57d2a3f999f541ea9f89482fda5a7b32))
* **deps:** update golangci-lint to 2.7.1 ([#74](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/74)) ([e0a66e5](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/e0a66e5b04b0982992d942d33a9c179dc4d21518))
* **deps:** update golangci-lint to 2.7.2 ([#76](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/76)) ([d145ed2](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d145ed2403bf72b3a3d95d99db18eabda3e3a237))
* **deps:** update module github.com/mark3labs/mcp-go to v0.41.1 ([#52](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/52)) ([566ea91](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/566ea91f06c3ade082fd8a7ba814d790a0e6ad1f))
* **deps:** update module github.com/mark3labs/mcp-go to v0.43.2 ([#67](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/67)) ([ad02e85](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ad02e852d35bb7d543c8d305f8b1695e90b1ee6e))
* **deps:** update module go.uber.org/zap to v1.27.1 ([#65](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/65)) ([319ba59](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/319ba593e8d8cd27e69a5485c0405bb464ed0afe))
* **deps:** update module helm.sh/helm/v3 to v3.18.2 ([480233d](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/480233dbd3411d521542094b4f73c83920bab214))
* **deps:** update module helm.sh/helm/v3 to v3.18.2 ([a4ac9b6](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/a4ac9b6296e338ee47e6d0a54fa24d7ba522e25c))
* **deps:** update module helm.sh/helm/v3 to v3.18.3 ([293c22a](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/293c22a08137c5ace61d091c4fece62e6eb5cc24))
* **deps:** update module helm.sh/helm/v3 to v3.18.3 ([2d13614](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/2d13614568931821cfcea838ef4439e9a5ae126d))
* **deps:** update module helm.sh/helm/v3 to v3.18.4 ([3a316ab](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/3a316abc4b19a7da42f4ab3457b42b2b1f1de16f))
* **deps:** update module helm.sh/helm/v3 to v3.18.4 ([37680e1](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/37680e12e72314dc3e2beb3f912175218d4991c9))
* **deps:** update module helm.sh/helm/v3 to v3.18.5 ([#35](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/35)) ([da7d1ec](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/da7d1ec9984a5b62a1e71efb41b93101e884ff30))
* **deps:** update module helm.sh/helm/v3 to v3.18.6 ([#39](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/39)) ([12f42b8](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/12f42b8b98171c0922240d10c145970156b96e91))
* **deps:** update module helm.sh/helm/v3 to v3.19.2 ([#61](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/61)) ([d242650](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d2426502a9dca563375e5b210867fc18b0c4858e))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.1 ([4ecdd76](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/4ecdd7689323147f5304b114729684f5f42eabad))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.1 ([2ad1b35](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/2ad1b35c0568540602b02fa6998d98ce21a798f0))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.2 ([0704b80](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/0704b80264e9c5c4d9820cb9e187d177839b4cb1))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.2 ([869681f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/869681fbe8f3b5369aec6cba97c10fec4c440585))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.3 ([20b81d6](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/20b81d6f00c29f9ca16bfc62fcaf5b593df01482))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.3 ([478dbfd](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/478dbfd9e4034d260b99567291b48743b867de65))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.1 ([57d5a98](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/57d5a98177fb98b224d320970ed126b1689a5627))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.1 ([821ef5b](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/821ef5b5d991ce8436cd936a914298fa486eb280))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.2 ([cd94b61](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/cd94b614aa67792776a553bb908c9fd4733a54af))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.2 ([29783f2](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/29783f24e39c702f59e7400ef312b0a1c47db128))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.3 ([#34](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/34)) ([80e8982](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/80e89823c7de4efd50bb4a9c18fef0dc1f49f2d8))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.5 ([#37](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/37)) ([35f8da4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/35f8da4b07cda49a7cdc7bc8af1e051c081e937b))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.6 ([#43](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/43)) ([3cdf3f9](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/3cdf3f98b4bc22644d79b0233aaf8dde50d8537a))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.7 ([#47](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/47)) ([8ad2c33](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/8ad2c334eef81b6e72545fba336cd49f5c8c4e87))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.8 ([#48](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/48)) ([7b6751f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/7b6751f26976ae9fd739e541b818017617f9bdf1))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.17.1 ([#55](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/55)) ([56c2532](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/56c2532594ce8718e95c9f83b4239f6d79b98689))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.17.2 ([#57](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/57)) ([ba33f79](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ba33f79637557b481c3a2d7feec32f9152aeb527))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.17.5 ([#75](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/75)) ([f829f68](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/f829f6898c318ce0d91d1b6e87f52d02b97ad5a6))
* **deps:** update ubi:air-verse/air to 1.63.1 ([#64](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/64)) ([2ba08f9](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/2ba08f962fcda4bf285fb5c12b69b2ed7e213bf6))
* **deps:** update ubi:air-verse/air to 1.63.4 ([#68](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/68)) ([b2ebd5e](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/b2ebd5ece54f9504a05ee40f1598a86a491aa5c7))
* fix typo in tools names ([a1e9878](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/a1e98781b0931979919a252baa5af1a2807a7c5c))


### Miscellaneous Chores

* initial release ([acdf9cb](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/acdf9cb39cfb8d605b797b9663c4097eee917682))


### Continuous Integration

* **github-action:** Update actions/checkout action to v5 ([#33](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/33)) ([68f8d13](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/68f8d13c1f847a4f9c534ba3367ff2a1f62addea))
* **github-action:** Update actions/checkout action to v6 ([#66](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/66)) ([c062704](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/c0627040cea7eb853208fde5fea1fef1bcb0c918))
* **github-action:** Update actions/setup-go action to v6 ([#44](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/44)) ([d11fa1f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d11fa1f8fa158fe254f45715ff2bd92941d23b5d))
* **github-action:** Update jdx/mise-action action to v3 ([#38](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/38)) ([5d61eb4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/5d61eb41fbbf703b2bebb00552b389a9b8f9a5ef))

## [1.0.0](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/compare/v1.0.0...v1.0.0) (2025-12-16)


### Miscellaneous Chores

* initial release ([acdf9cb](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/acdf9cb39cfb8d605b797b9663c4097eee917682))

## 1.0.0 (2025-12-16)


### ⚠ BREAKING CHANGES

* **deps:** Update module helm.sh/helm/v3 to v4.0.1 ([#70](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/70))
* **deps:** Update module helm.sh/helm/v3 to v4.0.1 ([#62](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/62))
* **github-action:** Update actions/checkout action to v6 ([#66](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/66))
* **github-action:** Update actions/setup-go action to v6 ([#44](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/44))
* **github-action:** Update jdx/mise-action action to v3 ([#38](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/38))
* **github-action:** Update actions/checkout action to v5 ([#33](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/33))
* **deps:** Update module gopkg.in/yaml.v2 to v3.0.1

### Features

* add an ability to extract full contents of the chart including dependencies ([cf95450](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/cf954508415779b96f05594322acf3d4a561de24))
* add support of streamable HTTP transport ([4757a4e](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/4757a4ed0935b5ca2117e1618e2bbcd0c5ed706f))
* add tool to extract list of dependency charts ([6351de5](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/6351de5aab5251d779e44d0d9ba7f4012cabe29f))
* add tool to get values for chart ([cfccab4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/cfccab4ed5e6ce555c492bc80ba73e1469756f18))
* add tool to retrieve full chart contents ([243fe92](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/243fe92e28b0b78dcb49cd05842cd8149be50d5b))
* **deps:** update golangci-lint to 2.2.1 ([30e07f4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/30e07f43cb72454bd1580da58782cf1e3701af79))
* **deps:** update golangci-lint to 2.2.1 ([74a7c64](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/74a7c64f36090e28d59f67cc9088725f59acd07c))
* **deps:** update golangci-lint to 2.3.0 ([b00615b](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/b00615b5ba95f991bb3af011be6c2bbc23d98ad6))
* **deps:** update golangci-lint to 2.3.0 ([8eb7cd6](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/8eb7cd60a4344d9aecf591b044f30d6211b4ad39))
* **deps:** update golangci-lint to 2.4.0 ([#36](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/36)) ([9c89eaa](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/9c89eaaab71a9413fc449869b81348d35510b442))
* **deps:** update golangci-lint to 2.5.0 ([#50](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/50)) ([ad17d10](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ad17d10cbbba6357e3e3414598c9ac0fcd4e3a51))
* **deps:** update golangci-lint to 2.6.0 ([#58](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/58)) ([7accfff](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/7accfff14d1f4811e7d136c1e014e483570bfeaf))
* **deps:** update golangci-lint to 2.7.0 ([#72](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/72)) ([fc9280b](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/fc9280b53e9cfa3e8dc1fd4c11111062e377f5a7))
* **deps:** update module github.com/mark3labs/mcp-go to v0.31.0 ([482ccca](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/482cccac5fd74e3e3efbe8ab9422d66f9562e6cb))
* **deps:** update module github.com/mark3labs/mcp-go to v0.31.0 ([eaeca24](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/eaeca24636d9e73d520e1e2a136b2ad1ea99e836))
* **deps:** update module github.com/mark3labs/mcp-go to v0.32.0 ([e440c65](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/e440c65d7e9a7a8094bfa7b51ccfc109c186d622))
* **deps:** update module github.com/mark3labs/mcp-go to v0.32.0 ([da512d4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/da512d46acc3b42bcbf2a8a5c714646ef5aa82ec))
* **deps:** update module github.com/mark3labs/mcp-go to v0.33.0 ([fc7397f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/fc7397f6134e85f0824b58f46d0ca8b4b07b9c74))
* **deps:** update module github.com/mark3labs/mcp-go to v0.33.0 ([b38b9b7](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/b38b9b73a225fd53a81ded7d218883281a53f1ea))
* **deps:** update module github.com/mark3labs/mcp-go to v0.34.0 ([d1362e9](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d1362e928cb9012946ea7830b3664f00b648a2cc))
* **deps:** update module github.com/mark3labs/mcp-go to v0.34.0 ([8c770da](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/8c770daa3b090aeb0626e2fe7b73672f110731a5))
* **deps:** update module github.com/mark3labs/mcp-go to v0.35.0 ([aa6f350](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/aa6f350cd34a9f1f35158a5cde4a430dd866b6cc))
* **deps:** update module github.com/mark3labs/mcp-go to v0.35.0 ([8f6da6e](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/8f6da6efa47d2f03f47e5c3766e0d4b737650333))
* **deps:** update module github.com/mark3labs/mcp-go to v0.36.0 ([4d8cbbc](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/4d8cbbcdb6451b5353c58fbf1a83cfb862a5b31a))
* **deps:** update module github.com/mark3labs/mcp-go to v0.36.0 ([17cc752](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/17cc75204935224d6fcecd566c354265e472b774))
* **deps:** update module github.com/mark3labs/mcp-go to v0.37.0 ([#31](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/31)) ([d580d7f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d580d7fc21d1b5e9fa0d246121c6c8cb382a6275))
* **deps:** update module github.com/mark3labs/mcp-go to v0.38.0 ([#40](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/40)) ([050aec4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/050aec450bddf3e0e0d5a56df9fda296adb7724d))
* **deps:** update module github.com/mark3labs/mcp-go to v0.39.1 ([#42](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/42)) ([ff09cac](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ff09cac1bc462e2e297a9bb36778bbad429d0343))
* **deps:** update module github.com/mark3labs/mcp-go to v0.40.0 ([#49](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/49)) ([4658ec0](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/4658ec0cb8f60c15caf4dd6cbddbfb4e1f8d0555))
* **deps:** update module github.com/mark3labs/mcp-go to v0.41.0 ([#51](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/51)) ([26bb73f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/26bb73f6c47fc02a1a1a5efa525b7301f7c4d6ec))
* **deps:** update module github.com/mark3labs/mcp-go to v0.42.0 ([#56](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/56)) ([d4371fb](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d4371fbf2c721e1f4a77d377a45ea2dd53d5e04d))
* **deps:** update module github.com/mark3labs/mcp-go to v0.43.0 ([#59](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/59)) ([eadcd4f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/eadcd4f1384d323c6a33cf68f2d7546f2bc5cd58))
* **deps:** Update module gopkg.in/yaml.v2 to v3.0.1 ([fba4aa1](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/fba4aa140ed870ab97f34e9c2129162d40e42cf1))
* **deps:** update module helm.sh/helm/v3 to v3.19.0 ([#46](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/46)) ([7e8b00b](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/7e8b00b0c717201bf320ee2b067300eb2f7d916c))
* **deps:** Update module helm.sh/helm/v3 to v4.0.1 ([#62](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/62)) ([18e2e1c](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/18e2e1c00f692b123cf7abf60eea863abfe3b321))
* **deps:** Update module helm.sh/helm/v3 to v4.0.1 ([#70](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/70)) ([540d53f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/540d53f25dc6a3c22f5722bc396ab4d413270ee1))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.0 ([3cd753c](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/3cd753ca1992c3dee7e260e0f230d320908f92d9))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.0 ([e8013b9](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/e8013b93df3e6e2bdc8bee1ef98b19601afaa3d2))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.15.0 ([701c6af](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/701c6af424f428004bc6098bb02e751b2c57fab4))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.15.0 ([6877dc6](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/6877dc61c5a5f55fdf9cbbe5892ef01598daf6cd))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.0 ([deb13b5](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/deb13b58ceed580969a7c621042115870e4b2187))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.0 ([ee7a9f0](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ee7a9f0384da1ae8e82260862299f9d4b0d3d27a))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.17.0 ([#54](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/54)) ([b73619b](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/b73619b5e2a300d050841f051b9f9a048a04a801))
* **deps:** update ubi:air-verse/air to 1.62.0 ([479341e](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/479341e7b42cad3d3a4fc6058b95fdd37ea8f34b))
* **deps:** update ubi:air-verse/air to 1.62.0 ([c48af9c](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/c48af9cd7a9d6983c5a2aea964e24738194326ed))
* **deps:** update ubi:air-verse/air to 1.63.0 ([#45](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/45)) ([16c8597](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/16c85974d191f2989cc3c7e13f9a5c202c61fd96))
* **docker-image:** update alpine docker tag to v3.23.0 ([#73](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/73)) ([89b6226](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/89b6226e7b577edf5725b4c7f81257edceb8e6b9))
* migrate to official MCP SDK with cobra CLI and Streamable HTTP ([#3](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/3)) ([e134c69](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/e134c69805e7128f5f476fe714e9b579d69fe6ef))


### Bug Fixes

* **deps:** update act to 0.2.78 ([cf0651f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/cf0651f9a35af787fa5fc5845885094915a708a1))
* **deps:** update act to 0.2.78 ([af3da53](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/af3da537aa73d37dfdc4929004049336a5e5b8fa))
* **deps:** update act to 0.2.79 ([c58f189](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/c58f1890d360b07351f1a836966fe97d94b3f7f8))
* **deps:** update act to 0.2.79 ([ae26e6c](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ae26e6cefb19ca13b82524adf1f5f5925ec582f3))
* **deps:** update act to 0.2.80 ([f1faa38](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/f1faa3881fc2d1f8256ccb4373a6bc5072269aae))
* **deps:** update act to 0.2.80 ([86433ce](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/86433ce6f92dc3d0fcb8e57254f1af719d0fdf44))
* **deps:** update act to 0.2.81 ([#41](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/41)) ([ec7366c](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ec7366c622abc262fba0ac6bc8c7112a101fe51f))
* **deps:** update act to 0.2.82 ([#53](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/53)) ([5b2d2f3](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/5b2d2f336959bfb97be2ce1150a3aaf1f09f7db8))
* **deps:** update act to 0.2.83 ([#69](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/69)) ([5ece481](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/5ece481248aa92f85adea69c3374b0f51eb0ffcc))
* **deps:** update golangci-lint to 2.2.2 ([8678ae5](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/8678ae54e599b916ec96c2d89926c586da0e6f18))
* **deps:** update golangci-lint to 2.2.2 ([7720a48](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/7720a489133e88eb74423cfaf72354c35b9aad6c))
* **deps:** update golangci-lint to 2.3.1 ([1cc5b3a](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/1cc5b3a3386a19363464ad281c67c37c3362bb6a))
* **deps:** update golangci-lint to 2.6.1 ([#60](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/60)) ([6eca19e](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/6eca19e185d1a31acb44c1b24e599c6573167f84))
* **deps:** update golangci-lint to 2.6.2 ([#63](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/63)) ([57e7c94](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/57e7c94c57d2a3f999f541ea9f89482fda5a7b32))
* **deps:** update golangci-lint to 2.7.1 ([#74](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/74)) ([e0a66e5](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/e0a66e5b04b0982992d942d33a9c179dc4d21518))
* **deps:** update golangci-lint to 2.7.2 ([#76](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/76)) ([d145ed2](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d145ed2403bf72b3a3d95d99db18eabda3e3a237))
* **deps:** update module github.com/mark3labs/mcp-go to v0.41.1 ([#52](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/52)) ([566ea91](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/566ea91f06c3ade082fd8a7ba814d790a0e6ad1f))
* **deps:** update module github.com/mark3labs/mcp-go to v0.43.2 ([#67](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/67)) ([ad02e85](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ad02e852d35bb7d543c8d305f8b1695e90b1ee6e))
* **deps:** update module go.uber.org/zap to v1.27.1 ([#65](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/65)) ([319ba59](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/319ba593e8d8cd27e69a5485c0405bb464ed0afe))
* **deps:** update module helm.sh/helm/v3 to v3.18.2 ([480233d](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/480233dbd3411d521542094b4f73c83920bab214))
* **deps:** update module helm.sh/helm/v3 to v3.18.2 ([a4ac9b6](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/a4ac9b6296e338ee47e6d0a54fa24d7ba522e25c))
* **deps:** update module helm.sh/helm/v3 to v3.18.3 ([293c22a](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/293c22a08137c5ace61d091c4fece62e6eb5cc24))
* **deps:** update module helm.sh/helm/v3 to v3.18.3 ([2d13614](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/2d13614568931821cfcea838ef4439e9a5ae126d))
* **deps:** update module helm.sh/helm/v3 to v3.18.4 ([3a316ab](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/3a316abc4b19a7da42f4ab3457b42b2b1f1de16f))
* **deps:** update module helm.sh/helm/v3 to v3.18.4 ([37680e1](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/37680e12e72314dc3e2beb3f912175218d4991c9))
* **deps:** update module helm.sh/helm/v3 to v3.18.5 ([#35](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/35)) ([da7d1ec](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/da7d1ec9984a5b62a1e71efb41b93101e884ff30))
* **deps:** update module helm.sh/helm/v3 to v3.18.6 ([#39](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/39)) ([12f42b8](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/12f42b8b98171c0922240d10c145970156b96e91))
* **deps:** update module helm.sh/helm/v3 to v3.19.2 ([#61](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/61)) ([d242650](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d2426502a9dca563375e5b210867fc18b0c4858e))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.1 ([4ecdd76](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/4ecdd7689323147f5304b114729684f5f42eabad))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.1 ([2ad1b35](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/2ad1b35c0568540602b02fa6998d98ce21a798f0))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.2 ([0704b80](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/0704b80264e9c5c4d9820cb9e187d177839b4cb1))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.2 ([869681f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/869681fbe8f3b5369aec6cba97c10fec4c440585))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.3 ([20b81d6](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/20b81d6f00c29f9ca16bfc62fcaf5b593df01482))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.14.3 ([478dbfd](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/478dbfd9e4034d260b99567291b48743b867de65))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.1 ([57d5a98](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/57d5a98177fb98b224d320970ed126b1689a5627))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.1 ([821ef5b](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/821ef5b5d991ce8436cd936a914298fa486eb280))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.2 ([cd94b61](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/cd94b614aa67792776a553bb908c9fd4733a54af))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.2 ([29783f2](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/29783f24e39c702f59e7400ef312b0a1c47db128))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.3 ([#34](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/34)) ([80e8982](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/80e89823c7de4efd50bb4a9c18fef0dc1f49f2d8))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.5 ([#37](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/37)) ([35f8da4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/35f8da4b07cda49a7cdc7bc8af1e051c081e937b))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.6 ([#43](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/43)) ([3cdf3f9](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/3cdf3f98b4bc22644d79b0233aaf8dde50d8537a))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.7 ([#47](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/47)) ([8ad2c33](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/8ad2c334eef81b6e72545fba336cd49f5c8c4e87))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.16.8 ([#48](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/48)) ([7b6751f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/7b6751f26976ae9fd739e541b818017617f9bdf1))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.17.1 ([#55](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/55)) ([56c2532](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/56c2532594ce8718e95c9f83b4239f6d79b98689))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.17.2 ([#57](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/57)) ([ba33f79](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/ba33f79637557b481c3a2d7feec32f9152aeb527))
* **deps:** update npm:@modelcontextprotocol/inspector to 0.17.5 ([#75](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/75)) ([f829f68](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/f829f6898c318ce0d91d1b6e87f52d02b97ad5a6))
* **deps:** update ubi:air-verse/air to 1.63.1 ([#64](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/64)) ([2ba08f9](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/2ba08f962fcda4bf285fb5c12b69b2ed7e213bf6))
* **deps:** update ubi:air-verse/air to 1.63.4 ([#68](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/68)) ([b2ebd5e](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/b2ebd5ece54f9504a05ee40f1598a86a491aa5c7))
* fix typo in tools names ([a1e9878](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/a1e98781b0931979919a252baa5af1a2807a7c5c))


### Continuous Integration

* **github-action:** Update actions/checkout action to v5 ([#33](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/33)) ([68f8d13](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/68f8d13c1f847a4f9c534ba3367ff2a1f62addea))
* **github-action:** Update actions/checkout action to v6 ([#66](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/66)) ([c062704](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/c0627040cea7eb853208fde5fea1fef1bcb0c918))
* **github-action:** Update actions/setup-go action to v6 ([#44](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/44)) ([d11fa1f](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/d11fa1f8fa158fe254f45715ff2bd92941d23b5d))
* **github-action:** Update jdx/mise-action action to v3 ([#38](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/38)) ([5d61eb4](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/5d61eb41fbbf703b2bebb00552b389a9b8f9a5ef))
