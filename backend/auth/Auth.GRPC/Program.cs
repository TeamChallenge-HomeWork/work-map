using Auth.GRPC.Data;
using Auth.GRPC.Services;
using Microsoft.EntityFrameworkCore;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.IdentityModel.Tokens;
using System.Text;
using Auth.GRPC.Redis;
using Auth.GRPC.Extensions;

var builder = WebApplication.CreateBuilder(args);
builder.SetConfiguration();

builder.Services.AddGrpc();

builder.Services.AddDbContext<DataContext>(options =>
{
    string host = builder.Configuration["POSTGRES_HOST"]!;
    string port = builder.Configuration["POSTGRES_PORT"]!;
    string dbname = builder.Configuration["POSTGRES_DB"]!;
    string user = builder.Configuration["POSTGRES_USER"]!;
    string password = builder.Configuration["POSTGRES_PASSWORD"]!;
    string connStr = $"Server={host};Port={port};Database={dbname};User Id={user};Password={password};Include Error Detail = true";
    options.UseNpgsql(connStr);
});

builder.Services.AddStackExchangeRedisCache(options =>
{
    string host = builder.Configuration["REDIS_HOST"]!;
    string port = builder.Configuration["REDIS_PORT"]!;
    string password = builder.Configuration["REDIS_PASSWORD"]!;
    string connStr = $"{host}:{port},password={password}";
    options.Configuration = connStr;
});

builder.Services.AddAuthorization();

builder.Services.AddScoped<TokenService>();
builder.Services.AddScoped<ITokenRepository, TokenRepository>();


var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(builder.Configuration["JWT_ACCESS_SECRET_KEY"]!));

builder.Services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
.AddJwtBearer(opt =>
    {
        opt.TokenValidationParameters = new TokenValidationParameters
        {
            ValidateIssuerSigningKey = true,
            IssuerSigningKey = key,
            ValidateIssuer = false,
            ValidateAudience = false
        };
    });

var app = builder.Build();

// Configure the HTTP request pipeline.

app.UseMigration();

app.UseAuthentication();
app.UseAuthorization();

app.MapGrpcService<AuthService>();

app.Run();
