using Auth.Application.AppUsers;
using Auth.GRPC.Controllers;
using Auth.GRPC.Extensions;
using Auth.Infrastructure;
using Auth.Infrastructure.Persistance.Extensions;

var builder = WebApplication.CreateBuilder(args);

builder.SetConfiguration();

builder.Services.AddGrpc();

builder.Services.AddInfrastructeServices(builder.Configuration);

builder.Services.AddMediatR(cfg => cfg.RegisterServicesFromAssemblies(typeof(Register.Handler).Assembly));

builder.Services.AddAuthorization();

var app = builder.Build();

// Configure the HTTP request pipeline.

app.UseMigration();

app.UseAuthentication();
app.UseAuthorization();

app.MapGrpcService<AuthService>();

app.Run();
