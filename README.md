# 🌟 MagiCards 🌟

#   Links Importantes 
-   Ganache (https://archive.trufflesuite.com/ganache/)
-   Truffle (https://www.softobotics.org/blogs/truffle-a-comprehensive-guide-to-smart-contract-development-on-the-blockchain/)
-   Abigen (https://geth.ethereum.org/docs/tools/abigen)


## 🚀 Como rodar o projeto

### ✅ Depêndencias 
- # Instalar Truffle e Ganache 
   npm install -g truffle ganache
- # Instalar abigen (ferramenta Go)
   go install github.com/ethereum/go-ethereum/cmd/abigen@latest
- # Instalar dependências do projeto (npm)
   npm install 


---

### 🌟 Rodando a **Blockchain** 

 Em um terminal rode:
   ```bash
      ganache --port 7545 --deterministic
   ```
     
      Você vai ver algo assim:
      Ganache CLI v7.9.1
      
      Available Accounts
      ==================
      (0) 0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1 (100 ETH)
---   

### 🌟 Migrando os contratos inteligentes 

   ```bash
      cd blockchain
      truffle compile
      truffle migrate 
   ```
   ```bash
      Você vai ver algo assim:
      2_deploy_contracts.js
   =====================

   Deploying 'SimpleStorage'
   -------------------------
   > transaction hash:    0x...
   > contract address:    0x5FbDB2315678afecb367f032d93F642f64180aa3
   > block number:        1
   > account:             0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1
   ```
---


### 🌟 Rodando banco de dados local 
 ```
      sudo docker-compose up --build mysql-a
```

---

### 🌟 Rodando servidor 
 Monte um arquivo executavel com a seguinte configuração e execute:
 ```
#!/bin/bash

export PORT=8080
export GOSSIP_PORT=7947
export RAFT_BOOTSTRAP=true 

export DB_HOST=localhost
export DB_PORT=3307
export DB_USER=root
export DB_PASSWORD=senha_a
export DB_NAME=game_server_a
export PK=chave privada
export DISCOVERY_PORT=9000

export MATCH_CONTRACT=endereçoDocontrato
export PACKAGE_CONTRACT=endereçoDocontrato
export CARD_CONTRACT=endereçoDocontrato

cd cmd
go run main.go

```
---

### 🌟 Acessando  **Menu Interativo da Bloockchain**  

1. Navegue até o diretório do menu:
   ```bash  
      cd cmd/blockchain-menu
   ```
2. Monte um arquivo executavel com a seguinte configuração e execute:
   ```bash
      export MATCH_CONTRACT=endereçoDoContrato
      export PACKAGE_CONTRACT=endereçoDoContrato
      export CARD_CONTRACT=endereçoDoContrato
   
      go run . 
---

### 🌟 Acessando como **Cliente**  

1. Navegue até o diretório do cliente:
   ```bash  
      cd cmd/client
   ```
   
2. Execute o cliente:
   ```bash
      go run .
   ``` 
---

### 🌟 Dicas de uso com truffle
para pegar o endereço de um contrato
   ```bash  
      truffle console
      > NOME DO CONTRATO.address
   ```



