using Auth.Application.AppUsers;
using Auth.Domain;
using Auth.Infrastructure.Persistance;
using Auth.Infrastructure.Redis;
using Auth.Infrastructure.Services;
using Grpc.Core;
using Microsoft.EntityFrameworkCore;
using Moq;
using System.Security.Claims;

namespace Auth.Application.Tests.UnitTests
{
    public class RefreshTokenTests
    {
        private readonly DataContext _context;

        private readonly Mock<ITokenService> _tokenServiceMock;
        private readonly Mock<ITokenRepository> _tokenCashRepositoryMock;

        private string UserId => Guid.NewGuid().ToString();
        private string RefreshToken = "refreshToken";
        private string AccessToken = "accessToken";
        public RefreshTokenTests()
        {
            var options = new DbContextOptionsBuilder<DataContext>()
            .UseInMemoryDatabase(databaseName: "AuthTestDb")
            .Options;

            _context = new DataContext(options);

            _tokenServiceMock = new Mock<ITokenService>();
            _tokenServiceMock.Setup(ts => ts.CreateAccessToken(It.IsAny<AppUser>())).ReturnsAsync(AccessToken);

            _tokenCashRepositoryMock = new Mock<ITokenRepository>();
            _tokenCashRepositoryMock.Setup(tr => tr.GetToken(It.IsAny<string>(), It.IsAny<CancellationToken>())).ReturnsAsync(RefreshToken);
        }

        private async Task<AppUser> CreateUserAsync(string userId)
        {
            var user = new AppUser { Id = Guid.Parse(userId), Email = "test@example.com", Password = "hashedPassword" };
            _context.AppUsers.Add(user);
            await _context.SaveChangesAsync();
            return user;
        }

/*        [Fact]
        public async Task Should_Return_Failure_When_RefreshToken_Is_Expired()
        {
            // Arrange
            var userId = UserId;
            await CreateUserAsync(userId);

            var claims = new List<Claim> { new Claim(ClaimTypes.NameIdentifier, userId) };
            var principal = new ClaimsPrincipal(new ClaimsIdentity(claims));
            _tokenServiceMock.Setup(ts => ts.GetPrincipalAndExpirationFromToken(It.IsAny<string>()))
                .Returns((principal, DateTime.UtcNow.AddMinutes(-10)));

            // Act
            var command = new RefreshToken.Command { Request = new RefreshToken.RefreshTokenCommand(RefreshToken) };
            var handler = new RefreshToken.Handler(_context, _tokenServiceMock.Object, _tokenCashRepositoryMock.Object);

            var result = await handler.Handle(command, CancellationToken.None);

            // Assert
            var exception = Assert.IsType<RpcException>(result.Error);
            Assert.Equal(StatusCode.InvalidArgument, exception.StatusCode);
        }*/

        [Fact]
        public async Task Should_Returen_Failure_When_User_Is_NotFound()
        {
            // Arrange
            var userId = Guid.NewGuid().ToString();

            var claims = new List<Claim> { new Claim(ClaimTypes.NameIdentifier, userId) };
            var principal = new ClaimsPrincipal(new ClaimsIdentity(claims));
            _tokenServiceMock.Setup(ts => ts.GetPrincipalFromToken(It.IsAny<string>())).Returns((principal));
            
            // Act
            var command = new RefreshToken.Command { Request = new RefreshToken.RefreshTokenCommand(RefreshToken) };
            var handler = new RefreshToken.Handler(_context, _tokenServiceMock.Object, _tokenCashRepositoryMock.Object);

            var result = await handler.Handle(command, CancellationToken.None);

            var exception = Assert.IsType<RpcException>(result.Error);
            Assert.Equal(StatusCode.NotFound, exception.StatusCode);
        }

        [Fact]
        public async Task Should_Returen_Failure_When_Token_Is_NotFound()
        {
            // Arrange
            var userId = UserId;
            await CreateUserAsync(userId);

            _tokenCashRepositoryMock.Setup(tcr => tcr.GetToken(It.IsAny<string>(), It.IsAny<CancellationToken>())).ReturnsAsync((string)null);
            _tokenServiceMock.Setup(tc => tc.CreateAccessToken(It.IsAny<AppUser>())).ReturnsAsync(AccessToken);

            var claims = new List<Claim> { new Claim(ClaimTypes.NameIdentifier, userId) };
            var principal = new ClaimsPrincipal(new ClaimsIdentity(claims));
            _tokenServiceMock.Setup(ts => ts.GetPrincipalFromToken(It.IsAny<string>())).Returns((principal));

            // Act
            var command = new RefreshToken.Command { Request = new RefreshToken.RefreshTokenCommand(RefreshToken) };
            var handler = new RefreshToken.Handler(_context, _tokenServiceMock.Object, _tokenCashRepositoryMock.Object);

            var result = await handler.Handle(command, CancellationToken.None);

            // Assert
            var exception = Assert.IsType<RpcException>(result.Error);
            Assert.Equal(StatusCode.NotFound, exception.StatusCode);
        }

        [Fact]
        public async Task Should_Return_Success_When_RefreshToken_Is_Valid()
        {
            // Arrange
            var userId = UserId;
            await CreateUserAsync(userId);

            var claims = new List<Claim> { new Claim(ClaimTypes.NameIdentifier, userId) };
            var principal = new ClaimsPrincipal(new ClaimsIdentity(claims));
            _tokenServiceMock.Setup(ts => ts.GetPrincipalFromToken(It.IsAny<string>()))
                .Returns((principal));

            // Act
            var command = new RefreshToken.Command { Request = new RefreshToken.RefreshTokenCommand(RefreshToken) };
            var handler = new RefreshToken.Handler(_context, _tokenServiceMock.Object, _tokenCashRepositoryMock.Object);

            var result = await handler.Handle(command, CancellationToken.None);

            // Assert
            Assert.Equal(AccessToken, result.Value.AccessToken);
            _tokenCashRepositoryMock.Verify(tr => tr.GetToken(userId, It.IsAny<CancellationToken>()), Times.Once);
            _tokenServiceMock.Verify(ts => ts.CreateAccessToken(It.IsAny<AppUser>()), Times.Once);

        }
    }
}
