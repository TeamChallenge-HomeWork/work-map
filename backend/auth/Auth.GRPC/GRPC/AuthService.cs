﻿using Auth.Application.AppUsers;
using Grpc.Core;
using MediatR;

namespace Auth.GRPC.Controllers
{
    public class AuthService(ILogger<AuthService> logger, IMediator mediator) : GRPC.AuthService.AuthServiceBase
    {
        public override async Task<RegisterReply> Register(RegisterRequest request, ServerCallContext context)
        {
            logger.LogInformation("Registering new user with email: {Email}", request.Email);
            var command = new Register.Command { Request = new Register.RegisterCommand(request.Email, request.Password) };
            var result = await mediator.Send(command);

            if (!result.IsSuccess)
            {
                throw result.Error;
            }

            return new RegisterReply
            {
                AccessToken = result.Value.AccessToken,
                RefreshToken = result.Value.RefreshToken
            };
        }

        public override async Task<LoginReply> Login(LoginRequest request, ServerCallContext context)
        {
            logger.LogInformation("Loging user with email: {Email}", request.Email);
            var command = new Login.Command { Request = new Login.LoginCommand(request.Email, request.Password) };
            var result = await mediator.Send(command);

            if (!result.IsSuccess)
            {
                throw result.Error;
            }

            return new LoginReply
            {
                AccessToken = result.Value.AccessToken,
                RefreshToken = result.Value.RefreshToken
            };
        }

        public override async Task<RefreshTokenReply> RefreshToken(RefreshTokenRequest request, ServerCallContext context)
        {
            logger.LogInformation("Refreshing token for request with token: {Token}", request.RefreshToken);
            var command = new RefreshToken.Command { Request = new RefreshToken.RefreshTokenCommand(request.RefreshToken) };
            var result = await mediator.Send(command);

            if (!result.IsSuccess)
            {
                throw result.Error;
            }

            return new RefreshTokenReply
            {
                AccessToken = result.Value.AccessToken,
            };
        }

        public override async Task<LogoutReply> Logout(LogoutRequest request, ServerCallContext context)
        {
            logger.LogInformation("Logout user with token: {Token}", request.RefreshToken);
            var command = new Logout.Command { Request = new Logout.LogoutCommand(request.RefreshToken) };
            var result = await mediator.Send(command);

            if (!result.IsSuccess)
            {
                throw result.Error;
            }

            return new LogoutReply
            {
                IsSuccess = result.IsSuccess,
            };
        }
    }
}
