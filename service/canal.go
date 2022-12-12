package service

import (
	"fmt"
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/buzhiyun/go-mysql-replication/model"
	"github.com/buzhiyun/go-mysql-replication/prometheus"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/kataras/golog"
	"io/ioutil"
	"strings"
	"time"

	"sync"
)

type Canal struct {
	canal        *canal.Canal
	canalCfg     *canal.Config
	canalHandler *canalHandler
	canalEnable  bool
	lockOfCanal  sync.Mutex
	firstsStart  bool
	wg           sync.WaitGroup
}

func (c *Canal) GetCanelState() bool {
	return c.canalEnable
}

func (c *Canal) initialize() (err error) {
	c.canalCfg = canal.NewDefaultConfig()
	c.canalCfg.Addr = config.GlobalConfig.Addr
	c.canalCfg.User = config.GlobalConfig.User
	c.canalCfg.Password = config.GlobalConfig.Password
	c.canalCfg.Charset = config.GlobalConfig.Charset
	c.canalCfg.Flavor = config.GlobalConfig.Flavor
	c.canalCfg.ServerID = config.GlobalConfig.SlaveID
	c.canalCfg.Dump.ExecutionPath = config.GlobalConfig.DumpExec
	c.canalCfg.Dump.DiscardErr = false
	c.canalCfg.Dump.SkipMasterData = config.GlobalConfig.SkipMasterData
	c.canalCfg.Dump.Databases = config.GlobalConfig.OnlyDbs
	if len(config.GlobalConfig.OnlyTables) > 0 && len(config.GlobalConfig.OnlyDbs) == 1 {
		c.canalCfg.Dump.Tables = config.GlobalConfig.OnlyTables
		c.canalCfg.Dump.TableDB = config.GlobalConfig.OnlyDbs[0]
	}

	if err := c.createCanal(); err != nil {
		return err
	}

	// 同步表结构
	c.getTableDDL()

	//c.addDumpDatabaseOrTable()

	return
}

func (c *Canal) fixGtid() (fromGtidSetStr string) {
	var fromGtidSet string
	if len(config.GlobalConfig.FromGtidFile) > 0 {
		gtid, err := ioutil.ReadFile(config.GlobalConfig.FromGtidFile)
		if err != nil {
			golog.Errorf("gtid 文件读取 %s 失败 : %v", config.GlobalConfig.FromGtidFile, err.Error())
		} else {
			fromGtidSet = string(gtid)
		}

		// 先保存一遍看文件能否保存
		err = ioutil.WriteFile(config.GlobalConfig.FromGtidFile, gtid, 0777)
		if err != nil {
			golog.Error("from_gtid_file 配置的文件权限异常： ", err.Error())
			Close()
		}
	}

	// 改造传入的gtid位点，适应master
	if len(fromGtidSet) > 0 {
		gs := strings.Split(fromGtidSet, ":")
		if len(gs) != 2 {
			golog.Errorf("初始gtid解析错误, %v", fromGtidSet)
			return
		}
		fromUUID := gs[0]
		fromPos := gs[1]

		// 查询 gtid 位点
		res, err := c.canal.Execute("SHOW VARIABLES LIKE 'gtid_purged';")
		if err != nil {
			golog.Warnf("查询 master GTID 位点出错，从最开始 binlog 开始")
		} else {
			//golog.Infof("res:  %s", res.Values[0][1])
			gtid, err := res.GetString(0, 1)
			if err != nil {
				golog.Errorf("获取master gtid_purged 出错")
			} else {
				gtids := strings.Split(gtid, ",")
				var masterUUID, masterPos string
				for i, s := range gtids {
					g := strings.Split(s, ":")
					masterUUID = strings.TrimSpace(g[0])
					masterPos = strings.TrimSpace(g[1])
					if masterUUID == fromUUID {
						gtids[i] = strings.TrimSpace(fromUUID + ":" + strings.Split(masterPos, "-")[0] + "-" + fromPos)
					} else {
						gtids[i] = strings.TrimSpace(s)
					}
				}

				fromGtidSetStr = strings.Join(gtids, ",")
			}
		}

	}

	return
}

func (c *Canal) getTableDDL() {

	len_db := len(config.GlobalConfig.OnlyDbs)
	len_table := len(config.GlobalConfig.OnlyTables)

	sql_where := ""
	var sql_where_table, sql_where_db, sql_where_and string
	if len_db > 0 || len_table > 0 {
		sql_where = " WHERE "

		if len_db > 0 {
			sql_where_db = fmt.Sprintf(`table_schema in ("%s")`, strings.Join(config.GlobalConfig.OnlyDbs, `","`))
		}
		if len_table > 0 {
			sql_where_table = fmt.Sprintf(`table_name in ("%s")`, strings.Join(config.GlobalConfig.OnlyTables, `","`))
		}
		if len_table > 0 && len_db > 0 {
			sql_where_and = "AND"
		}

	}

	sql := `SELECT table_name,table_schema FROM information_schema.tables` + sql_where + sql_where_db + sql_where_and + sql_where_table + ";"
	golog.Info("查询表结构 ", sql)

	res, err := c.canal.Execute(sql)
	if err != nil {
		golog.Error("查询异常 ", err)
		return
	}

	for i := 0; i < res.Resultset.RowNumber(); i++ {
		tableName, _ := res.GetString(i, 0)
		schemaName, _ := res.GetString(i, 1)
		table := schemaName + "." + tableName
		golog.Debugf("获取表 %v.%v 结构", schemaName, tableName)
		tableMeta, err := c.canal.GetTable(schemaName, tableName)
		if err != nil {
			golog.Error("获取表结构出错 ", err)
			return
		}
		model.TableInfo[table] = model.TableInfomation{
			TableInfo:       tableMeta,
			TableColumnSize: len(tableMeta.Columns),
		}
	}

}

