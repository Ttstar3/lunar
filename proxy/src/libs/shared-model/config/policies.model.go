package config

import "github.com/go-playground/validator/v10"

type PoliciesConfig struct {
	Global    Global                `yaml:"global"`
	Endpoints []EndpointConfig      `yaml:"endpoints"`
	Accounts  map[AccountID]Account `yaml:"accounts"  validate:"dive"`
	Exporters Exporters             `yaml:"exporters"`
}

type (
	AccountID    string
	QuotaGroupID int
)

type Account struct {
	Tokens []Token `yaml:"tokens"`
}

type Token struct {
	Header *Header `yaml:"header" validate:"required"`
}

type Header struct {
	Name  string `yaml:"name"  validate:"required"`
	Value string `yaml:"value" validate:"required"`
}

type Exporters struct {
	File       *FileExporterConfig    `yaml:"file"`
	S3         *S3ExporterConfig      `yaml:"s3"`
	S3Minio    *S3MinioExporterConfig `yaml:"s3_minio"`
	Prometheus *PrometheusConfig      `yaml:"prometheus"`
}

type Global struct {
	Remedies  []Remedy    `yaml:"remedies"  validate:"dive"`
	Diagnosis []Diagnosis `yaml:"diagnosis" validate:"dive"`
}

type EndpointConfig struct {
	URL       string      `yaml:"url"`
	Method    string      `yaml:"method"    validate:"required"`
	Remedies  []Remedy    `yaml:"remedies"  validate:"dive"`
	Diagnosis []Diagnosis `yaml:"diagnosis" validate:"dive"`
}

// Remedy

type Remedy struct {
	Enabled    bool         `yaml:"enabled"`
	Name       string       `yaml:"name"    validate:"required"`
	Config     RemedyConfig `yaml:"config"`
	remedyType RemedyType
}

type RemedyConfig struct {
	Caching                    *CachingConfig                    `yaml:"caching"`
	ResponseBasedThrottling    *ResponseBasedThrottlingConfig    `yaml:"response_based_throttling"`    //nolint:lll
	StrategyBasedThrottling    *StrategyBasedThrottlingConfig    `yaml:"strategy_based_throttling"`    //nolint:lll
	ConcurrencyBasedThrottling *ConcurrencyBasedThrottlingConfig `yaml:"concurrency_based_throttling"` //nolint:lll
	AccountOrchestration       *AccountOrchestrationConfig       `yaml:"account_orchestration"`        //nolint:lll
	FixedResponse              *FixedResponseConfig              `yaml:"fixed_response"`               //nolint:lll
	Retry                      *RetryConfig                      `yaml:"retry"`
}

type RemedyType int

const (
	RemedyUndefined RemedyType = iota
	RemedyCaching
	RemedyResponseBasedThrottling
	RemedyStrategyBasedThrottling
	RemedyConcurrencyBasedThrottling
	RemedyAccountOrchestration
	RemedyFixedResponse
	RemedyRetry
)

type CachingConfig struct {
	RequestKeys string  `yaml:"request_keys"`
	TTLSeconds  float32 `yaml:"ttl_seconds"`
	MaxBytes    int     `yaml:"max_bytes"`
}

type ResponseBasedThrottlingConfig struct {
	QuotaGroup       int            `yaml:"quota_group"`
	RetryAfterHeader string         `yaml:"retry_after_header"`
	RetryAfterType   RetryAfterType `yaml:"retry_after_type"`
	RelevantStatuses []int          `yaml:"relevant_statuses"  validate:"required,dive,min=100,max=599"` //nolint:lll
}

type StrategyBasedThrottlingConfig struct {
	AllowedRequestCount  int                   `yaml:"allowed_request_count"`
	WindowSizeInSeconds  int                   `yaml:"window_size_in_seconds"`
	GroupQuotaAllocation *GroupQuotaAllocation `yaml:"group_quota_allocation"`
	ResponseStatusCode   int                   `yaml:"response_status_code"`
}

type ConcurrencyBasedThrottlingConfig struct {
	MaxConcurrentRequests int `yaml:"max_concurrent_requests"`
	ResponseStatusCode    int `yaml:"response_status_code"    validate:"required,min=100,max=599"` //nolint:lll
}

type AccountOrchestrationConfig struct {
	RoundRobin []AccountID `yaml:"round_robin" validate:"required"`
}

type FixedResponseConfig struct {
	StatusCode int `yaml:"status_code" validate:"required,min=100,max=599"`
}

type RetryConfig struct {
	Attempts               int                   `yaml:"attempts"`
	InitialCooldownSeconds int                   `yaml:"initial_cooldown_seconds"`
	CooldownMultiplier     int                   `yaml:"cooldown_multiplier"`
	Conditions             RetryConfigConditions `yaml:"conditions"`
}

type RetryConfigConditions struct {
	StatusCode []Range[int] `yaml:"status_code" validate:"required"`
}

