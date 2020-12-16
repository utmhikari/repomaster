package cfg

import (
	"github.com/utmhikari/repomaster/internal/model"
	"github.com/utmhikari/repomaster/pkg/util"
)

var GlobalCfg *model.Config

// NewConfigFromFile offers api to get config from file
func NewConfigFromFile(cfgPath string) (model.Config, error){
	cfg := model.Config{}
	err := util.ReadJsonFile(cfgPath, &cfg)
	if err != nil{
		return cfg, err
	}
	return cfg, nil
}

// InitGlobalConfig Initialize global config
func InitGlobalConfig(cfgPath string) error{
	cfg, err := NewConfigFromFile(cfgPath)
	if err != nil{
		return err
	}
	err = cfg.Check()
	if err != nil{
		return err
	}
	GlobalCfg = &cfg
	return nil
}
