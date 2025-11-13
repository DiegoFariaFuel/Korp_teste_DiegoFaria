// internal/domain/responses.go
package domain

func NewErrorResponse(code, message string) map[string]any {
    return map[string]any{
        "code":  code,
        "error": message,
    }
}