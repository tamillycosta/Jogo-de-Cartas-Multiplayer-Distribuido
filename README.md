# 🌟 MagiCards 🌟


## 🚀 Como rodar o projeto

### ✅ Pré-requisitos
- [Go](https://go.dev/dl/)
- [Docker](https://www.docker.com/) 

---

### 🖥️ Rodando a **Aplicação** 

1. Clone o repositório:
   ```bash
     https://github.com/tamillycosta/Jogo-de-Cartas-Multiplayer-Distribuido.git
     cd Jogo-de-Cartas-Multiplayer-Distribuido
   ```
   
2. Configure as variáveis de ambiente

   Crie um arquivo `.env` na raiz do projeto com as informações do banco de dados:
      ```env  
        DB_PASSWORD=senha_a
        DB_NAME=game_server_a
      ```
      
4. Suba os containers do servidor
   
   Para subir todos os serviços:
     ```bash  
        docker compose up --build
     ```
   Para subir apenas um servidor:
      ```bash
         docker compose up --build server-a
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


export DB_PORT=3307
export DB_USER=root
export DB_PASSWORD=senha_a
export DB_NAME=game_server_a
export RAFT_BOOTSTRAP=true
export PK=4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d

go run main.go
