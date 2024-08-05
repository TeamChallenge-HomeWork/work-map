using Auth.Domain;
using Auth.Infrastructure.Persistance;
using Auth.Infrastructure.Redis;
using Auth.Infrastructure.Services;
using Grpc.Core;
using Microsoft.EntityFrameworkCore;
using Moq;
using System.Security.Claims;
using static Auth.Application.AppUsers.Logout;

namespace Auth.Application.Tests.UnitTests
{
    public class LogoutTests
    {
        private readonly DataContext _context;

        private readonly Mock<ITokenService> _tokenServiceMock;
        private readonly Mock<ITokenRepository> _tokenCashRepositoryMock;

        private readonly Handler _handler;

        private readonly Guid userId = Guid.NewGuid();

        public LogoutTests()
        {
            var options = new DbContextOptionsBuilder<DataContext>()
            .UseInMemoryDatabase(databaseName: "AuthTestDb")
            .Options;

            _context = new DataContext(options);
            var user = new AppUser { Id = userId, Email = "email@test.com", Password = "TestPassw0rd" };

            _context.AppUsers.Add(user);
            _context.SaveChanges();

            _tokenServiceMock = new Mock<ITokenService>();

            _tokenCashRepositoryMock = new Mock<ITokenRepository>();

            _handler = new Handler(_tokenServiceMock.Object, _context, _tokenCashRepositoryMock.Object);
        }

        [Fact]
        public async Task Should_Return_Failure_When_User_Not_Found()
        {
            //Arrange
            var token = "inValidTOken";
            var command = new Command { Request = new LogoutCommand(token) };
            var principal = new ClaimsPrincipal(new ClaimsIdentity(new[] { new Claim(ClaimTypes.NameIdentifier, "1") }));

            _tokenServiceMock.Setup(ts => ts.GetPrincipalFromToken(token)).Returns(principal);

            //Action
            var result = await _handler.Handle(command, CancellationToken.None);

            //Assert
            var exception = Assert.IsType<RpcException>(result.Error);
            Assert.Equal(StatusCode.NotFound, exception.StatusCode);
            Assert.Equal("User not found", exception.Status.Detail);
        }

        //StatusCode.NotFound, "Token not found"
        [Fact]
        public async Task Should_Return_Failure_When_Token_Not_Found()
        {
            //Arrange
            var token = "validTOken";

            var command = new Command { Request = new LogoutCommand(token) };
            var principal = new ClaimsPrincipal(new ClaimsIdentity(new[] { new Claim(ClaimTypes.NameIdentifier, userId.ToString()) }));

            _tokenServiceMock.Setup(ts => ts.GetPrincipalFromToken(token)).Returns(principal);
            _tokenCashRepositoryMock.Setup(tsc => tsc.RemoveToken(token)).ReturnsAsync(false);

            //Act
            var result = await _handler.Handle(command, CancellationToken.None);

            //Assert
            var exception = Assert.IsType<RpcException>(result.Error);
            Assert.Equal(StatusCode.NotFound, exception.StatusCode);
            Assert.Equal("Token not found", exception.Status.Detail);
        }

        //success
        [Fact]
        public async Task Handle_Should_Return_Success()
        {
            //Arrange
            var token = "validTOken";

            var command = new Command { Request = new LogoutCommand(token) };
            var principal = new ClaimsPrincipal(new ClaimsIdentity(new[] { new Claim(ClaimTypes.NameIdentifier, userId.ToString()) }));

            _tokenServiceMock.Setup(ts => ts.GetPrincipalFromToken(token)).Returns(principal);
            _tokenCashRepositoryMock.Setup(tsc => tsc.RemoveToken(userId.ToString())).ReturnsAsync(true);

            //Act
            var result = await _handler.Handle(command, CancellationToken.None);

            //Assert
            Assert.True(result.IsSuccess);
        }

    }
}