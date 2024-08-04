namespace Auth.Infrastructure.Redis
{
    public interface ITokenRepository
    {
        public Task<string> GetToken(string userId);
        public Task<bool> StoreToken(string userId, string token);
    }
}
