using Auth.Infrastructure.Persistance;
using Auth.Infrastructure.Redis;
using Auth.Infrastructure.Services;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.IdentityModel.Tokens;
using System.Text;

namespace Auth.Infrastructure
{
    public static class DependencyInjection
    {
        public static IServiceCollection AddInfrastructeServices(this IServiceCollection services, IConfiguration configuration)
        {
            services.AddDbContext<DataContext>(options =>
            {
                string host = configuration["POSTGRES_HOST"]!;
                string port = configuration["POSTGRES_PORT"]!;
                string dbname = configuration["POSTGRES_DB"]!;
                string user = configuration["POSTGRES_USER"]!;
                string password = configuration["POSTGRES_PASSWORD"]!;
                string connStr = $"Server={host};Port={port};Database={dbname};User Id={user};Password={password};Include Error Detail = true";
                options.UseNpgsql(connStr);
            });

            services.AddStackExchangeRedisCache(options =>
            {
                string host = configuration["REDIS_HOST"]!;
                string port = configuration["REDIS_PORT"]!;
                string password = configuration["REDIS_PASSWORD"]!;
                string connStr = $"{host}:{port},password={password}";
                options.Configuration = connStr;
            });
            services.AddScoped<ITokenService, TokenService>();
            services.AddScoped<ITokenRepository, TokenRepository>();

            var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(configuration["JWT_ACCESS_SECRET_KEY"]!));

            services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
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

            return services;
        }
    }
}
