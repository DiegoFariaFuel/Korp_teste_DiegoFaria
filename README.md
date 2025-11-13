# Korp_teste_DiegoFaria

**Sistema de Notas Fiscais com Microserviços em Go + .NET 8**  
`Docker` `Go` `.NET 8` `PostgreSQL` `Redis` `Swagger` `CI/CD`

[![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat&logo=docker&logoColor=white)](https://www.docker.com/)
[![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![.NET](https://img.shields.io/badge/.NET-512BD4?style=flat&logo=.net&logoColor=white)](https://dotnet.microsoft.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-336791?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-DC382D?style=flat&logo=redis&logoColor=white)](https://redis.io/)
[![GitHub Actions](https://img.shields.io/badge/GitHub_Actions-2088FF?style=flat&logo=github-actions&logoColor=white)](https://github.com/features/actions)

---

## 📋 Funcionalidades

| Serviço       | Endpoint                            | Descrição                                  |
|---------------|-------------------------------------|--------------------------------------------|
| **Estoque**   | `GET /health`                       | Health check com uptime                    |
|               | `GET /api/produtos`                 | Lista com paginação                        |
|               | `POST /api/produtos/reservar`       | Reserva com *lock distribuído* via Redis    |
| **Faturamento** | `GET /api/notas-fiscais`          | Lista com filtro por data/status           |
|               | `POST /api/notas-fiscais`           | Emissão com *idempotência*                 |
|               | `POST /api/notas-fiscais/{id}/imprimir` | Baixa estoque + gera PDF               |

---

## 🏗️ Arquitetura

```mermaid
graph TD
    A[Cliente] -->|HTTP| B[Nginx]
    B -->|8080| C[servico-estoque<br/>Go]
    B -->|5000| D[servico-faturamento<br/>.NET 8]
    C --> E[PostgreSQL<br/>estoque]
    D --> F[PostgreSQL<br/>faturamento]
    C --> G[Redis<br/>lock + cache]
    D --> G

    classDef go fill:#00ADD8,color:#fff
    classDef dotnet fill:#512BD4,color:#fff
    classDef db fill:#336791,color:#fff
    classDef redis fill:#DC382D,color:#fff
    classDef nginx fill:#2496ED,color:#fff

    class C go
    class D dotnet
    class E,F db
    class G redis
    class B nginx

## 🛠️ Tecnologias

| Tecnologia       | Versão       | Uso                            |
|------------------|--------------|--------------------------------|
| Go               | `1.23`       | `servico-estoque`              |
| .NET             | `8`          | `servico-faturamento`          |
| PostgreSQL       | `16`         | Bancos `estoque` e `faturamento` |
| Redis            | `7`          | Lock distribuído e cache       |
| Docker + Compose | -            | Contêineres e orquestração     |
| Nginx            | -            | Proxy reverso                  |
| GitHub Actions   | -            | CI/CD                          |

---

## 🚀 Como Rodar

### Pré-requisitos
- Docker Desktop
- Docker Compose
- Git
- `jq` (opcional, para formatar JSON)

### Instalação

```bash
# 1. Clone o projeto
git clone https://github.com/DiegoFariaFuel/Korp_teste_DiegoFaria.git
cd Korp_teste_DiegoFaria

# 2. Alias útil (opcional)
echo "alias dc='docker compose'" >> ~/.bashrc && source ~/.bashrc

# 3. Suba os serviços
dc up -d --build

# 4. Aguarde inicialização (~20s)
sleep 20
```

### Endpoints Disponíveis

#### **Estoque** → `http://localhost:8080`
```
GET  /health
GET  /api/produtos?page=1&size=10
POST /api/produtos
POST /api/produtos/reservar
```

#### **Faturamento** → `http://localhost:5000`
```
GET  /api/notas-fiscais
POST /api/notas-fiscais
POST /api/notas-fiscais/{id}/imprimir
GET  /api/notas-fiscais/{id}/pdf
```

---

## 💡 Exemplo: Emissão de Nota Fiscal

```bash
curl -X POST http://localhost:5000/api/notas-fiscais \
  -H "Content-Type: application/json" \
  -d '{
    "clienteId": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "itens": [
      {
        "produtoId": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
        "descricaoProduto": "Campeão do Docker",
        "quantidade": 1,
        "valorUnitario": 999999.00
      }
    ]
  }' | jq .
```

---

## ⚙️ Comandos Úteis

```bash
# Iniciar serviços
dc up -d

# Parar e remover volumes
dc down -v

# Ver logs em tempo real
dc logs -f estoque
dc logs -f faturamento
dc logs -f nginx

# Reconstruir imagens
dc build --no-cache

# Acessar bancos de dados
psql -h localhost -p 5432 -U postgres -d faturamento
psql -h localhost -p 5433 -U postgres -d estoque

# Ver containers rodando
dc ps

# Ver redes
dc network ls
```

---

## 🗄️ Banco de Dados

| Parâmetro       | Valor                     |
|-----------------|---------------------------|
| **Host**        | `localhost`               |
| **Porta (Faturamento)** | `5432`              |
| **Porta (Estoque)**     | `5433`              |
| **Bancos**      | `faturamento`, `estoque`  |
| **Usuário**     | `postgres`                |
| **Senha**       | `postgres`                |

---

## 🔄 CI/CD com GitHub Actions

```yaml
name: CI/CD
on: [push, pull_request]
jobs:
  build-and-test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: postgres
        ports: [5432:5432]
      redis:
        image: redis:7
        ports: [6379:6379]
    steps:
      - uses: actions/checkout@v4
      - name: Build Docker
        run: docker compose build
      - name: Test
        run: docker compose up --abort-on-container-exit
```

---

## 📈 Próximos Passos

- [ ] Autenticação JWT + RBAC  
- [ ] Front-end em **React + Tailwind** ou **Angular**  
- [ ] Monitoramento com **Prometheus + Grafana**  
- [ ] Deploy na **AWS (ECS + Fargate)**  
- [ ] Testes E2E com **Playwright**

---

## 📄 Licença

**MIT License** – Veja o arquivo [LICENSE](LICENSE) para detalhes.

---

<div align="center">
  
**Feito com 💙 por Diego Faria**  
<a href="https://github.com/DiegoFariaFuel"><img src="https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white" alt="GitHub"></a>
<a href="https://www.linkedin.com/in/diegofaria"><img src="https://img.shields.io/badge/LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white" alt="LinkedIn"></a>

</div>