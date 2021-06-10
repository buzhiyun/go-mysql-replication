import httpClient from '../utils/httpClient'
import host from "./host";

function dateFormat(fmt, date) {
    let ret;
    const opt = {
        "Y+": date.getFullYear().toString(),        // 年
        "m+": (date.getMonth() + 1).toString(),     // 月
        "d+": date.getDate().toString(),            // 日
        "H+": date.getHours().toString(),           // 时
        "M+": date.getMinutes().toString(),         // 分
        "S+": date.getSeconds().toString()          // 秒
        // 有其他格式化字符需求可以继续添加，必须转化成字符串
    };
    for (let k in opt) {
        ret = new RegExp("(" + k + ")").exec(fmt);
        if (ret) {
            fmt = fmt.replace(ret[1], (ret[1].length == 1) ? (opt[k]) : (opt[k].padStart(ret[1].length, "0")))
        };
    };
    return fmt;
}

export default {

    get_metric(callback) {
        httpClient.get(host.goappHost +'metric',{
            headers: {
                'Content-Type': 'text/plain;charset=UTF-8'
            },
            responseType: 'text/plain'
        }, data => {
            // console.log(data.data)
            // let d = data.data
            let date = new Date()
            let timeStr = dateFormat("YYYY-mm-dd HH:MM:SS", date)
            let metric = {}
            data.data.split("\n").forEach(
                item => {
                    if (!item.startsWith("#") && item.length > 0 ) {
                        if (item.startsWith("transfer_msg ")) {
                            let i = item.split(" ")
                            metric.transfer_msg = [timeStr,i[1]]
                        }
                    }

                    if (!item.startsWith("#") && item.length > 0 ) {
                        if (item.startsWith("go_threads ")) {
                            let i = item.split(" ")
                            metric.go_threads = [timeStr,i[1]]
                        }
                    }
                    if (!item.startsWith("#") && item.length > 0 ) {
                        if (item.startsWith("go_goroutines ")) {
                            let i = item.split(" ")
                            metric.go_goroutines = [timeStr,i[1]]
                        }
                    }
                }
            )
            // console.log(metric)
            callback(metric)
        })
    }
}
