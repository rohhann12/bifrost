// Configuration types that match the Go backend structures

import { KNOWN_PROVIDERS } from "@/lib/constants/logs";

// Known provider names - all supported standard providers
export type KnownProvider = (typeof KNOWN_PROVIDERS)[number];

// Branded type for custom provider names to prevent collision with known providers
export type CustomProviderName = string & { readonly __brand: "CustomProviderName" };

// ModelProvider union - either known providers or branded custom providers
export type ModelProvider = KnownProvider | CustomProviderName;

// Helper function to check if a provider name is a known provider
export const isKnownProvider = (provider: string): provider is KnownProvider => {
	return KNOWN_PROVIDERS.includes(provider.toLowerCase() as KnownProvider);
};

// AzureKeyConfig matching Go's schemas.AzureKeyConfig
export interface AzureKeyConfig {
	endpoint: string;
	deployments?: Record<string, string> | string; // Allow string during editing
	api_version?: string;
}

// VertexKeyConfig matching Go's schemas.VertexKeyConfig
export interface VertexKeyConfig {
	project_id: string;
	region: string;
	auth_credentials: string; // Always string - JSON string or env var
}

// BedrockKeyConfig matching Go's schemas.BedrockKeyConfig
export interface BedrockKeyConfig {
	access_key: string;
	secret_key: string;
	session_token?: string;
	region?: string;
	arn?: string;
	deployments?: Record<string, string> | string; // Allow string during editing
}

// Key structure matching Go's schemas.Key
export interface Key {
	id: string;
	value: string;
	models: string[];
	weight: number;
	azure_key_config?: AzureKeyConfig;
	vertex_key_config?: VertexKeyConfig;
	bedrock_key_config?: BedrockKeyConfig;
}

// NetworkConfig matching Go's schemas.NetworkConfig
export interface NetworkConfig {
	base_url?: string;
	extra_headers?: Record<string, string>;
	default_request_timeout_in_seconds: number;
	max_retries: number;
	retry_backoff_initial: number; // Duration in milliseconds
	retry_backoff_max: number; // Duration in milliseconds
}

// ConcurrencyAndBufferSize matching Go's schemas.ConcurrencyAndBufferSize
export interface ConcurrencyAndBufferSize {
	concurrency: number;
	buffer_size: number;
}

// Proxy types matching Go's schemas.ProxyType
export type ProxyType = "none" | "http" | "socks5" | "environment";

// ProxyConfig matching Go's schemas.ProxyConfig
export interface ProxyConfig {
	type: ProxyType;
	url?: string;
	username?: string;
	password?: string;
}

// CustomProviderConfig matching Go's schemas.CustomProviderConfig
export interface CustomProviderConfig {
	base_provider_type: KnownProvider;
	allowed_requests?: AllowedRequests;
}

// AllowedRequests matching Go's schemas.AllowedRequests
export interface AllowedRequests {
	text_completion: boolean;
	chat_completion: boolean;
	chat_completion_stream: boolean;
	embedding: boolean;
	speech: boolean;
	speech_stream: boolean;
	transcription: boolean;
	transcription_stream: boolean;
}

// ProviderConfig matching Go's lib.ProviderConfig
export interface ProviderConfig {
	keys: Key[];
	network_config: NetworkConfig;
	concurrency_and_buffer_size: ConcurrencyAndBufferSize;
	proxy_config?: ProxyConfig;
	send_back_raw_response?: boolean;
	custom_provider_config?: CustomProviderConfig;
}

// ProviderResponse matching Go's ProviderResponse
export interface ProviderResponse extends ProviderConfig {
	name: ModelProvider;
}

// ListProvidersResponse matching Go's ListProvidersResponse
export interface ListProvidersResponse {
	providers: ProviderResponse[];
	total: number;
}

// AddProviderRequest matching Go's AddProviderRequest
export interface AddProviderRequest {
	provider: ModelProvider;
	keys: Key[];
	network_config?: NetworkConfig;
	concurrency_and_buffer_size?: ConcurrencyAndBufferSize;
	proxy_config?: ProxyConfig;
	send_back_raw_response?: boolean;
	custom_provider_config?: CustomProviderConfig;
}

// UpdateProviderRequest matching Go's UpdateProviderRequest
export interface UpdateProviderRequest {
	keys: Key[];
	network_config: NetworkConfig;
	concurrency_and_buffer_size: ConcurrencyAndBufferSize;
	proxy_config: ProxyConfig;
	send_back_raw_response?: boolean;
	custom_provider_config?: CustomProviderConfig;
}
// Define DB configs
export interface SqliteConfig {
	type: "sqlite";
	config: {
		path: string;
	};
}

export interface PostgresConfig {
	type: "postgres";
	config: {
		host: string;
		port: number;
		user: string;
		password: string;
		database: string;
	};
}
// Union type
export type DatabaseConfig = SqliteConfig | PostgresConfig;

// BifrostErrorResponse matching Go's schemas.BifrostError
export interface BifrostErrorResponse {
	event_id?: string;
	type?: string;
	is_bifrost_error: boolean;
	status_code?: number;
	error: {
		message: string;
		type?: string;
		code?: string;
		param?: string;
	};
}

// Bifrost Config
export interface BifrostConfig {
	client_config: CoreConfig;
	is_db_connected: boolean;
	is_cache_connected: boolean;
	is_logs_connected: boolean;
}

// Core Bifrost configuration types
export interface CoreConfig {
	drop_excess_requests: boolean;
	initial_pool_size: number;
	prometheus_labels: string[];
	enable_logging: boolean;
	enable_governance: boolean;
	enforce_governance_header: boolean;
	allow_direct_keys: boolean;
	allowed_origins: string[];
	db: DatabaseConfig;
	config_store: DatabaseConfig
}

// Semantic cache configuration types
export interface CacheConfig {
	provider: ModelProvider;
	keys: Key[];
	embedding_model: string;
	ttl_seconds: number;
	threshold: number;
	cache_by_model: boolean;
	cache_by_provider: boolean;
	created_at?: string;
	updated_at?: string;
}

// Utility types for form handling
export interface ProviderFormData {
	provider: ModelProvider | "";
	keys: Array<{ value: string; models: string[]; weight: number }>;
	network_config: {
		baseURL?: string;
		defaultRequestTimeoutInSeconds: number;
		maxRetries: number;
	};
	concurrency_and_buffer_size: {
		concurrency: number;
		bufferSize: number;
	};
	custom_provider_config?: CustomProviderConfig;
}

// Status types
export type ProviderStatus = "active" | "error" | "added" | "updated" | "deleted";
