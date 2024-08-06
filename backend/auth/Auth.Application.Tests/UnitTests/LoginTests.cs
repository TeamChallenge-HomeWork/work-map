using Auth.Domain;
using Auth.Infrastructure.Persistance;
using Auth.Infrastructure.Redis;
using Auth.Infrastructure.Services;
using Grpc.Core;
using Microsoft.EntityFrameworkCore;
using Moq;
using static Auth.Application.AppUsers.Login;

namespace Auth.Application.Tests.UnitTests
{
    public class LoginTests
    {
        private readonly DataContext _context;

        private readonly Mock<ITokenService> _tokenServiceMock;
        private readonly Mock<ITokenRepository> _tokenCashRepositoryMock;
        public LoginTests()
        {
            var options = new DbContextOptionsBuilder<DataContext>()
            .UseInMemoryDatabase(databaseName: "AuthTestDb")
            .Options;

            _context = new DataContext(options);

            _tokenServiceMock = new Mock<ITokenService>();

            _tokenCashRepositoryMock = new Mock<ITokenRepository>();
        }

        [Theory]
        [InlineData("correct@email.com", "1ncorPass")]
        [InlineData("incorrect@email.com", "c0rrPass")]
        public async Task Should_Return_Failure_When_Creds_Is_Invalid(string email, string password)
        {
            var hashedPassword = BCrypt.Net.BCrypt.HashPassword(password);

            // Arrange
            var exEemail = "correct@email.com";
            var exPassword = "c0rrPass";
            var exHashedPassword = BCrypt.Net.BCrypt.HashPassword(exPassword);

            var command = new Command { Request = new LoginCommand(email, password) };

            _context.AppUsers.Add(new AppUser { Email = exEemail, Password = exHashedPassword });
            await _context.SaveChangesAsync();

            // Act
            var result = await new Handler(_context, _tokenServiceMock.Object, _tokenCashRepositoryMock.Object).Handle(command);

            // Assert
            var exception = Assert.IsType<RpcException>(result.Error);
            Assert.Equal(StatusCode.Unauthenticated, exception.StatusCode);

        }

        [Fact]
        public async Task Should_Return_Success_When_Login_Is_Valid()
        {
            // Arrange
            var email = "success@example.com";
            var password = "TestPassw0rd";
            var hashedPassword = BCrypt.Net.BCrypt.HashPassword(password);
            var command = new Command { Request = new LoginCommand(email, password) };

            _context.AppUsers.Add(new AppUser { Email = email, Password = hashedPassword });
            await _context.SaveChangesAsync();

            _tokenServiceMock.Setup(ts => ts.CreateAccessToken(It.IsAny<AppUser>())).ReturnsAsync("accessToken");
            _tokenServiceMock.Setup(ts => ts.CreateRefreshToken(It.IsAny<AppUser>())).ReturnsAsync("refreshToken");
            _tokenCashRepositoryMock.Setup(tr => tr.StoreToken(It.IsAny<string>(), It.IsAny<string>())).ReturnsAsync(true);

            // Act
            var result = await new Handler(_context, _tokenServiceMock.Object, _tokenCashRepositoryMock.Object).Handle(command);

            // Assert
            Assert.True(result.IsSuccess);

            Assert.Equal("accessToken", result.Value.AccessToken);
            Assert.Equal("refreshToken", result.Value.RefreshToken);

            _tokenCashRepositoryMock.Verify(tr => tr.StoreToken(It.IsAny<string>(), It.IsAny<string>()), Times.Once);
        }
    }
}
