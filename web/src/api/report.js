import httpClient from '../utils/httpClient'
import host from './host'


export default {
    get_table_report (callback) {
        httpClient.get(host.goappHost + "api/report/table","",data => {
            let map = data.data.data

            let table_report = []
            for (let tablename in map) {
                // console.log(map[tablename])
                table_report.push({
                    tablename: tablename,
                    insertCount: map[tablename]['insert_count'],
                    updateCount: map[tablename]['update_count'],
                    deleteCount: map[tablename]['delete_count'],
                })
            }
            callback(table_report)
        })
    },

    get_schema_report (callback) {
        httpClient.get(host.goappHost + "api/report/schema","",data => {
            let map = data.data.data
            let schema_report = []
            for (let schema in map) {
                schema_report.push({
                    schemaname: schema,
                    ddlCount: map[schema]
                })
            }
            callback(schema_report)
        })
    },

    get_canal_state (callback) {
        httpClient.get(host.goappHost + 'api/canal/state' , "", data => {
            if (data.data.data) {
                callback("正常","enable",true)
            }else {
                callback("异常","disable",false)
            }

        })
    },

    get_canal_info (callback) {
        httpClient.get(host.goappHost + 'api/canal/info' , "", data => {
            callback(data.data.data)
        })

    },

    get_endpoint_state (callback) {
        httpClient.get(host.goappHost + 'api/transfer/endpoint/state' , "", data => {
            if (data.data.data) {
                callback("正常","enable",true)
            }else {
                callback("异常","disable",false)
            }

        })
    },

    get_endpoint_info (callback) {
        httpClient.get(host.goappHost + 'api/transfer/endpoint' , "", data => {
            callback(data.data.data)
        })
    },

    get_transfer_state (callback) {
        httpClient.get(host.goappHost + 'api/transfer/state' , "", data => {
            if (data.data.data) {
                callback("正常","enable",true)
            }else {
                callback("异常","disable",false)
            }

        })
    },

    set_canal_state (state ,gtid,callback) {
        let operator = "start"
        if (!state) {
            operator = "stop"
        }
        httpClient.post(host.goappHost + "api/canal/" + operator ,{gtid: gtid}, data => {
            callback(data.data)
        })
    },

    set_transfer_state (state , callback){
        // 目前 transfer 只有开启接口，没有停止接口
        let operator = "start"
        if (!state) {
            callback({
                msg: "禁止关闭 transfer",
                data: false
            })
            return
        }
        httpClient.post(host.goappHost + "api/transfer/" + operator ,{}, data => {
            callback(data.data)
        })
    }
}