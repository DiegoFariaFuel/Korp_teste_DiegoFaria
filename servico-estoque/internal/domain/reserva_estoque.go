// internal/domain/reserva_estoque.go
package domain

import (
    "time"

    "github.com/google/uuid"
)

type ReservaEstoque struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    ProdutoID    uuid.UUID `gorm:"type:uuid;not null;index" json:"produtoId"`
    NotaFiscalID uuid.UUID `gorm:"type:uuid;not null;index" json:"notaFiscalId"`
    Quantidade   int       `gorm:"not null" json:"quantidade"`
    Status       string    `gorm:"default:'PENDENTE'" json:"status"`
    ExpiresAt    time.Time `json:"expiresAt"`
    CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type ItemReserva struct {
    ProdutoID  uuid.UUID `json:"produtoId" binding:"required"`
    Quantidade int       `json:"quantidade" binding:"required,gt=0"`
}

type ReservaResult struct {
    ReservaID uuid.UUID       `json:"reservaId"`
    Reservas  []ReservaEstoque `json:"reservas"`
    Mensagem  string          `json:"mensagem"`
}