package main

import (
	_ "embed"
	"flag"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"webhookGo/bilibiliInfo"
	"webhookGo/messageSender"
	"webhookGo/webhookHandler"
)

type initStruct struct {
	listenAddress   string
	bililiveRecoder struct {
		enable bool
		path   string
	}
	blrec struct {
		enable bool
		path   string
	}
	ddtv struct {
		enable bool
		path   string
	}
}

type ConfigLoader struct {
	ListenAddress   string                          `yaml:"address"`
	ContactBilibili bool                            `yaml:"contact_bilibili"`
	Barks           []messageSender.BarkServer      `yaml:"Bark"`
	WXWorkApps      []messageSender.WXWorkAppTarget `yaml:"WXWorkApp"`
	BililiveRecoder struct {
		Enable bool                            `yaml:"enable"`
		Path   string                          `yaml:"path"`
		Events map[string]webhookHandler.Event `yaml:"events"`
	} `yaml:"BililiveRecoder"`
	Blrec struct {
		Enable bool                            `yaml:"enable"`
		Path   string                          `yaml:"path"`
		Events map[string]webhookHandler.Event `yaml:"events"`
	} `yaml:"Blrec"`
	DDTV struct {
		Enable bool                            `yaml:"enable"`
		Path   string                          `yaml:"path"`
		Events map[string]webhookHandler.Event `yaml:"events"`
	} `yaml:"DDTV"`
}

func loadConfig() initStruct {
	var configuration ConfigLoader
	var configFile string
	flag.StringVar(&configFile, "c", "config.yml", "config file")
	flag.Parse()
	file, err := os.ReadFile(configFile)
	if err == nil {
		if len(file) < 5 {
			writeDefaultConfig(configFile)
		}
		err = yaml.Unmarshal(file, &configuration)
		if err != nil {
			log.Fatal("配置文件不合法!", err)
		}
	} else {
		writeDefaultConfig(configFile)
	}

	if !(configuration.BililiveRecoder.Enable || configuration.Blrec.Enable || configuration.DDTV.Enable) {
		log.Fatal("没有关注的事件")
	}

	var barkCount int
	var weWorkCount int
	if len(configuration.Barks) > 0 {
		for _, bark := range configuration.Barks {
			if bark.BarkSecrets == "" {
				continue
			}
			if bark.ServerUrl == "" {
				bark.ServerUrl = "https://api.day.app/"
			}
			messageSender.RegisterBarkServer(bark)
			barkCount++
		}
	}
	if len(configuration.WXWorkApps) > 0 {
		for _, app := range configuration.WXWorkApps {
			if app.CorpId == "" || app.AppSecret == "" || app.AgentID == "" {
				continue
			}
			if app.ToUser == "" {
				app.ToUser = "@all"
			}
			messageSender.RegisterWXWorkApp(app)
			weWorkCount++
		}
	}
	if barkCount == 0 && weWorkCount == 0 {
		log.Warn("没有有效的推送目标")
		for k, v := range configuration.BililiveRecoder.Events {
			v.Notify = false
			configuration.BililiveRecoder.Events[k] = v
		}
		for k, v := range configuration.Blrec.Events {
			v.Notify = false
			configuration.BililiveRecoder.Events[k] = v
		}
		for k, v := range configuration.DDTV.Events {
			v.Notify = false
			configuration.BililiveRecoder.Events[k] = v
		}
	}

	if configuration.ListenAddress == "" {
		log.Warn("未指定监听地址，将使用默认地址 127.0.0.1:14000")
		configuration.ListenAddress = "127.0.0.1:14000"
	}
	if !configuration.ContactBilibili {
		log.Warn("不允许访问Bilibili服务器，将无法获取直播间封禁状态和主播头像。")
	}
	bilibiliInfo.ContactBilibili = configuration.ContactBilibili
	config := initStruct{
		listenAddress: configuration.ListenAddress,
		bililiveRecoder: struct {
			enable bool
			path   string
		}{configuration.BililiveRecoder.Enable, configuration.BililiveRecoder.Path},
		blrec: struct {
			enable bool
			path   string
		}{configuration.Blrec.Enable, configuration.Blrec.Path},
		ddtv: struct {
			enable bool
			path   string
		}{configuration.DDTV.Enable, configuration.DDTV.Path},
	}
	if configuration.BililiveRecoder.Enable {
		webhookHandler.UpdateBililiveRecoderSettings(configuration.BililiveRecoder.Events)
	}
	if configuration.Blrec.Enable {
		webhookHandler.UpdateBlrecSettings(configuration.Blrec.Events)
	}
	if configuration.DDTV.Enable {
		webhookHandler.UpdateDDTVSettings(configuration.DDTV.Events)
	}
	return config
}

//go:embed defaultConfig.yml
var defaultConfig []byte

// writeDefaultConfig 没有读取到配置文件时，新建一个。
func writeDefaultConfig(secretFile string) {
	err := os.WriteFile(secretFile, defaultConfig, 0o644)
	if err != nil {
		log.Fatal("写入默认secrets文件失败!", err)
	} else {
		log.Info("写入默认secrets文件成功，请修改配置文件后重启程序。")
	}
	os.Exit(0)
}
