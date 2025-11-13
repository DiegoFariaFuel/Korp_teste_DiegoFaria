using Microsoft.EntityFrameworkCore;

var builder = WebApplication.CreateBuilder(args);

// === SWAGGER ===
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

// === POSTGRES ===
builder.Services.AddDbContext<AppDbContext>(options =>
    options.UseNpgsql("Host=postgres;Database=faturamento;Username=postgres;Password=postgres"));

builder.Services.AddControllers();

var app = builder.Build();

// === SWAGGER UI (DESENVOLVIMENTO) ===
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

// === MAPEAR CONTROLLERS ===
app.MapControllers();

// === MIGRAÇÃO AUTOMÁTICA COM LOG E TRATAMENTO DE ERRO ===
using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
    try
    {
        Console.WriteLine("Aplicando migrações no banco de dados...");
        db.Database.Migrate();
        Console.WriteLine("Migrações aplicadas com sucesso!");
    }
    catch (Exception ex)
    {
        Console.WriteLine($"ERRO ao aplicar migrações: {ex.Message}");
        Console.WriteLine($"Stack Trace: {ex.StackTrace}");
        throw; // Para o container não subir com DB quebrado
    }
}

app.Run();

// ==================================================================
// ======================= DbContext e Modelos ======================
// ==================================================================

public class AppDbContext : DbContext
{
    public AppDbContext(DbContextOptions<AppDbContext> options) : base(options) { }

    // ENTIDADES COM DbSet (OBRIGATÓRIO!)
    public DbSet<NotaFiscal> NotasFiscais => Set<NotaFiscal>();
    public DbSet<NotaFiscalItem> NotaFiscalItens => Set<NotaFiscalItem>(); // CORRIGIDO!

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        // Configuração da chave primária do item
        modelBuilder.Entity<NotaFiscalItem>()
            .HasKey(i => i.Id);

        modelBuilder.Entity<NotaFiscalItem>()
            .Property(i => i.Id)
            .ValueGeneratedOnAdd();

        // Relacionamento 1:N (NotaFiscal -> Itens)
        modelBuilder.Entity<NotaFiscalItem>()
            .HasOne<NotaFiscal>()
            .WithMany(n => n.Itens)
            .HasForeignKey(i => i.NotaFiscalId);
    }
}

// ======================= ENTIDADES ======================

public class NotaFiscal
{
    public Guid Id { get; set; } = Guid.NewGuid();
    public string Numero { get; set; } = "";
    public Guid ClienteId { get; set; }
    public decimal ValorTotal { get; set; }
    public string Status { get; set; } = "Rascunho";
    public DateTime DataEmissao { get; set; } = DateTime.UtcNow;
    public DateTime? DataImpressao { get; set; }
    public List<NotaFiscalItem> Itens { get; set; } = new();
}

public class NotaFiscalItem
{
    public int Id { get; set; }
    public Guid NotaFiscalId { get; set; }
    public Guid ProdutoId { get; set; }
    public string DescricaoProduto { get; set; } = "";
    public int Quantidade { get; set; }
    public decimal ValorUnitario { get; set; }
}