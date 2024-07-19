using Auth.GRPC.Models;
using Microsoft.EntityFrameworkCore;

namespace Auth.GRPC.Data
{
    public class DataContext : DbContext
    {
        public DataContext(DbContextOptions options) : base(options)
        {
        }

        public DbSet<AppUser> AppUsers { get; set; }
    }
}
