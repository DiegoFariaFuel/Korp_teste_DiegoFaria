FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS base
WORKDIR /app
EXPOSE 8080

FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src
COPY ["servico-faturamento.csproj", "."]
RUN dotnet restore "./servico-faturamento.csproj"
COPY . .
WORKDIR "/src"
RUN dotnet build "./servico-faturamento.csproj" -c Release -o /app/build

FROM build AS publish
RUN dotnet publish "./servico-faturamento.csproj" -c Release -o /app/publish

FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .
ENTRYPOINT ["dotnet", "servico-faturamento.dll"]
