using Auth.Application.Core;
using Auth.Domain;
using Auth.Infrastructure.Persistance;
using Auth.Infrastructure.Redis;
using Auth.Infrastructure.Services;
using Grpc.Core;
using MediatR;
using Microsoft.EntityFrameworkCore;
using System.Text.RegularExpressions;

namespace Auth.Application.AppUsers
{
    public class Register
    {
        public record RegisterCommand(string Email, string Password);
        public record RegisterResult(string AccessToken, string RefreshToken);
        public class Command : IRequest<Result<RegisterResult>>
        {
            public RegisterCommand Request { get; set; }
        }

        public class Handler : IRequestHandler<Command, Result<RegisterResult>>
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
            public async Task<Result<RegisterResult>> Handle(Command command, CancellationToken cancellationToken = default)
            {
                if (string.IsNullOrEmpty(command.Request.Email) || !Regex.IsMatch(command.Request.Email, @"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"))
                {
                    return Result<RegisterResult>.Failure(new RpcException(new Status(StatusCode.InvalidArgument, "Invalid email format")));
                }

                if (string.IsNullOrEmpty(command.Request.Password) || !Regex.IsMatch(command.Request.Password, @"^(?=.*\d)(?=.*[a-z])(?=.*[A-Z]).{4,16}$"))
                {
                    return Result<RegisterResult>.Failure(new RpcException(new Status(StatusCode.InvalidArgument, "The password must contain at least one digit, one lowercase letter, one uppercase letter, and be between 4 and 16 characters long.")));
                }

                if (await _context.AppUsers.AnyAsync(user => user.Email == command.Request.Email))
                {
                    return Result<RegisterResult>.Failure(new RpcException(new Status(StatusCode.AlreadyExists, "User email taken")));
                }

                var newUser = new AppUser
                {
                    Email = command.Request.Email,
                };

                string password = BCrypt.Net.BCrypt.HashPassword(command.Request.Password);
                newUser.Password = password;

                _context.AppUsers.Add(newUser);
                await _context.SaveChangesAsync();

                string accessToken = await _tokenService.CreateAccessToken(newUser);

                string refreshToken = await _tokenService.CreateRefreshToken(newUser);
                await _tokenCashRepository.StoreToken(newUser.Id.ToString(), refreshToken, cancellationToken);

                return Result<RegisterResult>.Success(new RegisterResult(accessToken, refreshToken));
            }
        }
    }
}
