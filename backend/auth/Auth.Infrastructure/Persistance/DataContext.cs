﻿using Auth.Domain;
using Microsoft.EntityFrameworkCore;

namespace Auth.Infrastructure.Persistance
{
    public class DataContext : DbContext
    {
        public DataContext(DbContextOptions options) : base(options)
        {
        }

        public DbSet<AppUser> AppUsers { get; set; }
    }
}
