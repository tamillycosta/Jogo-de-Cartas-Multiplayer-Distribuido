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


---

### 🖥️ Rodando a **Blockchain** 

 Em um terminal rode:
   ```bash
      ganache --port 7545 --deterministic
   ```
     ```bash
      Você vai ver algo assim:
      Ganache CLI v7.9.1
      
      Available Accounts
      ==================
      (0) 0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1 (100 ETH)
   ```

###  Migrando os contratos inteligentes 

   ```bash
      cd blockchain
      truffle migrate --reset
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
    
###  Rodando banco de dados local 
 ```bash
      sudo docker-compose up --build mysql-a
```
###  Rodando servidor 
 ```bash
   cd cmd

   export DB_PORT=3307
   export DB_USER=root
   export DB_PASSWORD=senha_a
   export DB_NAME=game_server_a
   export RAFT_BOOTSTRAP=true // se subir mais de um servidor mude para false 
   export PK=alguma chave privada gerada automatiacamente pela blockchain
   
   go run main.go
```

### 🖥️ Acessando como **Cliente**  

1. Navegue até o diretório do cliente:
   ```bash  
      cd cmd/client
   ```
   
2. Execute o cliente:
   ```bash
      go run .
   ``` 
---




