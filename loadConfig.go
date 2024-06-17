package main

import (
	_ "embed"
	"flag"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"webhookGo/bilibiliInfo"
	"webhookGo/messageSender"
	"webhookGo/webhookHandler"
)

type options struct {
	enable bool
	path   string
}
type initStruct struct {
	listenAddress   string
	bililiveRecoder options
	blrec           options
	ddtv3           options
}
type ConfigLoader struct {
	Debug           bool                            `yaml:"debug"`
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
	DDTV3 struct {
		Enable bool                            `yaml:"enable"`
		Path   string                          `yaml:"path"`
		Events map[string]webhookHandler.Event `yaml:"events"`
	} `yaml:"DDTV3"`
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
		err = yaml.Unmarshal(defaultConfig, &configuration)
		if err != nil {
			log.Fatal(err)
		}
	}

	if configuration.Debug {
		log.SetLevel(log.DebugLevel)
		log.Warnf("已开启Debug模式.")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	if !(configuration.BililiveRecoder.Enable || configuration.Blrec.Enable || configuration.DDTV3.Enable) {
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
			bark.RegisterBarkServer()
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
			app.RegisterWXWorkApp()
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
		for k, v := range configuration.DDTV3.Events {
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
		listenAddress:   configuration.ListenAddress,
		bililiveRecoder: options{enable: configuration.BililiveRecoder.Enable, path: configuration.BililiveRecoder.Path},
		blrec:           options{enable: configuration.Blrec.Enable, path: configuration.Blrec.Path},
		ddtv3:           options{enable: configuration.DDTV3.Enable, path: configuration.DDTV3.Path},
	}
	if configuration.BililiveRecoder.Enable {
		webhookHandler.UpdateBililiveRecoderSettings(configuration.BililiveRecoder.Events)
	}
	if configuration.Blrec.Enable {
		webhookHandler.UpdateBlrecSettings(configuration.Blrec.Events)
	}
	if configuration.DDTV3.Enable {
		webhookHandler.UpdateDDTVSettings(configuration.DDTV3.Events)
	}
	return config
}

//go:embed defaultConfig.yml
var defaultConfig []byte

// writeDefaultConfig 没有读取到配置文件时，新建一个。
func writeDefaultConfig(secretFile string) {
	err := os.WriteFile(secretFile, defaultConfig, 0o644)
	if err != nil {
		log.Fatal("写入默认配置文件失败!", err)
	} else {
		log.Warn("写入默认配置文件成功，请修改配置文件后重启程序。")
	}
}
