using Microsoft.Extensions.Logging;
using StackExchange.Redis;

namespace Auth.Infrastructure.Redis
{
    public class TokenRepository(IConnectionMultiplexer connectionMultiplexer, ILogger<TokenRepository> logger) : ITokenRepository
    {
        public async Task<string> GetToken(string userId)
        {
            var db = connectionMultiplexer.GetDatabase();
            var token = await db.StringGetAsync(userId);
            logger.LogInformation("Get token from Redis");
            return token;
        }

        public async Task<bool> StoreToken(string userId, string token)
        {
            var db = connectionMultiplexer.GetDatabase();

            TimeSpan ttl = TimeSpan.FromDays(15);
            await db.StringSetAsync(userId, token, ttl);
            logger.LogInformation("store token to cache");
            return true;
        }
    }
}
