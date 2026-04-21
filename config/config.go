package config

import (
	"github.com/spf13/viper"
)

type MainConfig struct {
	AppName string `mapstructure:"app_name" json:"app_name"`
	Host    string `mapstructure:"host" json:"host"`
	Port    int    `mapstructure:"port" json:"port"`
}

type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
	DbName   string `mapstructure:"db_name" json:"db_name"`
}
type RedisConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Password string `mapstructure:"password" json:"password"`
	DB       int    `mapstructure:"db" json:"db"`
}
type AuthCodeConfig struct {
	AccessKeyId     string `mapstructure:"access_key_id" json:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" json:"access_key_secret"`
	SignName        string `mapstructure:"sign_name" json:"sign_name"`
	TemplateCode    string `mapstructure:"template_code" json:"template_code"`
}
type LogConfig struct {
	LogPath string `mapstructure:"log_path" json:"log_path"`
}

type JwtConfig struct {
	SecretKey string `mapstructure:"secret_key" json:"secret_key"`
	Issuer    string `mapstructure:"issuer" json:"issuer"`
	Subject   string `mapstructure:"subject" json:"subject"`
}
type RabbitMQConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Username string `mapstructure:"username" json:"username"`
	Password string `mapstructure:"password" json:"password"`
	Vhost    string `mapstructure:"vhost" json:"vhost"`
}
type StaticSrcConfig struct {
	StaticAvatarPath string `mapstructure:"static_avatar_path" json:"static_avatar_path"`
	StaticFilePath   string `mapstructure:"static_file_path" json:"static_file_path"`
}
type Smtp struct {
	EmailAddr  string `mapstructure:"email_addr" json:"email_addr"`
	SmtpKey    string `mapstructure:"smtp_key" json:"smtp_key"`
	SmtpServer string `mapstructure:"smtp_server" json:"smtp_server"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins" json:"allowed_origins"`
}

type RagConfig struct {
	RagDimension      int    `mapstructure:"rag_dimension" json:"rag_dimension"`
	RagChatModelName  string `mapstructure:"rag_chat_model_name" json:"rag_chat_model_name"`
	RagDocDir         string `mapstructure:"rag_doc_dir" json:"rag_doc_dir"`
	RagBaseURL        string `mapstructure:"rag_base_url" json:"rag_base_url"`
	RagEmbeddingModel string `mapstructure:"rag_embedding_model" json:"rag_embedding_model"`
	RagChunkSize      int    `mapstructure:"rag_chunk_size" json:"rag_chunk_size"`       // 每个 chunk 的目标 token 数
	RagChunkOverlap   int    `mapstructure:"rag_chunk_overlap" json:"rag_chunk_overlap"`  // 相邻 chunk 之间的重叠 token 数
}

type VoiceServiceConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id" json:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" json:"access_key_secret"`
	AppKey          string `mapstructure:"app_key" json:"app_key"`
}

// AIConfig AI 相关配置
type AIConfig struct {
	MaxContextMessages     int    `mapstructure:"max_context_messages" json:"max_context_messages"`         // 最大上下文消息数（轮次）
	MaxContextTokens      int    `mapstructure:"max_context_tokens" json:"max_context_tokens"`               // 最大上下文 Token 数
	ContextStrategy       string `mapstructure:"context_strategy" json:"context_strategy"`                   // 上下文策略：sliding_window / summary
	RequestQueueSize      int    `mapstructure:"request_queue_size" json:"request_queue_size"`             // 请求队列大小
	DefaultMaxIterations  int    `mapstructure:"default_max_iterations" json:"default_max_iterations"`      // Agent 默认最大迭代次数
	HelperCleanupInterval int    `mapstructure:"helper_cleanup_interval" json:"helper_cleanup_interval"`   // Helper 清理间隔（分钟）
	HelperExpireDuration  int    `mapstructure:"helper_expire_duration" json:"helper_expire_duration"`     // Helper 过期时间（分钟）
}

type QdrantConfig struct {
	Host       string `mapstructure:"host" json:"host"`
	Port       int    `mapstructure:"port" json:"port"`
	APIKey     string `mapstructure:"api_key" json:"api_key"`
	Collection string `mapstructure:"collection" json:"collection"`
}

type Config struct {
	MainConfig         MainConfig         `mapstructure:"main_config" json:"main_config"`
	MysqlConfig        MysqlConfig        `mapstructure:"mysql_config" json:"mysql_config"`
	RedisConfig        RedisConfig        `mapstructure:"redis_config" json:"redis_config"`
	AuthCodeConfig     AuthCodeConfig     `mapstructure:"auth_code_config" json:"auth_code_config"`
	LogConfig          LogConfig          `mapstructure:"log_config" json:"log_config"`
	RabbitMQConfig     RabbitMQConfig     `mapstructure:"rabbitmq_config" json:"rabbitmq_config"`
	StaticSrcConfig    StaticSrcConfig    `mapstructure:"static_src_config" json:"static_src_config"`
	Smtp               Smtp               `mapstructure:"smtp" json:"smtp"`
	JwtConfig          JwtConfig          `mapstructure:"jwt_config" json:"jwt_config"`
	CORSConfig         CORSConfig         `mapstructure:"cors_config" json:"cors_config"`
	RagConfig          RagConfig          `mapstructure:"rag_config" json:"rag_config"`
	QdrantConfig       QdrantConfig       `mapstructure:"qdrant_config" json:"qdrant_config"`
	VoiceServiceConfig VoiceServiceConfig `mapstructure:"voice_service_config" json:"voice_service_config"`
	AIConfig           AIConfig           `mapstructure:"ai_config" json:"ai_config"`
}

var config *Config

func LoadConfig() error {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")

	// 这里增加多种路径查找，以支持单元测试在不同目录运行
	v.AddConfigPath("./config")        // root 运行
	v.AddConfigPath("../../config")    // internal/service/gorms/ 运行
	v.AddConfigPath("../../../config") // 更深层 运行
	v.AddConfigPath(".")

	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	cfg := new(Config)
	err = v.Unmarshal(cfg)
	if err != nil {
		return err
	}
	config = cfg
	return nil
}

func GetConfig() *Config {
	if config == nil {
		panic("配置未初始化，请先调用 LoadConfig()")
	}
	return config
}
