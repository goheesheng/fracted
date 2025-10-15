# 生产环境
https://demo.fracted.xyz/generator.html

# 本地开发
http://localhost:8080/generator.html

生产环境quick-link生成：
https://demo.fracted.xyz/?merchant=0xB7aa464b19037CF3dB7F723504dFafE7b63aAb84&dstEid=40231&dstToken=0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d&amount=1000000

本地quick-link生成：
http://localhost:8080/?merchant=0xB7aa464b19037CF3dB7F723504dFafE7b63aAb84&dstEid=40231&dstToken=0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d&amount=1000000



dashboard界面
http://localhost:8081
http://localhost:8081/test-merchant.html (测试页面)
http://localhost:8081/login.html 



Deployed contract: MyOApp, network: arbitrum-sepolia, address: 0x9E3737436CD46a8ccF81a6720977081D5667f57C
Deployed contract: MyOApp, network: base-sepolia, address: 0x5C5254f25C24eC1dFb612067AB6CbD15E6430071



Deployed contract: MyOApp, network: arbitrum-sepolia, address: 0x7788F9FB737C48550Ea842fC9Cb192FbEA99a890
Deployed contract: MyOApp, network: base-sepolia, address: 0x0cfE9BdF5C027623C44991fE5Ca493A93B62bD27

pnpm run compile:hardhat

#命令
1.
npx hardhat deploy --network base-sepolia
npx hardhat deploy --network arbitrum-sepolia
pnpm hardhat lz:oapp:wire --oapp-config layerzero.config.ts

2
npx hardhat lz:oapp:setRoute --help


3.流动性
# 首先需要批准合约可以转移你的代币
npx hardhat lz:oapp:approveToken --network arbitrum-sepolia --token 0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d --amount 1000000000000000000

npx hardhat lz:oapp:yyapproveToken --network arbitrum-sepolia --token 0xdAC17F958D2ee523a2206206994597C13D831ec7 --amount 1000000000000000000


npx hardhat lz:oapp:depositToken --network arbitrum-sepolia --token 0xdAC17F958D2ee523a2206206994597C13D831ec7 --amount 10000000000

npx hardhat lz:oapp:depositToken --network arbitrum-sepolia --token 0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d --amount 10000

usdt：0xdAC17F958D2ee523a2206206994597C13D831ec7



arbi-usdt：0x30fA2FbE15c1EaDfbEF28C188b7B8dbd3c1Ff2eB
base-usdt：0x323e78f944A9a1FcF3a10efcC5319DBb0bB6e673
arbi-usdc：0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d
base-usdc：0x036CbD53842c5426634e7929541eC2318f3dCF7e


npx hardhat lz:oapp:requestPayoutToken --network base-sepolia --dst-eid 40231 --src-token 0x323e78f944A9a1FcF3a10efcC5319DBb0bB6e673 --merchant 0xB7aa464b19037CF3dB7F723504dFafE7b63aAb84 --amount 1000000


npx hardhat lz:oapp:approveToken --network base-sepolia --token 0x323e78f944A9a1FcF3a10efcC5319DBb0bB6e673 --amount 1000000000
npx hardhat lz:oapp:requestPayoutToken --network base-sepolia --dst-eid 40231 --src-token 0x323e78f944A9a1FcF3a10efcC5319DBb0bB6e673 --merchant 0xB7aa464b19037CF3dB7F723504dFafE7b63aAb84 --amount 10000000
