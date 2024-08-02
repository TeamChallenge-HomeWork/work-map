using Microsoft.Extensions.Caching.Distributed;
using Microsoft.Extensions.Logging;

namespace Auth.Infrastructure.Redis
{
    public class TokenRepository(IDistributedCache cache, ILogger<TokenRepository> logger) : ITokenRepository
    {
        public async Task<string> GetToken(string userId, CancellationToken cancellationToken = default)
        {
            var token = await cache.GetStringAsync(userId, cancellationToken);
            logger.LogInformation("Get token from cache");
            return token;
        }

        public async Task<bool> StoreToken(string userId, string token, CancellationToken cancellationToken = default)
        {
            var options = new DistributedCacheEntryOptions
            {
                AbsoluteExpirationRelativeToNow = TimeSpan.FromDays(15)
            };
            await cache.SetStringAsync(userId, token, options, cancellationToken);
            logger.LogInformation("store token to cache");
            return true;
        }
    }
}
