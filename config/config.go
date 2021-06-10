package config

import (
	"github.com/buzhiyun/go-mysql-replication/utils"
	"github.com/juju/errors"
	"github.com/kataras/golog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var GlobalConfig *Config

const (
	_targetRedis         = "REDIS"
	_targetMongodb       = "MONGODB"
	_targetRocketmq      = "ROCKETMQ"
	_targetRabbitmq      = "RABBITMQ"
	_targetKafka         = "KAFKA"
	_targetElasticsearch = "ELASTICSEARCH"
	_targetScript        = "SCRIPT"

	RedisGroupTypeSentinel = "sentinel"
	RedisGroupTypeCluster  = "cluster"

	_dataDir = "data"

	_flushBulkInterval = 1000
	_flushBulkSize     = 100
)

var (
	_pid         int
	_coordinator int
	_leaderFlag  bool
	_leaderNode  string
	_currentNode string
	_bootTime    time.Time
)

type EmailConfig struct {
	Host        string   `yaml:"host"`
	Port        int      `yaml:"port"`
	LoginUser   string   `yaml:"login_user"`
	LoginPasswd string   `yaml:"login_passwd"`
	ToUser      []string `yaml:"to_user"`
}

type DingtalkConfig struct {
	WebhookUrl string `yaml:"webhook_url"`
	Secret     string `yaml:"secret"`
}

type WechatWorkConfig struct {
	WebhookUrl string `yaml:"webhook_url"`
}

type Config struct {
	Target string `yaml:"target"` // 目标类型，支持redis、mongodb

	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"pass"`
	Charset  string `yaml:"charset"`

	SlaveID uint32 `yaml:"slave_id"`
	Flavor  string `yaml:"flavor"`
	DataDir string `yaml:"data_dir"`

	OnlyTables []string `yaml:"only_tables"`
	OnlyDbs    []string `yaml:"only_dbs"`

	DumpExec       string `yaml:"mysqldump"`
	SkipMasterData bool   `yaml:"skip_master_data"`

	Maxprocs int   `yaml:"maxprocs"` // 最大协程数，默认CPU核心数*2
	BulkSize int64 `yaml:"bulk_size"`

	FlushBulkInterval int `yaml:"flush_bulk_interval"`

	FromGtidFile string `yaml:"from_gtid_file"` // 保持开始gtid位置的文件

	//SkipNoPkTable bool `yaml:"skip_no_pk_table"`

	LoggerConfig *LoggerConfig `yaml:"logger"` // 日志配置

	EnableExporter bool `yaml:"enable_exporter"` // 启用prometheus exporter，默认false
	ExporterPort   int  `yaml:"exporter_addr"`   // prometheus exporter端口

	EnableWebAdmin bool `yaml:"enable_web_admin"` // 启用Web监控，默认false
	WebAdminPort   int  `yaml:"web_admin_port"`   // web监控端口,默认8060

	// ------------------- KAFKA -----------------
	KafkaAddr         string `yaml:"kafka_addrs"`         //kafka连接地址，多个用逗号分隔
	KafkaSASLUser     string `yaml:"kafka_sasl_user"`     //kafka SASL_PLAINTEXT认证模式 用户名
	KafkaSASLPassword string `yaml:"kafka_sasl_password"` //kafka SASL_PLAINTEXT认证模式 密码
	KafkaTopic        string `yaml:"kafka_topic"`         //kafka 传输的topic
	KafkaVersion      string `yaml:"kafka_version"`

	IsReserveRawData bool //保留原始数据
	isMQ             bool //是否消息队列
	Notice           struct {
		Email           EmailConfig      `yaml:"email"`
		DingtalkRobot   DingtalkConfig   `yaml:"dingtalk_robot"`
		WechatWorkRobot WechatWorkConfig `yaml:"wechat_work_robot"`
	} `yaml:"notice""`
}

var SchemaMap, TableMap map[string]interface{}

func initConfig(fileName string) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return errors.Trace(err)
	}

	var c Config

	if err := yaml.Unmarshal(data, &c); err != nil {
		return errors.Trace(err)
	}

	if err := checkConfig(&c); err != nil {
		return errors.Trace(err)
	}

	//switch strings.ToUpper(c.Target) {
	//case _targetRedis:
	//	if err := checkRedisConfig(&c); err != nil {
	//		return errors.Trace(err)
	//	}
	//case _targetRocketmq:
	//	if err := checkRocketmqConfig(&c); err != nil {
	//		return errors.Trace(err)
	//	}
	//case _targetMongodb:
	//	if err := checkMongodbConfig(&c); err != nil {
	//		return errors.Trace(err)
	//	}
	//case _targetRabbitmq:
	//	if err := checkRabbitmqConfig(&c); err != nil {
	//		return errors.Trace(err)
	//	}
	//case _targetKafka:
	//	if err := checkKafkaConfig(&c); err != nil {
	//		return errors.Trace(err)
	//	}
	//case _targetElasticsearch:
	//	if err := checkElsConfig(&c); err != nil {
	//		return errors.Trace(err)
	//	}
	//case _targetScript:

	//default:
	//	return errors.Errorf("unsupported target: %s", c.Target)
	//}

	GlobalConfig = &c

	return nil
}

