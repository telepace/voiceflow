<h1 align="center" style="border-bottom: none">
    <b>
        <a href="https://github.com/telepace/voiceflow">voiceflow</a><br>
    </b>
</h1>
<h3 align="center" style="border-bottom: none">
      ⭐️ 基于 Go 语言的实时语音交互框架 ⭐️ <br>
<h3>


<p align=center>
<a href="https://goreportcard.com/report/github.com/telepace/voiceflow"><img src="https://goreportcard.com/badge/github.com/telepace/voiceflow" alt="A+"></a>
<a href="https://github.com/issues?q=org%telepace+is%3Aissue+label%3A%22good+first+issue%22+no%3Aassignee"><img src="https://img.shields.io/github/issues/telepace/voiceflow/good%20first%20issue?logo=%22github%22" alt="good first"></a>
<a href="https://github.com/telepace/voiceflow"><img src="https://img.shields.io/github/stars/telepace/voiceflow.svg?style=flat&logo=github&colorB=deeppink&label=stars"></a>
<a href="https://join.slack.com/t/telepace/shared_invite/zt-1se0k2bae-lkYzz0_T~BYh3rjkvlcUqQ"><img src="https://img.shields.io/badge/Slack-100%2B-blueviolet?logo=slack&amp;logoColor=white"></a>
<a href="https://github.com/telepace/voiceflow/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-green"></a>
<a href="https://golang.org/"><img src="https://img.shields.io/badge/Language-Go-blue.svg"></a>
</p>


<p align="center">
    <a href="./README.md"><b>English</b></a> •
    <a href="./README_zh-CN.md"><b>中文</b></a>
</p>

## 目录

