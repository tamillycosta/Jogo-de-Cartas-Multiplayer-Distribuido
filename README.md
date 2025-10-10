# 🌟 MagiCards 🌟


## 🚀 Como rodar o projeto

### ✅ Pré-requisitos
- [Go](https://go.dev/dl/)
- [Docker](https://www.docker.com/) 

---

### 🖥️ Rodando o **Aplicação** 
1. Clone o repositório:
   ```bash
     https://github.com/tamillycosta/Jogo-de-Cartas-Multiplayer-Distribuido.git
     cd Jogo-de-Cartas-Multiplayer-Distribuido

2. Adicione config do ambiente virtual
   Ex .env:
      ```bash
        SERVER_ID=server-a
        SERVER_ADDRESS=server-a
        GOSSIP_PORT=7947
        PORT=8080
        SEED_SERVERS=
        
        DB_HOST=mysql-a
        DB_PORT=3306
        DB_USER=root
        DB_PASSWORD=senha_a
        DB_NAME=game_server_a
        DB_SSLMODE=disable

3. Suba o container do servidor:
   ```bash
     docker-compose up --build
   
