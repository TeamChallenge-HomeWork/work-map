using Auth.Application.AppUsers;
using Auth.Domain;
using Auth.Infrastructure.Persistance;
using Auth.Infrastructure.Redis;
using Auth.Infrastructure.Services;
using Grpc.Core;
using Microsoft.EntityFrameworkCore;
using Moq;
using System;
using static Auth.Application.AppUsers.Register;

namespace Auth.Application.Tests.UnitTests
{
    public class Handlers
    {
        private readonly DataContext _context;

        private readonly Mock<ITokenService> _tokenServiceMock;
        private readonly Mock<ITokenRepository> _tokenCashRepositoryMock;

        public Handlers()
        {
            var options = new DbContextOptionsBuilder<DataContext>()
            .UseInMemoryDatabase(databaseName: "AuthTestDb")
            .Options;

            _context = new DataContext(options);

            _tokenServiceMock = new Mock<ITokenService>();

            _tokenCashRepositoryMock = new Mock<ITokenRepository>();
        }

        //Register
        [Theory]
        [InlineData(null, "TestPassw0rd")]
        [InlineData("", "TestPassw0rd")]
        [InlineData("test@email.com", null)]
        [InlineData("test@email.com", "")]
        public async Task Should_Return_Failure_When_Email_Or_Password_Is_Null_Or_Empty(string email, string password)
        {
            // Arrange
            var command = new Command { Request = new RegisterCommand(email, password) };

            // Act
            var result = await new Handler(_context, _tokenServiceMock.Object, _tokenCashRepositoryMock.Object).Handle(command, CancellationToken.None);

            // Assert
            var exception = Assert.IsType<RpcException>(result.Error);
            Assert.Equal(StatusCode.InvalidArgument, exception.StatusCode);
        }

        [Theory]
        [InlineData("incorrectemail", "TestPassw0rd")]
        [InlineData("test@email.com", "incorrectpassword")]
        public async Task Should_Return_Failure_When_Email_Or_Password_Is_Invalid(string email, string password)
        {
            // Arrange
            var command = new Command { Request = new RegisterCommand(email, password) };

            // Act
            var result = await new Handler(_context, _tokenServiceMock.Object, _tokenCashRepositoryMock.Object).Handle(command, CancellationToken.None);

            // Assert
            var exception = Assert.IsType<RpcException>(result.Error);
            Assert.Equal(StatusCode.InvalidArgument, exception.StatusCode);
        }

        [Fact]
        public async Task Should_Return_Failure_When_Email_Already_Exists()
        {
            // Arrange

            var email = "test@example.com";
            var password = "TestPassw0rd";
            var command = new Command { Request = new RegisterCommand(email, password) };

            _context.AppUsers.Add(new AppUser { Email = email, Password = "hashedPassword" });

            // Act
            var result = await new Handler(_context, _tokenServiceMock.Object, _tokenCashRepositoryMock.Object).Handle(command);

            // Assert
            var exception = Assert.IsType<RpcException>(result.Error);
            Assert.Equal(StatusCode.AlreadyExists, exception.StatusCode);
        }

        [Fact]
        public async Task Should_Return_Success_When_Registration_Is_Valid()
        {
            // Arrange
            var email = "test@example.com";
            var password = "TestPassw0rd";
            var command = new Command { Request = new RegisterCommand(email, password) };

            _tokenServiceMock.Setup(ts => ts.CreateAccessToken(It.IsAny<AppUser>())).ReturnsAsync("accessToken");
            _tokenServiceMock.Setup(ts => ts.CreateRefreshToken(It.IsAny<AppUser>())).ReturnsAsync("refreshToken");
            _tokenCashRepositoryMock.Setup(tr => tr.StoreToken(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<CancellationToken>())).ReturnsAsync(true);

            // Act
            var result = await new Handler(_context, _tokenServiceMock.Object, _tokenCashRepositoryMock.Object).Handle(command);

            // Assert
            Assert.True(result.IsSuccess);

            Assert.Equal("accessToken", result.Value.AccessToken);
            Assert.Equal("refreshToken", result.Value.RefreshToken);

            _tokenCashRepositoryMock.Verify(tr => tr.StoreToken(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<CancellationToken>()), Times.Once);
        }
    }
}
