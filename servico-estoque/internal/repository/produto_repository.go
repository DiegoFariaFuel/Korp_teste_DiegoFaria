// internal/repository/produto_repository.go
package repository

import (
    "context"

    "servico-estoque/internal/domain"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type ProdutoRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*domain.Produto, error)
    FindByCodigo(ctx context.Context, codigo string) (*domain.Produto, error)
    FindAll(ctx context.Context) ([]domain.Produto, error)
    Search(ctx context.Context, query string) ([]domain.Produto, error)
    Create(ctx context.Context, p *domain.Produto) error
    Update(ctx context.Context, p *domain.Produto) error
    Delete(ctx context.Context, id uuid.UUID) error

    ReservarEstoque(ctx context.Context, r *domain.ReservaEstoque) error
    ConfirmarReserva(ctx context.Context, notaID uuid.UUID) error
    CancelarReserva(ctx context.Context, notaID uuid.UUID) error
    BaixarEstoque(ctx context.Context, produtoID uuid.UUID, qtd int) error
}

type produtoRepository struct {
    db *gorm.DB
}

func NewProdutoRepository(db *gorm.DB) ProdutoRepository {
    return &produtoRepository{db: db}
}

func (r *produtoRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Produto, error) {
    var p domain.Produto
    if err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, domain.ErrProdutoNaoEncontrado
        }
        return nil, err
    }
    return &p, nil
}

func (r *produtoRepository) FindByCodigo(ctx context.Context, codigo string) (*domain.Produto, error) {
    var p domain.Produto
    if err := r.db.WithContext(ctx).First(&p, "codigo = ?", codigo).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, err
    }
    return &p, nil
}

func (r *produtoRepository) FindAll(ctx context.Context) ([]domain.Produto, error) {
    var produtos []domain.Produto
    if err := r.db.WithContext(ctx).Find(&produtos).Error; err != nil {
        return nil, err
    }
    return produtos, nil
}

func (r *produtoRepository) Search(ctx context.Context, query string) ([]domain.Produto, error) {
    var produtos []domain.Produto
    q := "%" + query + "%"
    if err := r.db.WithContext(ctx).
        Where("descricao ILIKE ? OR codigo ILIKE ?", q, q).
        Find(&produtos).Error; err != nil {
        return nil, err
    }
    return produtos, nil
}

func (r *produtoRepository) Create(ctx context.Context, p *domain.Produto) error {
    return r.db.WithContext(ctx).Create(p).Error
}

func (r *produtoRepository) Update(ctx context.Context, p *domain.Produto) error {
    return r.db.WithContext(ctx).Updates(p).Error
}

func (r *produtoRepository) Delete(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&domain.Produto{}, "id = ?", id).Error
}

func (r *produtoRepository) ReservarEstoque(ctx context.Context, reserva *domain.ReservaEstoque) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var p domain.Produto
        if err := tx.Clauses(/* lock for update */).WithContext(ctx).
            First(&p, "id = ?", reserva.ProdutoID).Error; err != nil {
            return err
        }

        if !p.PodeReservar(reserva.Quantidade) {
            return domain.ErrEstoqueInsuficiente
        }

        p.Reservado += reserva.Quantidade
        if err := tx.Save(&p).Error; err != nil {
            return err
        }

        return tx.Create(reserva).Error
    })
}

func (r *produtoRepository) ConfirmarReserva(ctx context.Context, notaID uuid.UUID) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var reservas []domain.ReservaEstoque
        if err := tx.Where("nota_fiscal_id = ? AND status = 'PENDENTE'", notaID).
            Find(&reservas).Error; err != nil {
            return err
        }

        for _, r := range reservas {
            var p domain.Produto
            if err := tx.First(&p, "id = ?", r.ProdutoID).Error; err != nil {
                return err
            }

            p.Saldo -= r.Quantidade
            p.Reservado -= r.Quantidade
            if p.Saldo < 0 {
                return domain.ErrSaldoNegativo
            }

            if err := tx.Save(&p).Error; err != nil {
                return err
            }

            r.Status = "CONFIRMADO"
            if err := tx.Save(&r).Error; err != nil {
                return err
            }
        }
        return nil
    })
}

func (r *produtoRepository) CancelarReserva(ctx context.Context, notaID uuid.UUID) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var reservas []domain.ReservaEstoque
        if err := tx.Where("nota_fiscal_id = ? AND status = 'PENDENTE'", notaID).
            Find(&reservas).Error; err != nil {
            return err
        }

        for _, r := range reservas {
            var p domain.Produto
            if err := tx.First(&p, "id = ?", r.ProdutoID).Error; err != nil {
                return err
            }

            p.Reservado -= r.Quantidade
            if p.Reservado < 0 {
                p.Reservado = 0
            }

            if err := tx.Save(&p).Error; err != nil {
                return err
            }

            r.Status = "CANCELADO"
            if err := tx.Save(&r).Error; err != nil {
                return err
            }
        }
        return nil
    })
}

func (r *produtoRepository) BaixarEstoque(ctx context.Context, produtoID uuid.UUID, qtd int) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var p domain.Produto
        if err := tx.Clauses(/* lock for update */).WithContext(ctx).
            First(&p, "id = ?", produtoID).Error; err != nil {
            return err
        }

        if p.Saldo < qtd {
            return domain.ErrEstoqueInsuficiente
        }

        p.Saldo -= qtd
        if err := tx.Save(&p).Error; err != nil {
            return err
        }
        return nil
    })
}