// internal/domain/produto.go
package domain

import (
    "time"

    "github.com/google/uuid"
)

type Produto struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    Codigo      string    `gorm:"uniqueIndex;not null" json:"codigo"`
    Descricao   string    `gorm:"not null" json:"descricao"`
    Saldo       int       `gorm:"not null" json:"saldo"`
    Reservado   int       `gorm:"default:0" json:"reservado"`
    CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (p *Produto) PodeReservar(quantidade int) bool {
    return p.Saldo-p.Reservado >= quantidade
}