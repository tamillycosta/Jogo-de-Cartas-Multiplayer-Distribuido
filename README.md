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
        DB_PASSWORD=senha_a
        DB_NAME=game_server_a
       
3. Suba o container do servidor:



4. Para subir cada servidores separadamente use:   
   ```bash
     docker-compose up --build server-a

  
5.Para subir todos os servidores :   
   ```bash
     docker-compose up --build 
   
