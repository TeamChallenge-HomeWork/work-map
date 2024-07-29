namespace Auth.Infrastructure.Redis
{
    public interface ITokenRepository
    {
        public Task<string> GetToken(string userId, CancellationToken cancellationToken = default);
        public Task<bool> StoreToken(string userId, string token, CancellationToken cancellationToken = default);
    }
}
