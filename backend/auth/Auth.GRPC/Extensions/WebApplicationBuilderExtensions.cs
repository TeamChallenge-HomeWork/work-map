using DotNetEnv;

namespace Auth.GRPC.Extensions
{
    public static class WebApplicationBuilderExtensions
    {
        public static void SetConfiguration(this WebApplicationBuilder builder)
        {
            builder.Configuration.AddJsonFile("appsettings.json", optional: false, reloadOnChange: true);

            var envFilePath = Path.Combine(Directory.GetCurrentDirectory(), ".env");
            if (File.Exists(envFilePath))
            {
                Env.Load(envFilePath);
                builder.Configuration.AddEnvironmentVariables();
            }
        }
    }
}