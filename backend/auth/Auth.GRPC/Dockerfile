FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS base
WORKDIR /app
EXPOSE 8080

FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

ARG BUILD_CONFIGURATION=Release

COPY ["backend/auth/Auth.GRPC/Auth.GRPC.csproj", "backend/auth/Auth.GRPC/"] 
COPY ["backend/auth/Auth.Application/Auth.Application.csproj", "backend/auth/Auth.Application/"]
COPY ["backend/auth/Auth.Domain/Auth.Domain.csproj", "backend/auth/Auth.Domain/"]
COPY ["backend/auth/Auth.Infrastructure/Auth.Infrastructure.csproj", "backend/auth/Auth.Infrastructure/"]
RUN dotnet restore "./backend/auth/Auth.GRPC/Auth.GRPC.csproj"

COPY . .
WORKDIR "/src/backend/auth/Auth.GRPC"
RUN dotnet build "./Auth.GRPC.csproj" -c $BUILD_CONFIGURATION -o /app/build

FROM build AS publish
ARG BUILD_CONFIGURATION=Release
RUN dotnet publish "./Auth.GRPC.csproj" -c $BUILD_CONFIGURATION -o /app/publish /p:UseAppHost=false

FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .
ENTRYPOINT ["dotnet", "Auth.GRPC.dll"]