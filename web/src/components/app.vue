<template>
	<div class="a1">
		<h1>{{ msg }}</h1>
		<br><br>
		<div class="state">
			<img  :src="require('../img/icon_' + canal_enable + '.svg')" height="10" width="20">
			canal状态 : {{ canal_status }}
			<i-switch class="switch" v-model="canal_run_status" :loading="canal_switch_loading" @on-change="switch_canal" true-color="#13ce66" false-color="#ff4949" />
			<Modal
					v-model="modal_canal"
					title="是否真的要关闭canal"
					@on-ok="stop_canal"
			>
				<p>关闭后会中断抓取数据库的 binlog</p>
			</Modal>
			<Modal
					v-model="modal_canal_start"
					title="是否要重新设置开始的 GTIDSet 位点"
					@on-ok="start_canal(new_gtid)"
			>
				<Tooltip max-width="300" >
					<p>   重新设置 canal gtidset </p>
					<div slot="content">
						<p> 新的gtid 位点</p>
						<p>
							<i>到数据库中执行 “show binary logs;” 可以查到所有的binlog文件</i>
							<br>
							<i>再去执行 “show BINLOG EVENTS in 'mysql-bin.000001' ;”  binlog名换成对应的文件名 可以看到其中 Gtid Event 有相应的gtidset</i>
						</p>
					</div>
				</Tooltip>
				<Input v-model="new_gtid" :placeholder=gtid_set clearable style="width: 450px" />
			</Modal>

		</div>
		<br>

		<div>
			<img :src="require('../img/icon_' + transfer_enable + '.svg')"  height="10" width="20">
			transfer状态 : {{ transfer_status }}
			<i-switch class="switch" v-model="transfer_run_status" :disabled="switch_transfer_disable" :loading="transfer_switch_loading" @on-change="switch_transfer" true-color="#13ce66" false-color="#ff4949" />

		</div>



		<Divider  />

		<div class="state">
			<h3> canal 信息 </h3>
			<div>
				<div class="div-inline" style="font-weight: bold">GTIDSet    </div><div class="div-inline">{{ gtid_set }}</div>
			</div>
			<div>
				<div class="div-inline" style="font-weight: bold">Binlog File </div><div class="div-inline"> {{ binlog_file }}</div>
				<div class="div-inline" style="margin-left: 20px;font-weight: bold">Binlog Pos  </div><div class="div-inline">{{ binlog_pos }}</div>
			</div>
		</div>
		<br>
		<div>
			<h3>endpoint 信息</h3>
			endpoint信息 : {{ endpoint_info }}
		</div>
		<div>
			<img :src="require('../img/icon_' + endpoint_enable + '.svg') "  height="10" width="20">
			endpoint状态 : {{ endpoint_status }}
<!--			<i-switch v-model="endpoint_run_status" :disabled="endpoint_run_status" true-color="#13ce66" false-color="#ff4949" />-->

		</div>

		<Divider />
		<h3> table dml 情况 </h3>
		<div class="report">
			<Table border :columns="columns_table" :data="data_table"></Table>
		</div>
		<br>
<!--		<Divider />-->
		<h3> schema ddl 情况 </h3>

		<div class="report">
			<Table border :columns="columns_schema" :data="data_schema"></Table>
		</div>

		<Divider />

		<!--图表-->
		<div class="report">
			<v-chart :options="msgCount_table"/>
		</div>

		<br><br>

		<div class="report">

			<v-chart :options="app_thread"/>
		</div>
	</div>

</template>

