using Auth.GRPC.Data;
using Auth.GRPC.Services;
using Microsoft.EntityFrameworkCore;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.IdentityModel.Tokens;
using System.Text;
using Auth.GRPC.Redis;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.
builder.Services.AddGrpc();

builder.Services.AddDbContext<DataContext>(options =>
{
    string connStr = builder.Configuration.GetConnectionString("Postgres")!;
    options.UseNpgsql(connStr);
});

builder.Services.AddStackExchangeRedisCache(options =>
{
    options.Configuration = builder.Configuration.GetConnectionString("Redis");
});

builder.Services.AddAuthorization();

builder.Services.AddScoped<TokenService>();
builder.Services.AddScoped<ITokenRepository, TokenRepository>();


var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(builder.Configuration["AccessTokenKey"]!));

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
