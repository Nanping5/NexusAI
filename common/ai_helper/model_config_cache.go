package aihelper

import (
	"context"
	"encoding/json"
	"time"

	myredis "NexusAi/common/redis"
	"NexusAi/dao"
	"NexusAi/model"
)

// ModelConfigCacheTTL 模型配置缓存过期时间
const ModelConfigCacheTTL = 5 * time.Minute

// modelConfigCache 带APIKey的缓存结构（用于内部缓存）
type modelConfigCache struct {
	model.AIModelConfig
	APIKey string `json:"api_key"` // 单独存储 APIKey
}

// GetConfigByConfigID 根据 ConfigID 获取模型配置（纯 Redis 缓存）
func GetConfigByConfigID(ctx context.Context, configID string) (*model.AIModelConfig, error) {
	redisKey := myredis.ModelKey(configID)

	// 先从 Redis 缓存获取
	data, err := myredis.Rdb.Get(ctx, redisKey).Result()
	if err == nil {
		var cache modelConfigCache
		if err := json.Unmarshal([]byte(data), &cache); err == nil {
			config := cache.AIModelConfig
			config.APIKey = cache.APIKey // 恢复 APIKey
			return &config, nil
		}
	}

	// 从数据库加载
	config, err := dao.AIModelConfigDAO.GetByConfigID(ctx, configID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	// 写入 Redis 缓存（包含 APIKey）
	cache := modelConfigCache{
		AIModelConfig: *config,
		APIKey:        config.APIKey,
	}
	if data, err := json.Marshal(cache); err == nil {
		myredis.Rdb.Set(ctx, redisKey, string(data), ModelConfigCacheTTL)
	}

	return config, nil
}

// GetDefaultConfig 获取默认模型配置
func GetDefaultConfig(ctx context.Context) (*model.AIModelConfig, error) {
	// 默认配置使用特殊 key
	redisKey := myredis.KeyPrefix + myredis.Separator + "model:default"

	// 先从 Redis 缓存获取
	data, err := myredis.Rdb.Get(ctx, redisKey).Result()
	if err == nil {
		var cache modelConfigCache
		if err := json.Unmarshal([]byte(data), &cache); err == nil {
			config := cache.AIModelConfig
			config.APIKey = cache.APIKey
			return &config, nil
		}
	}

	// 从数据库加载
	config, err := dao.AIModelConfigDAO.GetDefault(ctx)
	if err != nil {
		return nil, err
	}
	if config != nil {
		// 写入 Redis 缓存（包含 APIKey）
		cache := modelConfigCache{
			AIModelConfig: *config,
			APIKey:        config.APIKey,
		}
		if data, err := json.Marshal(cache); err == nil {
			myredis.Rdb.Set(ctx, redisKey, string(data), ModelConfigCacheTTL)
		}
	}
	return config, nil
}

// GetAllEnabledConfigs 获取所有启用的模型配置
func GetAllEnabledConfigs(ctx context.Context) ([]model.AIModelConfig, error) {
	// 列表使用特殊 key
	redisKey := myredis.KeyPrefix + myredis.Separator + "model:enabled_list"

	// 先从 Redis 缓存获取
	data, err := myredis.Rdb.Get(ctx, redisKey).Result()
	if err == nil {
		var caches []modelConfigCache
		if err := json.Unmarshal([]byte(data), &caches); err == nil {
			configs := make([]model.AIModelConfig, 0, len(caches))
			for _, c := range caches {
				config := c.AIModelConfig
				config.APIKey = c.APIKey
				configs = append(configs, config)
			}
			return configs, nil
		}
	}

	// 从数据库加载
	configs, err := dao.AIModelConfigDAO.GetEnabled(ctx)
	if err != nil {
		return nil, err
	}

	// 写入 Redis 缓存（包含 APIKey）
	caches := make([]modelConfigCache, 0, len(configs))
	for _, c := range configs {
		caches = append(caches, modelConfigCache{
			AIModelConfig: c,
			APIKey:        c.APIKey,
		})
	}
	if data, err := json.Marshal(caches); err == nil {
		myredis.Rdb.Set(ctx, redisKey, string(data), ModelConfigCacheTTL)
	}

	return configs, nil
}

// InvalidateConfigCache 使指定配置缓存失效
func InvalidateConfigCache(configID string) {
	ctx := context.Background()
	myredis.Rdb.Del(ctx, myredis.ModelKey(configID))
	// 同时清除默认配置和列表缓存
	myredis.Rdb.Del(ctx, myredis.KeyPrefix+myredis.Separator+"model:default")
	myredis.Rdb.Del(ctx, myredis.KeyPrefix+myredis.Separator+"model:enabled_list")
}

// InvalidateAllConfigCache 使所有模型配置缓存失效
func InvalidateAllConfigCache() {
	ctx := context.Background()
	// 使用 SCAN 删除所有模型配置相关的 key
	iter := myredis.Rdb.Scan(ctx, 0, myredis.KeyPrefix+myredis.Separator+"model:*", 0).Iterator()
	for iter.Next(ctx) {
		myredis.Rdb.Del(ctx, iter.Val())
	}
}
