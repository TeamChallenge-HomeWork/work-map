namespace Auth.Domain
{
    public class AppUser
    {
        public Guid Id { get; set; }
        public string Email { get; set; }
        public string Password { get; set; }
    }
}