func checkConfig(c *Config) error {
	if c.Target == "" {
		return errors.Errorf("empty target not allowed")
	}

	if c.Addr == "" {
		return errors.Errorf("empty addr not allowed")
	}

	if c.User == "" {
		return errors.Errorf("empty user not allowed")
	}

	if c.Password == "" {
		return errors.Errorf("empty pass not allowed")
	}

	if c.Charset == "" {
		return errors.Errorf("empty charset not allowed")
	}

	if c.SlaveID == 0 {
		return errors.Errorf("empty slave_id not allowed")
	}

	if c.Flavor == "" {
		c.Flavor = "mysql"
	}

	if c.FlushBulkInterval == 0 {
		c.FlushBulkInterval = _flushBulkInterval
	}

	if c.BulkSize == 0 {
		c.BulkSize = _flushBulkSize
	}

	if c.DataDir == "" {
		c.DataDir = filepath.Join(utils.CurrentDirectory(), _dataDir)
	}

	if err := utils.MkdirIfNecessary(c.DataDir); err != nil {
		return err
	}

	if c.LoggerConfig == nil {
		c.LoggerConfig = &LoggerConfig{
			Dir: filepath.Join(c.DataDir, "log"),
		}
	}
	if c.LoggerConfig.Dir == "" {
		c.LoggerConfig.Dir = filepath.Join(c.DataDir, "log")
	}

	if err := utils.MkdirIfNecessary(c.LoggerConfig.Dir); err != nil {
		return err
	}

	if c.ExporterPort == 0 {
		c.ExporterPort = 9595
	}

	if c.WebAdminPort == 0 {
		c.WebAdminPort = 8060
	}

	if c.Maxprocs <= 0 {
		c.Maxprocs = runtime.NumCPU() * 2
	}

	if len(c.OnlyDbs) > 0 {
		SchemaMap = make(map[string]interface{}, len(c.OnlyDbs))
		for _, db := range c.OnlyDbs {
			SchemaMap[db] = ""
		}
	}
	if len(c.OnlyTables) > 0 {
		TableMap = make(map[string]interface{}, len(c.OnlyTables))
		for _, table := range c.OnlyTables {
			TableMap[table] = ""
		}
	}

	//if c.RuleConfigs == nil {
	//	return errors.Errorf("empty rules not allowed")
	//}

	//j ,_ := json.MarshalIndent(c,"","  ")
	//golog.Infof("%s",j)
	return nil
}

func Initialize(configPath string) error {
	if err := initConfig(configPath); err != nil {
		return err
	}
	runtime.GOMAXPROCS(GlobalConfig.Maxprocs)

	//streamHandler, err := sidlog.NewStreamHandler(golog.Writer())
	//if err != nil {
	//	return err
	//}
	//agent := sidlog.New(streamHandler, sidlog.Ltime|sidlog.Lfile|sidlog.Llevel)
	//sidlog.SetDefaultLogger(agent)
	//
	_bootTime = time.Now()
	_pid = syscall.Getpid()
	//
	//if _config.IsCluster(){
	//	if _config.EnableWebAdmin {
	//		_currentNode = _config.Cluster.BindIp + ":" + strconv.Itoa(_config.WebAdminPort)
	//	} else {
	//		_currentNode = _config.Cluster.BindIp + ":" + strconv.Itoa(_pid)
	//	}
	//}

	golog.Infof("process id: %d", _pid)
	golog.Infof("GOMAXPROCS :%d", GlobalConfig.Maxprocs)
	golog.Infof("source  %s(%s)", GlobalConfig.Flavor, GlobalConfig.Addr)
	//golog.Println(fmt.Sprintf("destination %s", GlobalConfig.Destination()))

	return nil
}

func (c *Config) IsKafka() bool {
	return strings.ToUpper(c.Target) == _targetKafka
}

func (c *Config) IsRabbitmq() bool {
	return strings.ToUpper(c.Target) == _targetRabbitmq
}
