package secrets

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
	"sync"
)

var SecretFile string

var Secrets = &TypeSecrets{}

type TypeSecrets struct {
	Barks      []BarkServer `yaml:"Bark"`
	WeworkApps []WeworkApp  `yaml:"WeWorkApp"`
}

// BarkServer Bark消息推送(iOS)
type BarkServer struct {
	ServerUrl   string `yaml:"url"`
	BarkSecrets string `yaml:"secrets"`
}

// WeworkApp 企业微信应用消息
type WeworkApp struct {
	sync.Mutex
	CorpId                    string `yaml:"corpId"`
	AppSecret                 string `yaml:"appSecret"`
	AgentID                   string `yaml:"agentId"`
	WeworkAccessToken         string
	WeworkAccessTokenExpireAt int64
}

func init() {
	flag.StringVar(&SecretFile, "s", "secrets.yml", "secret file")
	flag.Parse()
	file, err := os.ReadFile(SecretFile)
	if err == nil {
		err = yaml.NewDecoder(strings.NewReader(expand(string(file), os.Getenv))).Decode(Secrets)
		if err != nil {
			log.Fatal("配置文件不合法!", err)
		}
	} else {
		writeDefaultSecrets()
	}
}

// expand 使用正则进行环境变量展开
// os.ExpandEnv 字符 $ 无法逃逸
// https://github.com/golang/go/issues/43482
func expand(s string, mapping func(string) string) string {
	r := regexp.MustCompile(`\${([a-zA-Z_]+[a-zA-Z0-9_:/.]*)}`)
	return r.ReplaceAllStringFunc(s, func(s string) string {
		s = strings.Trim(s, "${}")
		before, after, ok := strings.Cut(s, ":")
		m := mapping(before)
		if ok && m == "" {
			return after
		}
		return m
	})
}

func writeDefaultSecrets() {
	var defaultSecrets = `Bark:
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
#    agentId: "1000002"`
	_ = os.WriteFile(SecretFile, []byte(defaultSecrets), 0o644)
	os.Exit(0)
}
