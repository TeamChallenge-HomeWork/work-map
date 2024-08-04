using Auth.Application.Core;
using Auth.Infrastructure.Persistance;
using Auth.Infrastructure.Redis;
using Auth.Infrastructure.Services;
using Grpc.Core;
using MediatR;
using Microsoft.EntityFrameworkCore;

namespace Auth.Application.AppUsers
{
    public class Login
    {
        public record LoginCommand(string Email, string Password);
        public record LoginResult(string AccessToken, string RefreshToken);
        public class Command : IRequest<Result<LoginResult>>
        {
            public LoginCommand Request { get; set; }
        }

        public class Handler : IRequestHandler<Command, Result<LoginResult>>
        {

            private readonly DataContext _context;
            private readonly ITokenService _tokenService;
            private readonly ITokenRepository _tokenCashRepository;

            public Handler(DataContext context, ITokenService tokenService, ITokenRepository tokenCashRepository)
            {
                _context = context;
                _tokenService = tokenService;
                _tokenCashRepository = tokenCashRepository;
            }
            public async Task<Result<LoginResult>> Handle(Command command, CancellationToken cancellationToken = default)
            {
                var user = await _context.AppUsers.FirstOrDefaultAsync(user => user.Email == command.Request.Email);

                if (user == null)
                {
                    return Result<LoginResult>.Failure(new RpcException(new Status(StatusCode.Unauthenticated, "User not found")));
                }

                if (!BCrypt.Net.BCrypt.Verify(command.Request.Password, user.Password))
                    return Result<LoginResult>.Failure(new RpcException(new Status(StatusCode.Unauthenticated, "User not found")));

                string accessToken = await _tokenService.CreateAccessToken(user);

                string refreshToken = await _tokenService.CreateRefreshToken(user);
                await _tokenCashRepository.StoreToken(user.Id.ToString(), refreshToken);

                return Result<LoginResult>.Success(new LoginResult(accessToken, refreshToken));
            }
        }
    }
}
