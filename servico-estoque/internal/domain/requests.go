// internal/domain/requests.go
package domain

import "github.com/google/uuid"

type CriarProdutoRequest struct {
    Codigo    string `json:"codigo" binding:"required"`
    Descricao string `json:"descricao" binding:"required"`
    Saldo     int    `json:"saldo" binding:"required,gte=0"`
}

type AtualizarProdutoRequest struct {
    Descricao *string `json:"descricao,omitempty"`
    Saldo     *int    `json:"saldo,omitempty"`
}

type ReservarEstoqueRequest struct {
    NotaFiscalID uuid.UUID     `json:"notaFiscalId" binding:"required"`
    Itens        []ItemReserva `json:"itens" binding:"required,dive"`
}

type ConfirmarReservaRequest struct {
    ReservaID uuid.UUID `json:"reservaId" binding:"required"`
}

type CancelarReservaRequest struct {
    ReservaID uuid.UUID `json:"reservaId" binding:"required"`
}

type BaixarEstoqueRequest struct {
    Itens []ItemReserva `json:"itens" binding:"required,dive"`
}