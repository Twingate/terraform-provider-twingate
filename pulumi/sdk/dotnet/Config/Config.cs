// *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Immutable;

namespace TwingateLabs.Twingate
{
    public static class Config
    {
        [System.Diagnostics.CodeAnalysis.SuppressMessage("Microsoft.Design", "IDE1006", Justification = 
        "Double underscore prefix used to avoid conflicts with variable names.")]
        private sealed class __Value<T>
        {
            private readonly Func<T> _getter;
            private T _value = default!;
            private bool _set;

            public __Value(Func<T> getter)
            {
                _getter = getter;
            }

            public T Get() => _set ? _value : _getter();

            public void Set(T value)
            {
                _value = value;
                _set = true;
            }
        }

        private static readonly global::Pulumi.Config __config = new global::Pulumi.Config("twingate");

        private static readonly __Value<string?> _apiToken = new __Value<string?>(() => __config.Get("apiToken"));
        /// <summary>
        /// The access key for API operations. You can retrieve this from the Twingate Admin Console
        /// ([documentation](https://docs.twingate.com/docs/api-overview)). Alternatively, this can be specified using the
        /// TWINGATE_API_TOKEN environment variable.
        /// </summary>
        public static string? ApiToken
        {
            get => _apiToken.Get();
            set => _apiToken.Set(value);
        }

        private static readonly __Value<int?> _httpMaxRetry = new __Value<int?>(() => __config.GetInt32("httpMaxRetry") ?? 5);
        /// <summary>
        /// Specifies a retry limit for the http requests made. This default value is 10. Alternatively, this can be specified using
        /// the TWINGATE_HTTP_MAX_RETRY environment variable
        /// </summary>
        public static int? HttpMaxRetry
        {
            get => _httpMaxRetry.Get();
            set => _httpMaxRetry.Set(value);
        }

        private static readonly __Value<int?> _httpTimeout = new __Value<int?>(() => __config.GetInt32("httpTimeout") ?? 10);
        /// <summary>
        /// Specifies a time limit in seconds for the http requests made. The default value is 10 seconds. Alternatively, this can
        /// be specified using the TWINGATE_HTTP_TIMEOUT environment variable
        /// </summary>
        public static int? HttpTimeout
        {
            get => _httpTimeout.Get();
            set => _httpTimeout.Set(value);
        }

        private static readonly __Value<string?> _network = new __Value<string?>(() => __config.Get("network"));
        /// <summary>
        /// Your Twingate network ID for API operations. You can find it in the Admin Console URL, for example:
        /// `autoco.twingate.com`, where `autoco` is your network ID Alternatively, this can be specified using the TWINGATE_NETWORK
        /// environment variable.
        /// </summary>
        public static string? Network
        {
            get => _network.Get();
            set => _network.Set(value);
        }

        private static readonly __Value<string?> _url = new __Value<string?>(() => __config.Get("url"));
        /// <summary>
        /// The default is 'twingate.com' This is optional and shouldn't be changed under normal circumstances.
        /// </summary>
        public static string? Url
        {
            get => _url.Get();
            set => _url.Set(value);
        }

    }
}
