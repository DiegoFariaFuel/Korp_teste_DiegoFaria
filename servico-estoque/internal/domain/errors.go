// internal/domain/errors.go
package domain

import "errors"

var (
    ErrProdutoNaoEncontrado   = errors.New("produto não encontrado")
    ErrCodigoDuplicado        = errors.New("código já existe")
    ErrEstoqueInsuficiente    = errors.New("estoque insuficiente")
    ErrReservaNaoEncontrada   = errors.New("reserva não encontrada")
    ErrReservaExpirada        = errors.New("reserva expirada")
    ErrReservaJaConfirmada    = errors.New("reserva já confirmada")
    ErrReservaJaCancelada     = errors.New("reserva já cancelada")
    ErrSaldoNegativo          = errors.New("saldo não pode ser negativo")
    ErrQuantidadeInvalida     = errors.New("quantidade deve ser maior que zero")
    ErrDadosInvalidos         = errors.New("dados inválidos")
    ErrOperacaoNaoPermitida   = errors.New("operação não permitida")
)