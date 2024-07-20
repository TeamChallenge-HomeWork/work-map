using System.ComponentModel.DataAnnotations;

namespace Auth.GRPC.Models
{
    public class AppUser
    {
        public Guid Id { get; set; }
        public string Email { get; set; }
        public string Password { get; set; }
    }
}
