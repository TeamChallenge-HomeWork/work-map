using Auth.GRPC.Models;
using Grpc.Core;
using Microsoft.IdentityModel.Tokens;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;

namespace Auth.GRPC.Services
{
    public class TokenService(IConfiguration config)
    {
        public async Task<string> CreateAccessToken(AppUser user)
        {
            var claims = new List<Claim>
            {
                new Claim(ClaimTypes.Email, user.Email)
            };

            var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(config["JWT_ACCESS_SECRET_KEY"]!));
            var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha512Signature);

            var tokenDescriptor = new SecurityTokenDescriptor
            {
                Subject = new ClaimsIdentity(claims),
                Expires = DateTime.UtcNow.AddMinutes(15),
                SigningCredentials = creds
            };

            var tokenHandler = new JwtSecurityTokenHandler();

            var token = tokenHandler.CreateToken(tokenDescriptor);

            return tokenHandler.WriteToken(token);
        }

        public async Task<string> CreateRefreshToken(AppUser user)
        {
            var claims = new List<Claim>
            {
                new Claim(ClaimTypes.NameIdentifier, user.Id.ToString())
            };

            var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(config["JWT_REFRESH_SECRET_KEY"]!));
            var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha512Signature);

            var tokenDescriptor = new SecurityTokenDescriptor
            {
                Subject = new ClaimsIdentity(claims),
                Expires = DateTime.UtcNow.AddDays(15),
                SigningCredentials = creds
            };

            var tokenHandler = new JwtSecurityTokenHandler();

            var token = tokenHandler.CreateToken(tokenDescriptor);

            return tokenHandler.WriteToken(token);
        }

        public (ClaimsPrincipal, DateTime) GetPrincipalAndExpirationFromToken(string token)
        {
            var tokenHandler = new JwtSecurityTokenHandler();
            var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(config["JWT_REFRESH_SECRET_KEY"]!));

            try
            {
                var principal = tokenHandler.ValidateToken(token, new TokenValidationParameters
                {
                    ValidateIssuer = false,
                    ValidateAudience = false,
                    ValidateLifetime = true, // Ensure lifetime validation is enabled
                    ValidateIssuerSigningKey = true,
                    IssuerSigningKey = key,
                }, out SecurityToken securityToken);

                if (!(securityToken is JwtSecurityToken jwtSecurityToken))
                {
                    throw new SecurityTokenException("Invalid token");
                }

                return (principal, jwtSecurityToken.ValidTo);
            }
            catch (SecurityTokenSignatureKeyNotFoundException ex)
            {
                throw new RpcException(new Status(StatusCode.Unauthenticated, "Token signature key not found"));
            }
            catch (Exception ex)
            {
                throw new RpcException(new Status(StatusCode.Internal, "Internal server error"));
            }
        }

    }
}
