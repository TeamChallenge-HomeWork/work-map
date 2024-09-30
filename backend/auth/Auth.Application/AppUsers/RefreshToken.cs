using Auth.Application.Core;
using Auth.Infrastructure.Persistance;
using Auth.Infrastructure.Redis;
using Auth.Infrastructure.Services;
using Grpc.Core;
using MediatR;
using Microsoft.EntityFrameworkCore;
using System.Security.Claims;

namespace Auth.Application.AppUsers
{
    public class RefreshToken
    {
        public record RefreshTokenCommand(string RefreshToken);
        public record RefreshTokenResult(string AccessToken);
        public class Command : IRequest<Result<RefreshTokenResult>>
        {
            public RefreshTokenCommand Request { get; set; }
        }

        public class Handler : IRequestHandler<Command, Result<RefreshTokenResult>>
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
            public async Task<Result<RefreshTokenResult>> Handle(Command command, CancellationToken cancellationToken)
            {
                var principal = _tokenService.GetPrincipalFromToken(command.Request.RefreshToken);

                string userId = principal.Claims.First(c => c.Type == ClaimTypes.NameIdentifier).Value;

                var user = await _context.AppUsers.FirstOrDefaultAsync(x => x.Id.ToString() == userId);
                if (user == null)
                {
                    return Result<RefreshTokenResult>.Failure(new RpcException(new Status(StatusCode.NotFound, "User not found")));
                }

                var storedRefreshToken = await _tokenCashRepository.GetToken(userId);
                if (storedRefreshToken != command.Request.RefreshToken)
                {
                    return Result<RefreshTokenResult>.Failure(new RpcException(new Status(StatusCode.NotFound, "Refresh token not found")));
                }

                string accessToken = await _tokenService.CreateAccessToken(user!);

                return Result<RefreshTokenResult>.Success(new RefreshTokenResult(accessToken));
            }
        }
    }
}