- [简介](#简介)
- [快速开始](#快速开始)
  - [安装](#安装)
  - [配置](#配置)
- [用法](#用法)
- [架构图](#架构图)
- [支持的操作](#支持的操作)
- [实践指南](#实践指南)
- [贡献](#贡献)
- [许可证](#许可证)

## 简介

voiceflow 是一个基于 Go 语言的开源项目，旨在提供实时语音与大型语言模型（LLM）的交互能力。通过集成多种第三方语音平台和本地模型，voiceflow 支持实时语音转文本（STT）、文本转语音（TTS），以及与 LLM 的智能交互。

## 核心功能 🌟

- **实时语音转文本（STT）**：支持集成多家云服务商的 STT 服务和本地模型，实时将用户语音转换为文本。
- **与 LLM 交互**：将识别的文本直接发送给支持音频的 LLM，获取智能回复。
- **文本转语音（TTS）**：将 LLM 的回复文本转换为语音，支持多种 TTS 服务和本地模型。
- **音频存储与访问**：通过 MinIO 等存储服务，将生成的音频文件存储并提供访问路径，供前端实时播放。
- **可插拔的服务集成**：采用模块化设计，支持各个 STT、TTS 服务和 LLM 的可插拔式集成，方便扩展和定制。🎉



## 快速开始

### 安装

1. **克隆仓库**

   ```bash
   git clone https://github.com/telepace/voiceflow.git
   cd voiceflow
   ```

2. **安装依赖**

   确保已安装 Go 1.16 或更高版本。

   ```bash
   go mod tidy
   ```

### 配置

1. **复制示例环境变量文件**

   ```bash
   cp configs/.env.example configs/.env
   ```

   **编辑 `.env` 文件** 并填入相应的配置值：

   ```env
   # 环境变量示例
   MINIO_ENDPOINT=play.min.io
   MINIO_ACCESS_KEY=youraccesskey
   MINIO_SECRET_KEY=yoursecretkey
   AZURE_STT_KEY=yourazuresttkey
   AZURE_TTS_KEY=yourazurettskey
   ```

2. **配置 `config.yaml`**

   根据项目需求编辑 `configs/config.yaml`：

   ```yaml
   server:
     port: 8080
     enable_tls: false

   minio:
     enabled: true
     bucket_name: voiceflow-audio

   stt:
     provider: azure  # 可选值：azure、google、local

   tts:
     provider: google  # 可选值：azure、google、local

   llm:
     provider: openai  # 可选值：openai、local

   logging:
     level: info
   ```

### 启动应用

在项目根目录下运行：

```bash
go run cmd/main.go
```

查看服务是否正常启动，访问 `http://localhost:8080`。

### 系统架构

```mermaid
graph TD
    A[前端浏览器] --> B[WebSocket 服务器]
    B --> C[语音转文本 (STT)]
    C --> D[大型语言模型 (LLM)]
    D --> E[文本转语音 (TTS)]
    E --> F[存储服务 (MinIO)]
    F --> B
    B --> A
```

- **前端浏览器**：用户通过浏览器录制语音，并通过 WebSocket 发送到服务器。
- **WebSocket 服务器**：接收前端的音频数据，协调各个服务模块。
- **语音转文本。(STT)**：将音频数据转换为文本。
- **大型语言模型 (LLM)**：根据文本生成智能回复。
- **文本转语音 (TTS)**：将回复文本转换为语音数据。
- **存储服务 (MinIO)**：存储生成的音频文件，并提供访问 URL。

### 目录结构

```bash
voiceflow/
├── cmd/
│   └── main.go            # 应用程序入口
├── configs/
│   ├── config.yaml        # 业务配置文件
│   └── .env               # 环境变量文件
├── internal/
│   ├── config/            # 配置加载模块
│   ├── server/            # WebSocket 服务器
│   ├── stt/               # 语音转文本模块
│   ├── tts/               # 文本转语音模块
│   ├── llm/               # LLM 交互模块
│   ├── storage/           # 存储模块
│   ├── models/            # 数据模型
│   └── utils/             # 工具函数
├── pkg/
│   └── logger/            # 日志模块
├── scripts/               # 构建和部署脚本
├── go.mod                 # Go 模块文件
└── README.md              # 项目说明文档
```

### 核心模块

1. **WebSocket 服务器**

   使用 `gorilla/websocket` 实现，负责与前端的实时通信，接收音频数据并返回处理结果。

2. **语音转文本 (STT)**

   - **接口定义**：`internal/stt/stt.go` 定义了 STT 服务的接口。
   - **可插拔实现**：支持 Azure、Google、本地模型等多种实现方式。

3. **文本转语音 (TTS)**

   - **接口定义**：`internal/tts/tts.go` 定义了 TTS 服务的接口。
   - **可插拔实现**：支持 Azure、Google、本地模型等多种实现方式。

4. **大型语言模型 (LLM)**

   - **接口定义**：`internal/llm/llm.go` 定义了与 LLM 交互的接口。
   - **可插拔实现**：支持 OpenAI、本地模型等多种实现方式。

5. **存储模块**

   - **接口定义**：`internal/storage/storage.go` 定义了存储服务的接口。
   - **实现方式**：默认使用 MinIO 进行音频文件的存储，也支持本地文件系统。

### TODO

1. 消息总线
2. 配置中心
3. 容器化部署
4. hooks

### 参考

- [https://openai.com/index/hello-gpt-4o/](https://openai.com/index/hello-gpt-4o/)
- [https://medium.com/@artificial--intelligence/the-differences-between-asr-and-tts-c85a08269c98](https://medium.com/@artificial--intelligence/the-differences-between-asr-and-tts-c85a08269c98#:~:text=We%20are%20familiar%20with%20the,analogous%20to%20the%20human%20mouth.)

### 参与贡献

我们欢迎任何形式的贡献！请阅读 CONTRIBUTING.md 了解更多信息。

- **提交问题**：如果您发现了 Bug，或者有新的功能建议，请在 Issues 中提交。
- **贡献代码**：Fork 本仓库，在您的分支上进行修改，提交 Pull Request。

### 开源协议

voiceflow 使用 [MIT](./LICENSE) 开源协议。

### 致谢

感谢所有为本项目做出贡献的开发者！

<a href="https://github.com/telepace/voiceflow/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=telepace/voiceflow" />
