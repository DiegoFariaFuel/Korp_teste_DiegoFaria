// internal/handler/produto_handler.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"servico-estoque/internal/domain"
	"servico-estoque/internal/service"
)

type ProdutoHandler struct {
	service *service.EstoqueService
	logger  *zap.Logger
}

func NewProdutoHandler(service *service.EstoqueService, logger *zap.Logger) *ProdutoHandler {
	return &ProdutoHandler{
		service: service,
		logger:  logger,
	}
}

// ListarProdutos retorna todos os produtos
// GET /api/produtos
func (h *ProdutoHandler) ListarProdutos(c *gin.Context) {
	produtos, err := h.service.ListarProdutos(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, produtos)
}

// ObterProduto retorna um produto por ID
// GET /api/produtos/:id
func (h *ProdutoHandler) ObterProduto(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_ID", "ID inválido"))
		return
	}

	produto, err := h.service.ObterProduto(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, produto)
}

// BuscarProdutos busca produtos por termo
// GET /api/produtos/busca?q=termo
func (h *ProdutoHandler) BuscarProdutos(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusOK, []domain.Produto{})
		return
	}

	produtos, err := h.service.BuscarProdutos(c.Request.Context(), query)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, produtos)
}

// CriarProduto cria um novo produto
// POST /api/produtos
func (h *ProdutoHandler) CriarProduto(c *gin.Context) {
	var req domain.CriarProdutoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	produto, err := h.service.CriarProduto(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, produto)
}

// AtualizarProduto atualiza um produto
// PUT /api/produtos/:id
func (h *ProdutoHandler) AtualizarProduto(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_ID", "ID inválido"))
		return
	}

	var req domain.AtualizarProdutoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	produto, err := h.service.AtualizarProduto(c.Request.Context(), id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, produto)
}

// DeletarProduto deleta um produto
// DELETE /api/produtos/:id
func (h *ProdutoHandler) DeletarProduto(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_ID", "ID inválido"))
		return
	}

	if err := h.service.DeletarProduto(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// VerificarDisponibilidade verifica disponibilidade de estoque
// GET /api/produtos/:id/disponibilidade?quantidade=10
func (h *ProdutoHandler) VerificarDisponibilidade(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_ID", "ID inválido"))
		return
	}

	quantidadeStr := c.Query("quantidade")
	quantidade, err := strconv.Atoi(quantidadeStr)
	if err != nil || quantidade <= 0 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_QUANTITY", "Quantidade inválida"))
		return
	}

	disponivel, err := h.service.VerificarDisponibilidade(c.Request.Context(), id, quantidade)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"disponivel": disponivel})
}

// ReservarEstoque reserva estoque para nota fiscal
// POST /api/produtos/reservar
func (h *ProdutoHandler) ReservarEstoque(c *gin.Context) {
	var req domain.ReservarEstoqueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	result, err := h.service.ReservarProdutos(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ConfirmarReserva confirma uma reserva
// POST /api/produtos/confirmar-reserva
func (h *ProdutoHandler) ConfirmarReserva(c *gin.Context) {
	var req domain.ConfirmarReservaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	if err := h.service.ConfirmarReserva(c.Request.Context(), req.ReservaID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reserva confirmada com sucesso"})
}

// CancelarReserva cancela uma reserva
// POST /api/produtos/cancelar-reserva
func (h *ProdutoHandler) CancelarReserva(c *gin.Context) {
	var req domain.CancelarReservaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	if err := h.service.CancelarReserva(c.Request.Context(), req.ReservaID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reserva cancelada com sucesso"})
}

// BaixarEstoque baixa estoque diretamente (chamado pelo serviço de faturamento)
// POST /api/produtos/baixar
func (h *ProdutoHandler) BaixarEstoque(c *gin.Context) {
	var req domain.BaixarEstoqueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	if err := h.service.BaixarEstoque(c.Request.Context(), req.Itens); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Estoque baixado com sucesso"})
}

// handleError trata erros de forma centralizada
func (h *ProdutoHandler) handleError(c *gin.Context, err error) {
	h.logger.Error("Erro no handler", zap.Error(err))

	switch err {
	case domain.ErrProdutoNaoEncontrado:
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("NOT_FOUND", err.Error()))
	case domain.ErrCodigoDuplicado:
		c.JSON(http.StatusConflict, domain.NewErrorResponse("DUPLICATE_CODE", err.Error()))
	case domain.ErrEstoqueInsuficiente:
		c.JSON(http.StatusConflict, domain.NewErrorResponse("INSUFFICIENT_STOCK", err.Error()))
	case domain.ErrReservaNaoEncontrada:
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("RESERVATION_NOT_FOUND", err.Error()))
	case domain.ErrReservaExpirada:
		c.JSON(http.StatusGone, domain.NewErrorResponse("RESERVATION_EXPIRED", err.Error()))
	case domain.ErrReservaJaConfirmada:
		c.JSON(http.StatusConflict, domain.NewErrorResponse("RESERVATION_ALREADY_CONFIRMED", err.Error()))
	case domain.ErrReservaJaCancelada:
		c.JSON(http.StatusConflict, domain.NewErrorResponse("RESERVATION_ALREADY_CANCELLED", err.Error()))
	case domain.ErrSaldoNegativo:
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("NEGATIVE_BALANCE", err.Error()))
	case domain.ErrQuantidadeInvalida:
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_QUANTITY", err.Error()))
	case domain.ErrDadosInvalidos:
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("INVALID_DATA", err.Error()))
	case domain.ErrOperacaoNaoPermitida:
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("OPERATION_NOT_ALLOWED", err.Error()))
	default:
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("INTERNAL_ERROR", "Erro interno do servidor"))
	}
}

// pkg/lock/distributed_lock.go
package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type DistributedLock struct {
	redis *redis.Client
}

func NewDistributedLock(redis *redis.Client) *DistributedLock {
	return &DistributedLock{redis: redis}
}

// AcquireLock tenta adquirir um lock distribuído
func (l *DistributedLock) AcquireLock(ctx context.Context, resource string, ttl time.Duration) (string, error) {
	lockKey := fmt.Sprintf("lock:%s", resource)
	lockValue := uuid.New().String()

	// SET NX EX - apenas se não existir, com expiração
	success, err := l.redis.SetNX(ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		return "", err
	}

	if !success {
		return "", fmt.Errorf("não foi possível adquirir lock para %s", resource)
	}

	return lockValue, nil
}

// ReleaseLock libera um lock distribuído
func (l *DistributedLock) ReleaseLock(ctx context.Context, resource, lockValue string) error {
	lockKey := fmt.Sprintf("lock:%s", resource)

	// Lua script para garantir atomicidade (só libera se o valor for o mesmo)
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.redis.Eval(ctx, script, []string{lockKey}, lockValue).Result()
	if err != nil {
		return err
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock não pertence a este cliente")
	}

	return nil
}