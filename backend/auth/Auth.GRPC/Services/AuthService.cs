using Auth.GRPC.Data;
using Auth.GRPC.Models;
using Auth.GRPC.Redis;
using Grpc.Core;
using Microsoft.EntityFrameworkCore;
using System.Security.Claims;
using System.Text.RegularExpressions;

namespace Auth.GRPC.Services
{
    public class AuthService(DataContext _context, TokenService tokenService, ITokenRepository tokenCashRepository, ILogger<AuthService> logger) : GRPC.AuthService.AuthServiceBase
    {
        public override async Task<RegisterReply> Register(RegisterRequest request, ServerCallContext context)
        {
            logger.LogInformation("Registering new user with email: {Email}", request.Email);

            if (string.IsNullOrEmpty(request.Email) || !Regex.IsMatch(request.Email, @"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"))
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "Invalid email format"));
            }

            if (string.IsNullOrEmpty(request.Password) || !Regex.IsMatch(request.Password, @"^(?=.*\d)(?=.*[a-z])(?=.*[A-Z]).{4,16}$"))
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "The password must contain at least one digit, one lowercase letter, one uppercase letter, and be between 4 and 16 characters long."));
            }

            if (await _context.AppUsers.AnyAsync(user => user.Email == request.Email))
            {
                throw new RpcException(new Status(StatusCode.AlreadyExists, "User email taken"));
            }

            var newUser = new AppUser
            {
                Email = request.Email,
            };

            string password = BCrypt.Net.BCrypt.HashPassword(request.Password);
            logger.LogInformation("Hashed password for user {Email}: {PasswordHash}", request.Email, password);
            newUser.Password = password;

            _context.AppUsers.Add(newUser);
            await _context.SaveChangesAsync();

            string accessToken = await tokenService.CreateAccessToken(newUser);

            string refreshToken = await tokenService.CreateRefreshToken(newUser);
            await tokenCashRepository.StoreToken(newUser.Id.ToString(), refreshToken);

            logger.LogInformation("User registered successfully with email: {Email}", request.Email);

            return new RegisterReply
            {
                RefreshToken = refreshToken,
                AccessToken = accessToken
            };
        }

        public override async Task<LoginReply> Login(LoginRequest request, ServerCallContext context)
        {
            logger.LogInformation("Loging user with email: {Email}", request.Email);

            var user = await _context.AppUsers.FirstOrDefaultAsync(user => user.Email == request.Email);

            if (user == null)
            {
                throw new RpcException(new Status(StatusCode.Unauthenticated, "User not found"));
            }

            if (!BCrypt.Net.BCrypt.Verify(request.Password, user.Password))
                throw new RpcException(new Status(StatusCode.Unauthenticated, "User not found"));

            string accessToken = await tokenService.CreateAccessToken(user);

            string refreshToken = await tokenService.CreateRefreshToken(user);
            await tokenCashRepository.StoreToken(user.Id.ToString(), refreshToken);

            logger.LogInformation("User loging successfully with email: {Email}", request.Email);

            return new LoginReply
            {
                RefreshToken = refreshToken,
                AccessToken = accessToken
            };
        }

        public override async Task<RefreshTokenReply> RefreshToken(RefreshTokenRequest request, ServerCallContext context)
        {
            logger.LogInformation("Refreshing token for request with token: {Token}", request.RefreshToken);

            var (principal, expiration ) = tokenService.GetPrincipalAndExpirationFromToken(request.RefreshToken);

            if (expiration < DateTime.UtcNow)
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "Refresh token has expired"));
            }

            string userId = principal.Claims.First(c => c.Type == ClaimTypes.NameIdentifier).Value;

            var storedRefreshToken = await tokenCashRepository.GetToken(userId);
            if (storedRefreshToken != request.RefreshToken)
            {
                throw new RpcException(new Status(StatusCode.NotFound, "Refresh token not found"));
            }

            var user = await _context.AppUsers.FirstOrDefaultAsync(x => x.Id.ToString() == userId);
            if (user == null)
            {
                throw new RpcException(new Status(StatusCode.NotFound, "User not found"));
            }

            string accessToken = await tokenService.CreateAccessToken(user!);

            logger.LogInformation("Token refreshed successfully for user with ID: {UserId}", userId);

            return new RefreshTokenReply
            {
                AccessToken = accessToken
            };
        }
    }
}