<script>
	import report from '../api/report.js'
	import metric from "../api/metric";
	import ECharts from 'vue-echarts/components/ECharts'
	import 'echarts/lib/chart/line'
	// import 'echarts/lib/component/polar'
	import 'echarts/lib/component/tooltip'
	import 'echarts/lib/component/title'
	import 'echarts/lib/component/toolbox'
	import 'echarts/lib/component/legend'



	export default {
		name: 'report',

		components: {
			'v-chart': ECharts
		},


		created() {
			this.refresh_all_data()

			this.timer()

		},
		data() {

			return {
				msg: "binlog transfer",
				columns_table: [
					{
						title: '表名',
						key: 'tablename'
					},
					{
						title: '插入行数',
						key: 'insertCount'
					},
					{
						title: '删除行数',
						key: 'deleteCount'
					},
					{
						title: '更新行数',
						key: 'updateCount'
					}
				],
				data_table: [
				],
				columns_schema: [
					{
						title: 'schema',
						key: 'schemaname'
					},
					{
						title: 'DDL次数',
						key: 'ddlCount'
					},
				],
				data_schema: [],
				canal_status: "",
				canal_enable: "disable",
				canal_run_status: false ,
				canal_switch_loading: false,
				gtid_set: "",
				binlog_file: "",
				binlog_pos: "",
				new_gtid: "",
				modal_canal: false,
				modal_canal_start: false,

				transfer_status: "",
				transfer_enable: "disable",
				transfer_run_status: false ,
				transfer_switch_loading: false,
				switch_transfer_disable: false,

				endpoint_info: "",
				endpoint_status: "",
				endpoint_enable: "disable",
				endpoint_run_status: false ,

				intervalId:null,

				msgCountData : [],
				msgCount_table: {
					title: {
						text: "待传输队列",
						subtext: "transfer 等待写入 endpoint 中的消息数量",
						left: "center",
						top: "auto",
						textStyle: {
							fontSize: 30
						},
						subtextStyle: {
							fontSize: 12
						}
					},
					tooltip: {
						trigger: 'axis',
						formatter: function (params) {
							params = params[0];
							var date = params.value[0].split(" ")[1];

							return  date + ' - msg count : ' + params.value[1];
						},
						axisPointer: {
							animation: false
						}
					},
					xAxis: {
						type: 'time',
						splitLine: {
							show: false
						}
					},
					yAxis: {
						type: 'value',
						boundaryGap: [0, '100%'],
						splitLine: {
							show: false
						}
					},
					series: [
						{
							data: [],
							type: 'line',
							showSymbol: false,
							hoverAnimation: false,
							smooth: true
						},
					],
				},

				app_thread: {
					title: {
						text: "线程情况",
						subtext: "线程和协程数量",
						left: "center",
						top: "auto",
						textStyle: {
							fontSize: 30
						},
						subtextStyle: {
							fontSize: 12
						}
					},
					legend: {
						// data: ['线程数',  '协程数'],
						y: 'bottom'

					},
					tooltip: {
						trigger: 'axis',
						// formatter: function (params) {
						// 	let a = params[0];
						// 	let date = a.value[0].split(" ")[1];
						// 	let b = params[1];
						// 	return  date + ' - 线程数 : ' + a.value[1] + '<br>' + date
						// 			+ ' - 协程数 : ' + b.value[1] ;
						// },
						// axisPointer: {
						// 	animation: false
						// }

					},
					xAxis: {
						type: 'time',
						splitLine: {
							show: false
						}
					},
					yAxis: {
						type: 'value',
						boundaryGap: [0, '100%'],
						splitLine: {
							show: false
						}
					},
					series: [
						{
							data: [],
							name: "线程数",
							type: 'line',
							showSymbol: false,
							hoverAnimation: false,
							smooth: true
						},
						{
							data: [],
							name: "协程数",
							type: 'line',
							showSymbol: false,
							hoverAnimation: false,
							smooth: true
						},
					],
				},

				c : 0
			}
		},
		methods: {
			refresh_graph () {
				metric.get_metric( data => {
					this.msgCount_table.series[0].data.push(data.transfer_msg)
					// console.log(this.msgCount_table.series)


					this.app_thread.series[0].data.push(data.go_threads)
					this.app_thread.series[1].data.push(data.go_goroutines)
					if (this.msgCount_table.series[0].data.length > 900) {
						this.msgCount_table.series[0].data.pop()
						this.app_thread.series[1].data.pop()
						this.app_thread.series[0].data.pop()
					}



				})
			},
			switch_canal (status) {
				if (!status) {
					// 告警确认是否真的要关闭
					this.modal_canal = true
				} else {
					this.modal_canal_start = true
				}
				this.refresh_canal_state()
			},
			start_canal (gtid) {
				this.canal_switch_loading = true
				report.set_canal_state(true, gtid ,data => {
					this.$Message.info('打开 canal：' + data.data + ",  " + data.msg);
					this.canal_switch_loading = false
					this.refresh_canal_state()
					return
				})
			},
			stop_canal(){
				this.canal_switch_loading = true
				report.set_canal_state(false, "",data => {
					this.$Message.info('关闭 canal：' + data.data + ",  " + data.msg);
					this.canal_switch_loading = false
					this.refresh_canal_state()
				})
			},

			switch_transfer(status) {
				if (status) {
					report.set_transfer_state(status,data => {
						this.$Message.info('操作 transfer：' + data.data + ",  " + data.msg);
					})
					this.refresh_canal_state()
				}
			},

			refresh_canal_state () {
				report.get_canal_state((data,enable,state) => {
					this.canal_status = data
					this.canal_enable = enable
					this.canal_run_status = state
				})
				report.get_canal_info( data => {
					this.gtid_set = data['gtid_set']
					this.binlog_file = data['binlog_file']
					this.binlog_pos = data['binlog_pos']
				})
			},
			refresh_transfer_state () {
				report.get_transfer_state((data , enable,state) => {
				this.transfer_status = data
				this.transfer_enable = enable
				this.transfer_run_status = state
					if (state) {
						this.switch_transfer_disable = true
					}
				})
			},
			refresh_endpoint_state () {
				report.get_endpoint_state((data ,enable,state) => {
					this.endpoint_status = data
					this.endpoint_enable = enable
					this.endpoint_run_status = state
				})
				report.get_endpoint_info(data => {
					this.endpoint_info = data
				})
			},
			refresh_all_data() {
				this.refresh_graph()
				this.refresh_canal_state()
				this.refresh_transfer_state()
				this.refresh_endpoint_state()

			},

			timer() {
				if (this.intervalId != null) {
					return;
				}
				// 计时器为空，操作
				this.intervalId = setInterval(() => {
					// console.log("刷新" + new Date());
					this.refresh_all_data(); //加载数据函数

					if (this.c % 5 == 0) {
						report.get_table_report(data => {
							this.data_table = data
						})
						report.get_schema_report(data => {
							this.data_schema = data
						})
					}

					this.c += 1
				}, 2000);
			},




		},

		destroyed() {
			clearInterval(this.intervalId);
			this.intervalId = null;
		},

	}
</script>

<style scoped>

	.state {}
	.a1 {
		font-family: monaco, "Helvetica Neue", Helvetica, "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "微软雅黑", Arial, sans-serif;
	}
	.switch {
		margin-left: 20px
	}
	.report {
		margin-left: 10% ;
		margin-right: 10% ;
	}
	.aa {
		margin-top: 15px;
		font-family: monaco, "Helvetica Neue", Helvetica, "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "微软雅黑", Arial, sans-serif;
		/*font-size: 20px;*/

		/*text-align:center;*/
		/*margin:0 auto;*/
		/*padding:0;*/
		/*clear:both;*/
	}
	.subdiv_allinline {
		margin:0;
		padding:0;
		display: inline-grid;
		/*_display:inline;*/
		/**display:inline;*/
		zoom:1;

	}
	.small {
		margin-left: 10%;
		margin-right: 10%;
		max-width: 600px;

	}
	.echarts {

		/*margin-left: 10%;*/
		/*margin-right: 10%;*/
		width: 100%;
		height: 400px;

	}
	.div-inline{ display:inline}
</style>
