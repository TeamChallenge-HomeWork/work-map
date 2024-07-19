namespace Auth.GRPC.Redis
{
    public interface ITokenRepository
    {
        Task<string> GetToken(string userId, CancellationToken cancellationToken = default);
        Task<bool> StoreToken(string userId, string token, CancellationToken cancellationToken = default);
    }
}