func (c *Canal) StartUpFromGtidSet() {

	c.lockOfCanal.Lock()
	defer c.lockOfCanal.Unlock()

	// 这段写的有点繁琐  后期改造
	if c.firstsStart {
		c.canalHandler = newCanalHandler()
		if c.canalHandler == nil {
			golog.Warnf("no handler  !!!!!")
		}
		if c.canal == nil {
			golog.Warnf("canal is null  !!!!!")
		}

		c.canal.SetEventHandler(c.canalHandler)
		c.firstsStart = false
		gtid := c.fixGtid()
		golog.Info("GTID set : ", gtid)
		c.runFromGtidSet(gtid)
	} else {
		//c.restart()
	}
}

func (c *Canal) StartUpFromPosition(p mysql.Position) {
	c.lockOfCanal.Lock()
	defer c.lockOfCanal.Unlock()

	if c.firstsStart {
		c.canalHandler = newCanalHandler()
		c.canal.SetEventHandler(c.canalHandler)
		//c.canalHandler.Start()
		c.firstsStart = false
		c.runFromPosition(p)
	} else {
		//c.restart()
	}
}

func (c *Canal) runFromGtidSet(fromGtidSet string) (err error) {
	c.wg.Add(1)
	gtidset, err := mysql.ParseMysqlGTIDSet(fromGtidSet)
	if err != nil {
		return
	}
	go func(gtid mysql.GTIDSet) {
		//go func(p mysql.Position) {
		c.canalEnable = true
		prometheus.SetCanalState(1)
		golog.Infof("Canal run from position(%s)", gtid.String())

		if err := c.canal.StartFromGTID(gtid); err != nil {
			golog.Errorf("canal : %v", err)
			c.canalEnable = false
			prometheus.SetCanalState(0)

		}

		golog.Info("Canal is Closed")
		c.canalEnable = false
		prometheus.SetCanalState(0)

		c.canal = nil
		c.wg.Done()
	}(gtidset)

	// canal未提供回调，停留几秒，确保StartFrom启动成功
	time.Sleep(3 * time.Second)
	return nil
	//return
}

func (c *Canal) runFromPosition(position mysql.Position) (err error) {
	c.wg.Add(1)
	if err != nil {
		return
	}
	go func(p mysql.Position) {
		//go func(p mysql.Position) {
		c.canalEnable = true
		prometheus.SetCanalState(1)

		golog.Infof("transfer run from position(%s %d)", p.Name, p.Pos)

		golog.Infof("start transfer : %v", err)

		if err := c.canal.RunFrom(position); err != nil {
			golog.Errorf("canal : %v", err)
			//if c.canalHandler != nil {
			//	c.canalHandler.stopListener()
			//}
			c.canalEnable = false
			prometheus.SetCanalState(0)

		}

		golog.Info("Canal is Closed")
		c.canalEnable = false
		prometheus.SetCanalState(0)

		c.canal = nil
		c.wg.Done()
	}(position)

	// canal未提供回调，停留一秒，确保RunFrom启动成功
	time.Sleep(time.Second)
	return nil
	//return
}

func (c *Canal) createCanal() (err error) {
	c.canal, err = canal.NewCanal(c.canalCfg)
	if err != nil {
		golog.Errorf("canal 初始化异常, %s", err.Error())
	}
	return
}

func (c *Canal) addDumpDatabaseOrTable() {
	if len(config.GlobalConfig.OnlyDbs) > 0 {
		c.canal.AddDumpDatabases(config.GlobalConfig.OnlyDbs...)
		golog.Infof("add db %s", config.GlobalConfig.OnlyDbs)
		return
	}
	if len(config.GlobalConfig.OnlyTables) > 0 && len(config.GlobalConfig.OnlyDbs) == 1 {
		c.canal.AddDumpTables(config.GlobalConfig.OnlyDbs[0], config.GlobalConfig.OnlyTables...)
	}
}

func (c *Canal) stopDump() {
	c.lockOfCanal.Lock()
	defer c.lockOfCanal.Unlock()

	if c.canal == nil {
		return
	}

	if !c.canalEnable {
		return
	}

	c.canal.Close()
	c.wg.Wait()

	c.canalEnable = false
	prometheus.SetCanalState(0)

	golog.Println("dumper stopped")
}

func (c *Canal) Close() {
	c.stopDump()
}

func (c *Canal) StartManual() (msg string) {
	if c.canalEnable {
		return "canal 已经启动，无需再次启动"
	}
	err := c.createCanal()
	if err != nil {
		return err.Error()
	}
	c.addDumpDatabaseOrTable()
	c.canalHandler = newCanalHandler()
	c.canal.SetEventHandler(c.canalHandler)
	err = c.runFromGtidSet(c.fixGtid())
	if err != nil {
		return err.Error()
	}
	return "发送启动信号 ok"
}
