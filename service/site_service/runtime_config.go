package site_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"myblogx/conf"
	"myblogx/global"
	"myblogx/models"
	"myblogx/utils/envyaml"

	"gorm.io/gorm"
)

const runtimeConfigName = "default"

type RuntimeConfig struct {
	Site conf.Site `json:"site"`
	AI   conf.AI   `json:"ai"`
}

var runtimeConfigCache struct {
	mu   sync.RWMutex
	data *RuntimeConfig
}

func defaultRuntimeConfig() *RuntimeConfig {
	cfg := &RuntimeConfig{}
	if global.Config != nil {
		cfg.Site = global.Config.Site
		cfg.AI = global.Config.AI
	}
	return cfg
}

func loadRuntimeDefaultConfigFromFile() (*RuntimeConfig, error) {
	if global.Flags == nil || global.Flags.File == "" {
		return nil, errors.New("运行时站点默认配置路径未设置")
	}
	defaultFile := filepath.Join(filepath.Dir(global.Flags.File), "site_default_settings.yaml")
	byteData, err := os.ReadFile(defaultFile)
	if err != nil {
		return nil, err
	}

	var raw conf.RuntimeSiteDefault
	if err = envyaml.Unmarshal(byteData, &raw); err != nil {
		return nil, fmt.Errorf("解析运行时站点默认配置失败: %w", err)
	}
	return &RuntimeConfig{
		Site: raw.Site,
		AI:   raw.AI,
	}, nil
}

func initialRuntimeConfig() *RuntimeConfig {
	cfg, err := loadRuntimeDefaultConfigFromFile()
	if err == nil {
		return cfg
	}
	if global.Logger != nil {
		global.Logger.Warnf("读取 site_default_settings.yaml 失败，回退到进程内默认值: %v", err)
	}
	return defaultRuntimeConfig()
}

func cloneRuntimeConfig(cfg *RuntimeConfig) *RuntimeConfig {
	if cfg == nil {
		return &RuntimeConfig{}
	}
	cp := *cfg
	return &cp
}

func setRuntimeConfigCache(cfg *RuntimeConfig) {
	runtimeConfigCache.mu.Lock()
	defer runtimeConfigCache.mu.Unlock()
	runtimeConfigCache.data = cloneRuntimeConfig(cfg)
}

func GetRuntimeConfig() RuntimeConfig {
	runtimeConfigCache.mu.RLock()
	if runtimeConfigCache.data != nil {
		cfg := *runtimeConfigCache.data
		runtimeConfigCache.mu.RUnlock()
		return cfg
	}
	runtimeConfigCache.mu.RUnlock()

	cfg := defaultRuntimeConfig()
	return *cfg
}

func GetRuntimeSite() conf.Site {
	return GetRuntimeConfig().Site
}

func GetRuntimeAI() conf.AI {
	return GetRuntimeConfig().AI
}

func applyRuntimeConfig(cfg *RuntimeConfig) {
	if global.Config != nil && cfg != nil {
		global.Config.Site = cfg.Site
		global.Config.AI = cfg.AI
	}
	setRuntimeConfigCache(cfg)
}

func marshalRuntimeConfig(cfg *RuntimeConfig) (siteJSON string, aiJSON string, err error) {
	siteData, err := json.Marshal(cfg.Site)
	if err != nil {
		return "", "", fmt.Errorf("序列化站点配置失败: %w", err)
	}
	aiData, err := json.Marshal(cfg.AI)
	if err != nil {
		return "", "", fmt.Errorf("序列化 AI 配置失败: %w", err)
	}
	return string(siteData), string(aiData), nil
}

func decodeRuntimeConfig(model *models.RuntimeSiteConfigModel) (*RuntimeConfig, error) {
	cfg := defaultRuntimeConfig()
	if model == nil {
		return cfg, nil
	}
	if model.SiteJSON != "" {
		if err := json.Unmarshal([]byte(model.SiteJSON), &cfg.Site); err != nil {
			return nil, fmt.Errorf("解析站点配置失败: %w", err)
		}
	}
	if model.AIJSON != "" {
		if err := json.Unmarshal([]byte(model.AIJSON), &cfg.AI); err != nil {
			return nil, fmt.Errorf("解析 AI 配置失败: %w", err)
		}
	}
	return cfg, nil
}

func ensureRuntimeConfigModel(tx *gorm.DB) (*models.RuntimeSiteConfigModel, *RuntimeConfig, error) {
	var model models.RuntimeSiteConfigModel
	err := tx.Where("name = ?", runtimeConfigName).Take(&model).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		cfg := initialRuntimeConfig()
		siteJSON, aiJSON, marshalErr := marshalRuntimeConfig(cfg)
		if marshalErr != nil {
			return nil, nil, marshalErr
		}
		model = models.RuntimeSiteConfigModel{
			Name:     runtimeConfigName,
			SiteJSON: siteJSON,
			AIJSON:   aiJSON,
		}
		if createErr := tx.Create(&model).Error; createErr != nil {
			return nil, nil, createErr
		}
		return &model, cfg, nil
	case err != nil:
		return nil, nil, err
	default:
		cfg, decodeErr := decodeRuntimeConfig(&model)
		if decodeErr != nil {
			return nil, nil, decodeErr
		}
		return &model, cfg, nil
	}
}

// InitRuntimeConfig 会在启动时把运行时站点配置从数据库加载到内存。
// 当数据库还没有初始化记录时，会使用 site_default_settings.yaml 中的默认值灌入数据库。
func InitRuntimeConfig() error {
	if global.DB == nil {
		return errors.New("数据库未初始化")
	}
	var loaded *RuntimeConfig
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		_, cfg, err := ensureRuntimeConfigModel(tx)
		if err != nil {
			return err
		}
		loaded = cfg
		return nil
	}); err != nil {
		return err
	}
	applyRuntimeConfig(loaded)
	return nil
}

func UpdateRuntimeSite(site conf.Site) error {
	if global.DB == nil {
		return errors.New("数据库未初始化")
	}
	var updated *RuntimeConfig
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		model, cfg, err := ensureRuntimeConfigModel(tx)
		if err != nil {
			return err
		}
		cfg.Site = site
		siteJSON, aiJSON, marshalErr := marshalRuntimeConfig(cfg)
		if marshalErr != nil {
			return marshalErr
		}
		if err = tx.Model(model).Updates(map[string]any{
			"site_json": siteJSON,
			"ai_json":   aiJSON,
		}).Error; err != nil {
			return err
		}
		updated = cfg
		return nil
	}); err != nil {
		return err
	}
	applyRuntimeConfig(updated)
	return nil
}

func UpdateRuntimeAI(ai conf.AI) error {
	if global.DB == nil {
		return errors.New("数据库未初始化")
	}
	var updated *RuntimeConfig
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		model, cfg, err := ensureRuntimeConfigModel(tx)
		if err != nil {
			return err
		}
		cfg.AI = ai
		siteJSON, aiJSON, marshalErr := marshalRuntimeConfig(cfg)
		if marshalErr != nil {
			return marshalErr
		}
		if err = tx.Model(model).Updates(map[string]any{
			"site_json": siteJSON,
			"ai_json":   aiJSON,
		}).Error; err != nil {
			return err
		}
		updated = cfg
		return nil
	}); err != nil {
		return err
	}
	applyRuntimeConfig(updated)
	return nil
}
