package config

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"webhookGo/messageSender"
)

type SecretsLoader struct {
	Barks      []messageSender.BarkServer      `yaml:"Bark"`
	WeworkApps []messageSender.WeworkAppTarget `yaml:"WeWorkApp"`
}

// init 初始化，导入包时会自动调用，无需额外调用。
func init() {
	var secrets SecretsLoader
	var SecretFile string
	flag.StringVar(&SecretFile, "s", "secrets.yml", "secret file")
	flag.Parse()
	file, err := os.ReadFile(SecretFile)
	if err == nil {
		err = yaml.Unmarshal(file, secrets)
		if err != nil {
			log.Fatal("配置文件不合法!", err)
		}
	} else {
		writeDefaultSecrets(SecretFile)
	}
	if len(secrets.Barks) > 0 {
		for _, bark := range secrets.Barks {
			messageSender.RegisterBarkServer(bark)
		}
	}
	if len(secrets.WeworkApps) > 0 {
		for i := 0; i < len(secrets.WeworkApps); i++ {
			messageSender.RegisterWeworkAppTarget(secrets.WeworkApps[i])
		}
	}
}

// writeDefaultSecrets 没有读取到配置文件时，新建一个。
func writeDefaultSecrets(secretFile string) {
	var defaultSecrets = []byte(`Bark:
  - url: "https://api.day.app/"
    secrets: ""
#    需要多个服务器可多复制几遍
#  - url: 推送服务器地址，默认"https://api.day.app/"
#    secrets: 你的推送密钥，格式为 "ABcDeFg1hIjkLmNOPQrstu"
#  - url: "https://api.day.app/"
#    secrets: "ABcDeFg1hIjkLmNOPQrstu"
WeWorkApp:
  - corpId: ""
    appSecret: ""
    agentId: ""
#  - corpId: "ww123456789a01b2c3"
#    appSecret: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"
#    agentId: "1000002"
#  - corpId: "ww123456789a01b2c3"
#    appSecret: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"
#    agentId: "1000002"`)
	err := os.WriteFile(secretFile, defaultSecrets, 0o644)
	if err != nil {
		log.Fatal("写入默认secrets文件失败!", err)
	} else {
		log.Info("写入默认secrets文件成功，请修改配置文件后重启程序。")
	}
	os.Exit(0)
}
