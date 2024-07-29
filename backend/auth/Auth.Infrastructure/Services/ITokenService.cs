using Auth.Domain;
using System.Security.Claims;

namespace Auth.Infrastructure.Services
{
    public interface ITokenService
    {
        Task<string> CreateAccessToken(AppUser user);
        Task<string> CreateRefreshToken(AppUser user);
        (ClaimsPrincipal, DateTime) GetPrincipalAndExpirationFromToken(string token);
    }
}
