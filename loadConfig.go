package main

import (
	_ "embed"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/orandin/lumberjackrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"webhookGo/bilibiliInfo"
	"webhookGo/messageSender"
	"webhookGo/webhookHandler"
)

type initStruct struct {
	listenAddress string
	receivers     []Receiver
}

type Receiver struct {
	Type   string                          `json:"type"`
	Enable bool                            `yaml:"enable"`
	Path   string                          `yaml:"path"`
	Events map[string]webhookHandler.Event `yaml:"events"`
}

type ConfigLoader struct {
	Debug           bool                            `yaml:"debug"`
	ListenAddress   string                          `yaml:"address"`
	ContactBilibili bool                            `yaml:"contact_bilibili"`
	Barks           []messageSender.BarkServer      `yaml:"Bark"`
	WXWorkApps      []messageSender.WXWorkAppTarget `yaml:"WXWorkApp"`
	Receivers       []Receiver                      `yaml:"Receivers"`
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
		log.AddHook(NewRotateHook())
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	var enabledReceivers = 0
	for i := 0; i < len(configuration.Receivers); i++ {
		if configuration.Receivers[i].Enable {
			enabledReceivers++
		}
	}
	if enabledReceivers == 0 {
		log.Fatal("没有启用任何接收器，程序将无法接收任何请求。")
	}

	var senderCount int
	if len(configuration.Barks) > 0 {
		for i := 0; i < len(configuration.Barks); i++ {
			if configuration.Barks[i].BarkSecrets == "" {
				continue
			}
			if configuration.Barks[i].ServerUrl == "" {
				configuration.Barks[i].ServerUrl = "https://api.day.app/"
			}
			configuration.Barks[i].RegisterServer()
			senderCount++
		}
	}
	if len(configuration.WXWorkApps) > 0 {
		for i := 0; i < len(configuration.WXWorkApps); i++ {
			if configuration.WXWorkApps[i].CorpId == "" || configuration.WXWorkApps[i].AppSecret == "" || configuration.WXWorkApps[i].AgentID == "" {
				continue
			}
			if configuration.WXWorkApps[i].ToUser == "" {
				configuration.WXWorkApps[i].ToUser = "@all"
			}
			configuration.WXWorkApps[i].RegisterServer()
			senderCount++
		}
	}
	if senderCount == 0 {
		log.Warn("没有有效的推送目标")
		for i := 0; i < len(configuration.Receivers); i++ {
			for k, event := range configuration.Receivers[i].Events {
				event.Notify = false
				configuration.Receivers[i].Events[k] = event
			}
		}
	}

	if configuration.ListenAddress == "" {
		log.Warn("未指定监听地址，将使用默认地址 127.0.0.1:14000")
		configuration.ListenAddress = "127.0.0.1:14000"
	}
	if !configuration.ContactBilibili {
		log.Warn("不允许访问Bilibili服务器，将无法获取直播间封禁状态和主播头像。")
		bilibiliInfo.ContactBilibili = false
		/*err = os.Setenv("ContactBilibili", "false")
		if err != nil {
			log.Fatal("设置环境变量失败!", err)
		}*/
	}

	config := initStruct{
		listenAddress: configuration.ListenAddress,
	}
	for i := 0; i < len(configuration.Receivers); i++ {
		if configuration.Receivers[i].Enable && configuration.Receivers[i].Type != "" {
			config.receivers = append(config.receivers, configuration.Receivers[i])
		}
	}
	for i := 0; i < len(config.receivers); i++ {
		switch config.receivers[i].Type {
		case "BililiveRecoder":
			webhookHandler.UpdateBililiveRecorderSettings(config.receivers[i].Path, config.receivers[i].Events)
		case "Blrec":
			webhookHandler.UpdateBlrecSettings(config.receivers[i].Path, config.receivers[i].Events)
		case "DDTV3":
			webhookHandler.UpdateDDTV3Settings(config.receivers[i].Path, config.receivers[i].Events)
		case "DDTV5":
			webhookHandler.UpdateDDTV5Settings(config.receivers[i].Path, config.receivers[i].Events)
		}
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

func NewRotateHook() log.Hook {
	hook, _ := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			// 通用日志配置
			Filename:   "output.log",
			MaxSize:    100,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
			LocalTime:  false,
		},
		log.DebugLevel,
		&log.TextFormatter{
			DisableColors: true,
			ForceColors:   false,
			FullTimestamp: true,
		},
		&lumberjackrus.LogFileOpts{
			// 针对不同日志级别的配置
			log.TraceLevel: &lumberjackrus.LogFile{
				Filename:   "trace.log",
				MaxSize:    100,
				MaxBackups: 1,
				MaxAge:     1,
				Compress:   false,
				LocalTime:  false,
			},
			log.DebugLevel: &lumberjackrus.LogFile{
				Filename:   "debug.log",
				MaxSize:    100,
				MaxBackups: 1,
				MaxAge:     2,
				Compress:   false,
				LocalTime:  false,
			},
			log.InfoLevel: &lumberjackrus.LogFile{
				Filename:   "info.log",
				MaxSize:    100,
				MaxBackups: 1,
				MaxAge:     3,
				Compress:   false,
				LocalTime:  false,
			},
			log.ErrorLevel: &lumberjackrus.LogFile{
				Filename:   "error.log",
				MaxSize:    10,
				MaxBackups: 1,
				MaxAge:     10,
				Compress:   false,
				LocalTime:  false,
			},
			log.FatalLevel: &lumberjackrus.LogFile{
				Filename:   "fatal.log",
				MaxSize:    10,
				MaxBackups: 1,
				MaxAge:     10,
				Compress:   false,
				LocalTime:  false,
			},
		},
	)
	return hook
}
