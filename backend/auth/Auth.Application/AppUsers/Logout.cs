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
    public class Logout
    {
        public record LogoutCommand(string RefreshToken);
        public record LogoutReply(bool IsSuccess);

        public class Command : IRequest<Result<LogoutReply>>
        {
            public LogoutCommand Request { get; set; }
        }

        public class Handler(ITokenService tokenService, DataContext dbContext, ITokenRepository tokenCashRepository) : IRequestHandler<Command, Result<LogoutReply>>
        {
            public async Task<Result<LogoutReply>> Handle(Command request, CancellationToken cancellationToken)
            {
                var principal = tokenService.GetPrincipalFromToken(request.Request.RefreshToken);

                string userId = principal.Claims.First(c => c.Type == ClaimTypes.NameIdentifier).Value;

                var user = await dbContext.AppUsers.FirstOrDefaultAsync(x => x.Id.ToString() == userId);
                if (user == null)
                {
                    return Result<LogoutReply>.Failure(new RpcException(new Status(StatusCode.NotFound, "User not found")));
                }

                var isSuccess = await tokenCashRepository.RemoveToken(userId);
                if (!isSuccess)
                {
                    return Result<LogoutReply>.Failure(new RpcException(new Status(StatusCode.NotFound, "Token not found")));
                }

                return Result<LogoutReply>.Success(new LogoutReply(true));
            }
        }

    }   
}
