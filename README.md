Korp_teste_DiegoFaria
Sistema de Notas Fiscais com Microserviços em Go + .NET 8
Docker | PostgreSQL | Redis | Swagger | CI/CD |

Docker
Go
.NET
PostgreSQL
Redis
GitHub Actions

Funcionalidades
Serviço,Endpoint,Descrição
Estoque,GET /health,Health check com uptime
,GET /api/produtos,Lista com paginação
,POST /api/produtos/reservar,Reserva com lock distribuído Redis
Faturamento,GET /api/notas-fiscais,Lista com filtro por data/status
,POST /api/notas-fiscais,Emissão com idempotência
,POST /api/notas-fiscais/{id}/imprimir,Baixa estoque + PDF

Arquitetura

graph TD
    A[Cliente] -->|HTTP| B(Nginx)
    B -->|8080| C[servico-estoque (Go)]
    B -->|5000| D[servico-faturamento (.NET 8)]
    C --> E[PostgreSQL (estoque)]
    D --> F[PostgreSQL (faturamento)]
    C --> G[Redis (lock + cache)]
    D --> G

Tecnologias

Go 1.23 → servico-estoque
.NET 8 → servico-faturamento
PostgreSQL 16
Redis 7
Docker + Docker Compose
Nginx (proxy reverso)
GitHub Actions (CI/CD)

Como Rodar

# Clone o projeto
git clone https://github.com/DiegoFariaFuel/Korp_teste_DiegoFaria.git
cd Korp_teste_DiegoFaria

# Alias útil
echo "alias dc='docker compose'" >> ~/.bashrc && source ~/.bashrc

# Suba tudo
dc up -d --build

# Aguarde
sleep 20

Endpoints
Estoque (localhost:8080)
GET /health
GET /api/produtos?page=1&size=10
POST /api/produtos
POST /api/produtos/reservar

Faturamento (localhost:5000)
GET /api/notas-fiscais
POST /api/notas-fiscais
POST /api/notas-fiscais/{id}/imprimir
GET /api/notas-fiscais/{id}/pdf


Exemplo de Emissão de Nota Fiscal

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

Comandos Úteis

# Iniciar
dc up -d

# Parar e limpar
dc down -v

# Ver logs
dc logs -f estoque
dc logs -f faturamento

# Rebuild
dc build --no-cache

# Banco de Dados
psql -h localhost -p 5432 -U postgres -d faturamento
psql -h localhost -p 5433 -U postgres -d estoque

CI/CD (GitHub Actions)
name: CI/CD
on: [push, pull_request]
jobs:
  build-and-test:
    runs-on: ubuntu-latest
    services:
      postgres: { image: postgres:16, env: { POSTGRES_PASSWORD: postgres }, ports: [5432:5432] }
      redis: { image: redis:7, ports: [6379:6379] }
    steps:
      - uses: actions/checkout@v4
      - name: Build Docker
        run: docker compose build
      - name: Test
        run: docker compose up --abort-on-container-exit


Banco de Dados
Host: localhost
Porta Faturamento: 5432
Porta Estoque: 5433
Banco: faturamento / estoque
Usuário: postgres
Senha: postgres

Próximos Passos

 Autenticação JWT + RBAC
 Front-end em React + Tailwind ou Angular
 Monitoramento com Prometheus + Grafana
 Deploy na AWS (ECS + Fargate)
 Testes E2E com Playwright

 MIT License