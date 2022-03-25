package prometheus

import (
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/buzhiyun/go-mysql-replication/model"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	canalEnable = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "canal_enable",
			Help: "canal state : 0=false ,1=true",
		})

	transMsg = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "transfer_msg",
			Help: "transfer 消息队列中的消息数 ",
		})

	transferEnable = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "transfer_enable",
			Help: "transfer state : 0=false ,1=true",
		})

	endpointEnable = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "endpoint_enable",
			Help: "endpoint state : 0=false ,1=true",
		})

	insertCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transfer_inserted_num",
			Help: "The number of data inserted to destination",
		}, []string{"table"},
	)

	updateCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transfer_updated_num",
			Help: "The number of data updated to destination",
		}, []string{"table"},
	)

	deleteCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transfer_deleted_num",
			Help: "The number of data deleted to destination",
		}, []string{"table"},
	)

	ddlCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transfer_ddl_num",
			Help: "The number of data ddl to destination",
		}, []string{"schema"},
	)
)

func Initialize() error {

	prometheus.MustRegister(canalEnable)
	prometheus.MustRegister(transferEnable)
	prometheus.MustRegister(endpointEnable)
	prometheus.MustRegister(transMsg)

	// 改到 iris 上直接注册到web 端口的 /metric 上
	//if config.GlobalConfig.EnableExporter {
	//	go func() {
	//		http.Handle("/", promhttp.Handler())
	//		http.ListenAndServe(fmt.Sprintf(":%d", config.GlobalConfig.ExporterPort), nil)
	//	}()
	//}
	if config.GlobalConfig.EnableWebAdmin {
		insertRecord = make(map[string]uint64)
		updateRecord = make(map[string]uint64)
		deleteRecord = make(map[string]uint64)
		ddlRecord = make(map[string]uint64)
		for table, v := range model.TableInfo {
			insertRecord[table] = 0
			updateRecord[table] = 0
			deleteRecord[table] = 0
			ddlRecord[v.TableInfo.Schema] = 0
		}

	}
	return nil
}

func UpdateActionNum(action, lab string) {
	if config.GlobalConfig.EnableExporter {
		switch action {
		case canal.InsertAction:
			insertCounter.WithLabelValues(lab).Inc()
		case canal.UpdateAction:
			updateCounter.WithLabelValues(lab).Inc()
		case canal.DeleteAction:
			deleteCounter.WithLabelValues(lab).Inc()
		case "ddl":
			ddlCounter.WithLabelValues(lab).Inc()
		}
	}
	if config.GlobalConfig.EnableWebAdmin {
		switch action {
		case canal.InsertAction:
			if _, ok := insertRecord[lab]; ok {
				insertRecord[lab] += 1
			}
		case canal.UpdateAction:
			if _, ok := updateRecord[lab]; ok {
				updateRecord[lab] += 1
			}
		case canal.DeleteAction:
			if _, ok := deleteRecord[lab]; ok {
				deleteRecord[lab] += 1
			}
		case "ddl":
			if _, ok := ddlRecord[lab]; ok {
				ddlRecord[lab] += 1
			}

		}

	}
}

func AddTransferMsg() {
	transMsg.Inc()
}

func DecTransferMsg() {
	transMsg.Dec()
}

func SetCanalState(state int) {
	canalEnable.Set(float64(state))
}

func SetTransferState(state int) {
	transferEnable.Set(float64((state)))
}

func SetEndpointState(state int) {
	endpointEnable.Set(float64((state)))
}
