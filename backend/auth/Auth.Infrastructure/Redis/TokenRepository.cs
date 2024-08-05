using Microsoft.Extensions.Logging;
using StackExchange.Redis;

namespace Auth.Infrastructure.Redis
{
    public class TokenRepository : ITokenRepository
    {
        private readonly IConnectionMultiplexer _connectionMultiplexer;
        private readonly ILogger<TokenRepository> _logger;

        private readonly IDatabase _dataContext;

        public TokenRepository(IConnectionMultiplexer connectionMultiplexer, ILogger<TokenRepository> logger)
        {
            _connectionMultiplexer = connectionMultiplexer;
            _logger = logger;

            _dataContext = _connectionMultiplexer.GetDatabase();
        }
        public async Task<string> GetToken(string userId)
        {
            var token = await _dataContext.StringGetAsync(userId);
            _logger.LogInformation("Get token from Redis");
            return token;
        }

        public async Task<bool> RemoveToken(string userId)
        {
            bool isSuccess = await _dataContext.KeyDeleteAsync(userId);
            _logger.LogInformation("Remove token from Redis");
            return isSuccess;
        }

        public async Task<bool> StoreToken(string userId, string token)
        {
            TimeSpan ttl = TimeSpan.FromDays(15);
            await _dataContext.StringSetAsync(userId, token, ttl);
            _logger.LogInformation("store token to cache");
            return true;
        }
    }
}
