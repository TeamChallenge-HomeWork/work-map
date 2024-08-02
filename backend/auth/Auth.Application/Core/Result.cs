using Grpc.Core;

namespace Auth.Application.Core
{
    public class Result<T>
    {
        public T? Value { get; set; }
        public RpcException? Error { get; set; }
        public bool IsSuccess => Error == null;
        public static Result<T> Success(T value) => new Result<T> { Value = value, Error = null };
        public static Result<T> Failure(RpcException error) => new Result<T> { Value = default, Error = error };
    }
}
