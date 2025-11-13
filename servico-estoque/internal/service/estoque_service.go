// internal/service/estoque_service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"servico-estoque/internal/domain"
	"servico-estoque/internal/repository"
	"servico-estoque/pkg/lock"
)

type EstoqueService struct {
	repo   repository.ProdutoRepository
	cache  *redis.Client
	lock   *lock.DistributedLock
	logger *zap.Logger
}

func NewEstoqueService(
	repo repository.ProdutoRepository,
	cache *redis.Client,
	lock *lock.DistributedLock,
	logger *zap.Logger,
) *EstoqueService {
	return &EstoqueService{
		repo:   repo,
		cache:  cache,
		lock:   lock,
		logger: logger,
	}
}

// CriarProduto cria um novo produto
func (s *EstoqueService) CriarProduto(ctx context.Context, req domain.CriarProdutoRequest) (*domain.Produto, error) {
	// Validar se código já existe
	existente, err := s.repo.FindByCodigo(ctx, req.Codigo)
	if err != nil && err != domain.ErrProdutoNaoEncontrado {
		return nil, err
	}
	if existente != nil {
		return nil, domain.ErrCodigoDuplicado
	}

	produto := &domain.Produto{
		Codigo:    req.Codigo,
		Descricao: req.Descricao,
		Saldo:     req.Saldo,
		Reservado: 0,
	}

	if err := s.repo.Create(ctx, produto); err != nil {
		s.logger.Error("Erro ao criar produto", zap.Error(err))
		return nil, err
	}

	// Invalidar cache
	s.invalidateCache(ctx, "produtos:*")

	s.logger.Info("Produto criado", zap.String("id", produto.ID.String()))
	return produto, nil
}

// ObterProduto busca produto por ID com cache
func (s *EstoqueService) ObterProduto(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
	// Tentar buscar no cache
	cacheKey := fmt.Sprintf("produtos:%s", id.String())
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var produto domain.Produto
		if err := json.Unmarshal([]byte(cached), &produto); err == nil {
			return &produto, nil
		}
	}

	// Buscar no banco
	produto, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Salvar no cache (TTL 5 minutos)
	produtoJSON, _ := json.Marshal(produto)
	s.cache.Set(ctx, cacheKey, produtoJSON, 5*time.Minute)

	return produto, nil
}

// ListarProdutos lista todos os produtos
func (s *EstoqueService) ListarProdutos(ctx context.Context) ([]domain.Produto, error) {
	// Tentar cache
	cacheKey := "produtos:list"
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var produtos []domain.Produto
		if err := json.Unmarshal([]byte(cached), &produtos); err == nil {
			return produtos, nil
		}
	}

	// Buscar no banco
	produtos, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Cachear resultado
	produtosJSON, _ := json.Marshal(produtos)
	s.cache.Set(ctx, cacheKey, produtosJSON, 2*time.Minute)

	return produtos, nil
}

// BuscarProdutos busca produtos por termo
func (s *EstoqueService) BuscarProdutos(ctx context.Context, query string) ([]domain.Produto, error) {
	return s.repo.Search(ctx, query)
}

// AtualizarProduto atualiza produto
func (s *EstoqueService) AtualizarProduto(ctx context.Context, id uuid.UUID, req domain.AtualizarProdutoRequest) (*domain.Produto, error) {
	produto, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Descricao != "" {
		produto.Descricao = req.Descricao
	}
	if req.Saldo != nil {
		produto.Saldo = *req.Saldo
	}

	if err := s.repo.Update(ctx, produto); err != nil {
		return nil, err
	}

	// Invalidar cache
	s.invalidateCache(ctx, "produtos:*")

	return produto, nil
}

// DeletarProduto deleta produto
func (s *EstoqueService) DeletarProduto(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidar cache
	s.invalidateCache(ctx, "produtos:*")

	return nil
}

