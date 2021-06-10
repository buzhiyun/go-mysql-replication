import axios from 'axios'
import Formdata from 'form-data'
axios.defaults.timeout = 10000
axios.defaults.responseType = 'json'


function handle (promise, next) {
  promise.then((res) => successCallback(res, next))
    .catch((error) => failureCallback(error))
}

function successCallback (res, next) {
  if (!checkResponseCode(res.data.code, res.data.msg)) {
    return
  }
  if (!next) {
    return
  }
  next(res.data.data, res.data.code, res.data.msg)
}

function failureCallback (error) {
  // Message.error({
  //   message: '请求失败 - ' + error
  // })
}



export default {
  get (uri, config, next) {
    const promise = axios.get(uri, config).then(next)
    // handle(promise, next)
  },

  batchGet (uriGroup, next) {
    const requests = []
    for (let item of uriGroup) {
      let params = {}
      if (item.params !== undefined) {
        params = item.params
      }
      requests.push(axios.get(item.uri, {params}))
    }

    axios.all(requests).then(axios.spread(function (...res) {
      const result = []
      for (let item of res) {
        if (!checkResponseCode(item.data.code, item.data.message)) {
          return
        }
        result.push(item.data.data)
      }
      next(...result)
    })).catch((error) => failureCallback(error))
  },

  post (uri, data, next) {
    const promise = axios.post(uri, data, {
    // const promise = axios.post(uri, Qs.stringify(data), {
      headers: {
        post: {
          'Content-Type': 'application/json'
        }
      }
    }).then(next)
    // handle(promise, next)
  }
}