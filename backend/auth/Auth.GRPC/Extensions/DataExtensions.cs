using Auth.GRPC.Data;
using Microsoft.EntityFrameworkCore;

namespace Auth.GRPC.Extensions
{
    public static class Extensions
    {
        public static IApplicationBuilder UseMigration(this IApplicationBuilder app)
        {
            using var scope = app.ApplicationServices.CreateScope();
            using var dbContext = scope.ServiceProvider.GetRequiredService<DataContext>();
            dbContext.Database.Migrate();

            return app;
        }
    }
}
