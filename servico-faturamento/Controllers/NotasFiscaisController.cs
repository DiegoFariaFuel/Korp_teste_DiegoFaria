using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

[ApiController]
[Route("api/notas-fiscais")] // AQUI ESTAVA FALTANDO O HÍFEN!!!
public class NotasFiscaisController : ControllerBase
{
    private readonly AppDbContext _context;

    public NotasFiscaisController(AppDbContext context)
    {
        _context = context;
    }

    [HttpGet]
    public async Task<ActionResult<IEnumerable<NotaFiscal>>> GetNotas()
    {
        return await _context.NotasFiscais.Include(n => n.Itens).ToListAsync();
    }

    [HttpPost]
    public async Task<ActionResult<NotaFiscal>> CriarNota(CriarNotaRequest request)
    {
        var nota = new NotaFiscal
        {
            ClienteId = request.ClienteId,
            Numero = $"NF{DateTime.Now:yyyyMMddHHmmss}",
            Itens = request.Itens.Select(i => new NotaFiscalItem
            {
                ProdutoId = i.ProdutoId,
                DescricaoProduto = i.DescricaoProduto,
                Quantidade = i.Quantidade,
                ValorUnitario = i.ValorUnitario
            }).ToList()
        };

        nota.ValorTotal = nota.Itens.Sum(i => i.Quantidade * i.ValorUnitario);

        _context.NotasFiscais.Add(nota);
        await _context.SaveChangesAsync();

        return CreatedAtAction(nameof(GetNotas), new { id = nota.Id }, nota);
    }

    [HttpPost("{id}/imprimir")]
    public async Task<ActionResult<NotaFiscal>> ImprimirNota(Guid id)
    {
        var nota = await _context.NotasFiscais
            .Include(n => n.Itens)
            .FirstOrDefaultAsync(n => n.Id == id);

        if (nota == null) return NotFound();

        nota.Status = "Impressa";
        nota.DataImpressao = DateTime.UtcNow;
        await _context.SaveChangesAsync();

        return Ok(new { nota, mensagem = "Nota impressa com sucesso!" });
    }
}

public class CriarNotaRequest
{
    public Guid ClienteId { get; set; }
    public List<CriarItemRequest> Itens { get; set; } = new();
}

public class CriarItemRequest
{
    public Guid ProdutoId { get; set; }
    public string DescricaoProduto { get; set; } = "";
    public int Quantidade { get; set; }
    public decimal ValorUnitario { get; set; }
}
