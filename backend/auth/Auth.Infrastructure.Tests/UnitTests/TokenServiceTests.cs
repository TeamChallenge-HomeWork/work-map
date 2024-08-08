using Auth.Domain;
using Auth.Infrastructure.Services;
using Grpc.Core;
using Microsoft.Extensions.Configuration;
using Microsoft.IdentityModel.Tokens;
using Moq;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;

namespace Auth.Infrastructure.Tests.UnitTests
{
    public class TokenServiceTests
    {
        private readonly Mock<IConfiguration> _configutationMock;
        private readonly TokenService _tokenService;

        private readonly AppUser _user;
        public TokenServiceTests()
        {
            _configutationMock = new Mock<IConfiguration>();
            _configutationMock.Setup(c => c["JWT_ACCESS_SECRET_KEY"]).Returns("dnmiz6tknysBDLZCneUrEUX8u1DHjEMbq6qoLb4D3QFwwyABkfBcp2N5R6m688hB");
            _configutationMock.Setup(c => c["JWT_REFRESH_SECRET_KEY"]).Returns("nmiz6tknysBDLZCneUrEUX8u1DHjEMbq6qoLb4D3Qdnmiz6tknysBDLZCneUrEUss");

            _user = new AppUser() { Id = Guid.NewGuid(), Email = "test@gmail.com" };

            _tokenService = new TokenService(_configutationMock.Object);
        }

        [Fact]
        public async Task Should_Succesfuly_Create_Access_Token()
        {
            // Act
            var token = await _tokenService.CreateAccessToken(_user);

            //Assert
            Assert.False(string.IsNullOrEmpty(token));
        }

        [Fact]
        public async Task Should_Succesfuly_Create_Refresh_Token()
        {
            // Act
            var token = await _tokenService.CreateRefreshToken(_user);

            //Assert
            Assert.False(string.IsNullOrEmpty(token));
        }

        [Fact]
        public async Task Should_Throw_Expired_Exception()
        {
            //Arrange
            var claims = new List<Claim>()
            {
                new Claim(ClaimTypes.Email, _user.Email)
            };

            var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes("nmiz6tknysBDLZCneUrEUX8u1DHjEMbq6qoLb4D3Qdnmiz6tknysBDLZCneUrEUss"));
            var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha512Signature);

            var tokenDescriptor = new SecurityTokenDescriptor
            {
                Subject = new ClaimsIdentity(claims),
                Expires = DateTime.UtcNow.AddMilliseconds(10),
                SigningCredentials = creds
            };

            var tokenHandler = new JwtSecurityTokenHandler();

            var token = tokenHandler.CreateToken(tokenDescriptor);

            var stoken = tokenHandler.WriteToken(token);

            await Task.Delay(20);

            // Act & Assert
            var exception = Assert.Throws<RpcException>(() => _tokenService.GetPrincipalFromToken(stoken));
            Assert.Equal(StatusCode.Unauthenticated, exception.StatusCode);
            Assert.Equal("Token has expired", exception.Status.Detail);
        }

        [Fact]
        public async Task Should_Throw_Invalid_Token_Signature_Exception()
        {
            //Arrange
            var token = await _tokenService.CreateRefreshToken(_user);
            _configutationMock.Setup(c => c["JWT_REFRESH_SECRET_KEY"]).Returns("inncorrectHashinncorrectHashinncorrectHashinncorrectHash");

            // Act & Assert
            var exception = Assert.Throws<RpcException>(() => _tokenService.GetPrincipalFromToken(token));
            Assert.Equal(StatusCode.Unauthenticated, exception.StatusCode);
            Assert.Equal("Invalid token signature", exception.Status.Detail);
        }

        [Fact]
        public void Should_Throw_Invalid_Token_Format_Exception()
        {
            // Arrange
            var invalidToken = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJuYW1laWQiOiIxZDY3ZTQxMS0zYzgyLTQ4MTMtYjFlNi0zZmI2NDAxOWMxNWYiLCJuYmYiOjE3MjI1MTcwOTYsImV4cCI6MTcyMzgxMzA5NiwiaWF0IjoxNzIyNTE3MDk2fQ";

            // Act & Assert
            var exception = Assert.Throws<RpcException>(() => _tokenService.GetPrincipalFromToken(invalidToken));
            Assert.Equal(StatusCode.Unauthenticated, exception.StatusCode);
            Assert.Equal("Invalid token format", exception.Status.Detail);
        }

        [Fact]
        public async Task Should_Succesfuly_Get_Principal_From_Token()
        {
            // Act
            var token = await _tokenService.CreateRefreshToken(_user);

            var result = _tokenService.GetPrincipalFromToken(token);
            //Assert
            string userId = result.Claims.First(c => c.Type == ClaimTypes.NameIdentifier).Value;
            Assert.Equal(_user.Id.ToString(), userId);
        }
    }
}