// ReservarProdutos reserva múltiplos produtos (com idempotência)
func (s *EstoqueService) ReservarProdutos(ctx context.Context, req domain.ReservarEstoqueRequest) (*domain.ReservaResult, error) {
	// Implementar idempotência
	idempotencyKey := fmt.Sprintf("reserva:%s", req.NotaFiscalID.String())
	exists, err := s.cache.Exists(ctx, idempotencyKey).Result()
	if err != nil {
		s.logger.Warn("Erro ao verificar idempotência", zap.Error(err))
	}
	if exists > 0 {
		// Retornar resultado cacheado
		cached, _ := s.cache.Get(ctx, idempotencyKey).Result()
		var result domain.ReservaResult
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			s.logger.Info("Operação idempotente detectada", zap.String("nota_id", req.NotaFiscalID.String()))
			return &result, nil
		}
	}

	var reservas []domain.ReservaEstoque

	// Processar cada item
	for _, item := range req.Itens {
		// Adquirir lock distribuído para evitar race conditions
		lockKey := fmt.Sprintf("produto:%s", item.ProdutoID.String())
		lockValue, err := s.lock.AcquireLock(ctx, lockKey, 10*time.Second)
		if err != nil {
			s.logger.Error("Falha ao adquirir lock", zap.String("produto_id", item.ProdutoID.String()))
			// Cancelar reservas anteriores
			s.cancelarReservasAnteriores(ctx, reservas)
			return nil, fmt.Errorf("produto está sendo processado simultaneamente")
		}

		// Criar reserva
		reserva := &domain.ReservaEstoque{
			ProdutoID:    item.ProdutoID,
			NotaFiscalID: req.NotaFiscalID,
			Quantidade:   item.Quantidade,
			ExpiresAt:    time.Now().Add(10 * time.Minute),
		}

		if err := s.repo.ReservarEstoque(ctx, reserva); err != nil {
			s.logger.Error("Falha ao reservar produto",
				zap.String("produto_id", item.ProdutoID.String()),
				zap.Error(err),
			)
			// Liberar lock
			s.lock.ReleaseLock(ctx, lockKey, lockValue)
			// Cancelar reservas anteriores
			s.cancelarReservasAnteriores(ctx, reservas)
			return nil, err
		}

		reservas = append(reservas, *reserva)

		// Liberar lock
		s.lock.ReleaseLock(ctx, lockKey, lockValue)
	}

	result := &domain.ReservaResult{
		ReservaID: req.NotaFiscalID,
		Reservas:  reservas,
		Mensagem:  fmt.Sprintf("%d produtos reservados com sucesso", len(reservas)),
	}

	// Salvar no cache para idempotência (TTL 1 hora)
	resultJSON, _ := json.Marshal(result)
	s.cache.Set(ctx, idempotencyKey, resultJSON, time.Hour)

	// Invalidar cache de produtos
	s.invalidateCache(ctx, "produtos:*")

	s.logger.Info("Reservas criadas", zap.Int("quantidade", len(reservas)))
	return result, nil
}

// ConfirmarReserva confirma uma reserva
func (s *EstoqueService) ConfirmarReserva(ctx context.Context, reservaID uuid.UUID) error {
	if err := s.repo.ConfirmarReserva(ctx, reservaID); err != nil {
		s.logger.Error("Erro ao confirmar reserva", zap.Error(err))
		return err
	}

	// Invalidar cache
	s.invalidateCache(ctx, "produtos:*")

	s.logger.Info("Reserva confirmada", zap.String("reserva_id", reservaID.String()))
	return nil
}

// CancelarReserva cancela uma reserva
func (s *EstoqueService) CancelarReserva(ctx context.Context, reservaID uuid.UUID) error {
	if err := s.repo.CancelarReserva(ctx, reservaID); err != nil {
		s.logger.Error("Erro ao cancelar reserva", zap.Error(err))
		return err
	}

	// Invalidar cache
	s.invalidateCache(ctx, "produtos:*")

	s.logger.Info("Reserva cancelada", zap.String("reserva_id", reservaID.String()))
	return nil
}

// BaixarEstoque baixa estoque diretamente (endpoint para serviço de faturamento)
func (s *EstoqueService) BaixarEstoque(ctx context.Context, itens []domain.ItemReserva) error {
	for _, item := range itens {
		// Lock para evitar concorrência
		lockKey := fmt.Sprintf("produto:%s", item.ProdutoID.String())
		lockValue, err := s.lock.AcquireLock(ctx, lockKey, 5*time.Second)
		if err != nil {
			return fmt.Errorf("não foi possível processar: %w", err)
		}

		if err := s.repo.BaixarEstoque(ctx, item.ProdutoID, item.Quantidade); err != nil {
			s.lock.ReleaseLock(ctx, lockKey, lockValue)
			s.logger.Error("Erro ao baixar estoque",
				zap.String("produto_id", item.ProdutoID.String()),
				zap.Error(err),
			)
			return err
		}

		s.lock.ReleaseLock(ctx, lockKey, lockValue)
	}

	// Invalidar cache
	s.invalidateCache(ctx, "produtos:*")

	s.logger.Info("Estoque baixado", zap.Int("itens", len(itens)))
	return nil
}

// VerificarDisponibilidade verifica se há estoque disponível
func (s *EstoqueService) VerificarDisponibilidade(ctx context.Context, produtoID uuid.UUID, quantidade int) (bool, error) {
	produto, err := s.repo.FindByID(ctx, produtoID)
	if err != nil {
		return false, err
	}

	return produto.PodeReservar(quantidade), nil
}

// Helpers

func (s *EstoqueService) cancelarReservasAnteriores(ctx context.Context, reservas []domain.ReservaEstoque) {
	for _, reserva := range reservas {
		if err := s.repo.CancelarReserva(ctx, reserva.ID); err != nil {
			s.logger.Error("Erro ao cancelar reserva no rollback",
				zap.String("reserva_id", reserva.ID.String()),
				zap.Error(err),
			)
		}
	}
}

func (s *EstoqueService) invalidateCache(ctx context.Context, pattern string) {
	keys, err := s.cache.Keys(ctx, pattern).Result()
	if err != nil {
		s.logger.Warn("Erro ao invalidar cache", zap.Error(err))
		return
	}

	if len(keys) > 0 {
		s.cache.Del(ctx, keys...)
	}
}