import React, { Component } from 'react'
import axios from 'axios'

export default class Index extends Component {
  constructor(props) {
    super(props)
    this.state = {
      file: null,
      msg: "",
    }
  }

  async fileChangedHandler(event) {
    try {
      const file = event.target.files[0]
      this.setState({
        file: file,
      })
    } catch (err) {
      this.setState({
        msg: err.toString(),
      })
    }
  }

  async upload(event) {
    try {
      event.preventDefault()

      const fileChunkSize = 10000000
      const numChunks = Math.floor(this.state.file.size / fileChunkSize) + 1

      const getUrlRequests = []
      const putObjectRequests = []
      const uploadPartsArray = []

      let start = ""
      let end = ""
      let blob = ""

      // 1. Calls the CreateMultipartUpload endpoint in the backend server
      const createMultipartUploadResponse = await axios.get('/api/create-multipart-upload', {
        params: {
          fileName: this.state.file.name,
        }
      })
      const uploadId = createMultipartUploadResponse.data

      for (let index = 1; index < numChunks + 1; index++) {
        getUrlRequests.push(
          axios.get('/api/get-upload-url', {
            params: {
              fileName: this.state.file.name,
              partNumber: index,
              uploadId: uploadId,
            }
          })
        )
      }

      // 2. Generate presigned URL for each part
      const getUrlResponses = await Promise.all(getUrlRequests)
      getUrlResponses.forEach((getUrlResponse, index) => {
        start = (index) * fileChunkSize
        end = (index + 1) * fileChunkSize
        blob = index + 1 < numChunks ? this.state.file.slice(start, end) : this.state.file.slice(start)
        putObjectRequests.push(
          axios.put(
            getUrlResponse.data,
            blob,
          )
        )
      })

      // 3. Puts each file part into the storage server
      const putObjectResponses = await Promise.all(putObjectRequests)
      putObjectResponses.forEach((putObjectResponse, index) => {
        uploadPartsArray.push({
          ETag: putObjectResponse.headers.etag,
          PartNumber: index + 1,
        })
      })

      // 4. Calls the CompleteMultipartUpload endpoint in the backend server
      const completeUploadResp = await axios.post('/api/complete-multipart-upload', {
        params: {
          fileName: this.state.file.name,
          parts: uploadPartsArray,
          uploadId: uploadId,
        }
      })
      this.setState({
        msg: completeUploadResp.data,
      })
    } catch (err) {
      this.setState({
        msg: err.toString(),
      })
    }
  }

  render() {
    return (
      <div>
        <form onSubmit={this.upload.bind(this)}>
          <input type='file' id='file' onChange={this.fileChangedHandler.bind(this)} />
          <button type='submit'>Upload</button>
        </form>
        <div id="msg">{this.state.msg}</div>
      </div>
    )
  }
}