type Range[T any] struct {
	From T `yaml:"from"`
	To   T `yaml:"to"`
}
type GroupBy struct {
	HeaderName string `yaml:"header_name"`
}
type GroupQuotaAllocation struct {
	GroupBy                     *GroupBy                         `yaml:"group_by"                      validate:"required"` //nolint:lll
	Groups                      []QuotaAllocation                `yaml:"groups"                        validate:"dive"`     //nolint:lll
	Default                     defaultQuotaGroupBehaviorLiteral `yaml:"default"`
	DefaultAllocationPercentage float64                          `yaml:"default_allocation_percentage" validate:"gte=0"` //nolint:lll
}

type (
	defaultQuotaGroupBehaviorLiteral = string
	DefaultQuotaGroupBehavior        int
)

const (
	DefaultQuotaGroupBehaviorUndefined DefaultQuotaGroupBehavior = iota
	DefaultQuotaGroupBehaviorAllow
	DefaultQuotaGroupBehaviorBlock
	DefaultQuotaGroupBehaviorUseDefaultAllocation
)

func (behavior DefaultQuotaGroupBehavior) String() string {
	var res string
	switch behavior {
	case DefaultQuotaGroupBehaviorAllow:
		res = "allow"
	case DefaultQuotaGroupBehaviorBlock:
		res = "block"
	case DefaultQuotaGroupBehaviorUseDefaultAllocation:
		res = "use_default_allocation"
	case DefaultQuotaGroupBehaviorUndefined:
		res = "undefined"
	}

	return res
}

type QuotaAllocation struct {
	GroupHeaderValue     string  `yaml:"group_header_value"`
	AllocationPercentage float64 `yaml:"allocation_percentage" validate:"gte=0"`
}

type RetryAfterType int

const (
	RetryAfterUndefined RetryAfterType = iota
	RetryAfterAbsoluteEpoch
	RetryAfterRelativeSeconds
)

// Diagnosis

type Diagnosis struct {
	Enabled bool            `yaml:"enabled"`
	Name    string          `yaml:"name"    validate:"required"`
	Config  DiagnosisConfig `yaml:"config"  validate:"required"`
	Export  string          `yaml:"export"  validate:"required"`

	exporterType  ExporterType
	diagnosisType DiagnosisType
}

type DiagnosisConfig struct {
	HARExporter      *HARExporterConfig      `yaml:"har_exporter"`
	MetricsCollector *MetricsCollectorConfig `yaml:"metrics_collector"`
	Void             *VoidConfig             `yaml:"void"`
}

type HARExporterConfig struct {
	TransactionMaxSize  int         `yaml:"transaction_max_size"`
	Obfuscate           Obfuscate   `yaml:"obfuscate"`
	RequestHeaderNames  HeaderNames `yaml:"request_header_names"`
	ResponseHeaderNames HeaderNames `yaml:"response_header_names"`
}

type Obfuscate struct {
	Enabled    bool                  `yaml:"enabled"`
	Exclusions ObfuscationExclusions `yaml:"exclusions"`
}

type ObfuscationExclusions struct {
	QueryParams       []string `yaml:"query_params"`
	PathParams        []string `yaml:"path_params"`
	RequestHeaders    []string `yaml:"request_headers"`
	ResponseHeaders   []string `yaml:"response_headers"`
	RequestBodyPaths  []string `yaml:"request_body_paths"`
	ResponseBodyPaths []string `yaml:"response_body_paths"`
}

type HeaderNames struct {
	ContentEncoding string `yaml:"content_encoding"`
}

type MetricsCollectorConfig struct {
	RequestHeaderNames []string `yaml:"request_header_names"`
}

type VoidConfig struct{}

type DiagnosisType int

const (
	DiagnosisUndefined DiagnosisType = iota
	DiagnosisHARExporter
	DiagnosisMetricsCollector
	DiagnosisVoid
)

// Export

type ExporterType int

const (
	ExporterUndefined ExporterType = iota
	ExporterFile
	ExporterS3
	ExporterS3Minio
	ExporterPrometheus
)

type ExporterName = string

const (
	ExporterNameUndefined  ExporterName = "undefined"
	ExporterNameFile       ExporterName = "file"
	ExporterNameS3         ExporterName = "s3"
	ExporterNameS3Minio    ExporterName = "s3_minio"
	ExporterNamePrometheus ExporterName = "prometheus"
)

func (exporterType ExporterType) Name() ExporterName {
	var res string
	switch exporterType {
	case ExporterFile:
		res = ExporterNameFile
	case ExporterS3:
		res = ExporterNameS3
	case ExporterS3Minio:
		res = ExporterNameS3Minio
	case ExporterPrometheus:
		res = ExporterNamePrometheus
	case ExporterUndefined:
		res = ExporterNameUndefined
	}

	return res
}

type FileExporterConfig struct {
	FileDir  string `yaml:"file_dir"  validate:"required"`
	FileName string `yaml:"file_name" validate:"required"`
}

type S3ExporterConfig struct {
	BucketName string `yaml:"bucket_name" validate:"required"`
	Region     string `yaml:"region"      validate:"required"`
}

type S3MinioExporterConfig struct {
	BucketName string `yaml:"bucket_name" validate:"required"`
	URL        string `yaml:"url"         validate:"required"`
}

type PrometheusConfig struct {
	BucketBoundaries []float64 `yaml:"bucket_boundaries"`
}

// use a single instance of Validate, it caches struct info
var Validate *validator.Validate = validator.New()
