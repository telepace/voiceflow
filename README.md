<h1 align="center" style="border-bottom: none">
    <b>
        <a href="https://github.com/telepace/voiceflow">voiceflow</a><br>
    </b>
</h1>
<h3 align="center" style="border-bottom: none">
    ⭐️ Real-time Voice Interaction Framework based on Go ⭐️ <br>
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
    <a href="./README_zh-CN.md"><b>中文 (Chinese)</b></a>
</p>

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Configuration](#configuration)
- [Usage](#usage)
- [Architecture Diagram](#architecture-diagram)
- [Supported Operations](#supported-operations) - [Practical Guide](#practical-guide) - [Contributing](#contributing)
- [License](#license)

## Introduction

`voiceflow` is an open-source project built with Go, designed to enable real-time voice interaction with Large Language Models (LLMs). By integrating various third-party voice platforms and local models, `voiceflow` supports real-time Speech-to-Text (STT), Text-to-Speech (TTS), and intelligent interaction with LLMs.

## Core Features 🌟

-   **Real-time Speech-to-Text (STT)**: Integrates with multiple cloud STT services (e.g., Azure, Google) and local models to convert user speech into text in real-time.
-   **LLM Interaction**: Sends the recognized text directly to audio-capable LLMs to obtain intelligent responses.
-   **Text-to-Speech (TTS)**: Converts the LLM's text responses back into speech, supporting various TTS services (e.g., Azure, Google) and local models.
-   **Audio Storage & Access**: Utilizes storage services like MinIO to store generated audio files and provide access URLs for real-time playback on the frontend.
-   **Pluggable Service Integration**: Features a modular design allowing for pluggable integration of different STT, TTS services, and LLMs, facilitating easy extension and customization. 🎉

## Quick Start

### Installation

1.  **Clone the Repository**

    ```bash
    git clone https://github.com/telepace/voiceflow.git
    cd voiceflow
    ```

2.  **Install Dependencies**

    Ensure you have Go 1.16 or higher installed.

    ```bash
    go mod tidy
    ```

### Configuration

1.  **Copy the Example Environment File**

    ```bash
    cp configs/.env.example configs/.env
    ```

    **Edit the `.env` file** and fill in the appropriate configuration values:

    ```env
    # Example Environment Variables
    MINIO_ENDPOINT=play.min.io        # Your MinIO server endpoint
    MINIO_ACCESS_KEY=youraccesskey    # Your MinIO access key
    MINIO_SECRET_KEY=yoursecretkey    # Your MinIO secret key
    AZURE_STT_KEY=yourazuresttkey     # Your Azure Speech-to-Text service key
    AZURE_TTS_KEY=yourazurettskey     # Your Azure Text-to-Speech service key
    # Add other necessary keys (e.g., Google Cloud, OpenAI API keys) as needed
    ```

2.  **Configure `config.yaml`**

    Edit `configs/config.yaml` according to your project requirements:

    ```yaml
    server:
      port: 8080          # Port the server will listen on
      enable_tls: false   # Set to true to enable TLS/SSL

    minio:
      enabled: true       # Set to true to enable MinIO storage
      bucket_name: voiceflow-audio # Name of the MinIO bucket for audio files

    stt: # Speech-to-Text Configuration
      provider: azure     # Options: azure, google, local (choose your STT provider)
      # Add provider-specific settings here if needed

    tts: # Text-to-Speech Configuration
      provider: google    # Options: azure, google, local (choose your TTS provider)
      # Add provider-specific settings here if needed

    llm: # Large Language Model Configuration
      provider: openai    # Options: openai, local (choose your LLM provider)
      # Add provider-specific settings here (e.g., API key, model name)

    logging:
      level: info         # Logging level (e.g., debug, info, warn, error)
    ```

### Start the Application

Run the following command in the project root directory:

```bash
go run cmd/main.go

```

Check if the service has started correctly by accessing `http://localhost:8080` (or your configured port).

## **Architecture Diagram**

```
graph TD
    A["Frontend (Browser)"] --> B["WebSocket Server (Go Backend)"]
    B --> C["Speech-to-Text (STT) Module"]
    C --> D["Large Language Model (LLM) Module"]
    D --> E["Text-to-Speech (TTS) Module"]
    E --> F["Storage Service (e.g., MinIO)"]
    F --> B  ["Provides Audio URL"]
    B --> A  ["Sends Audio URL/Data"]

```

- **Frontend (Browser)**: The user records voice input via the browser, sending audio data through a WebSocket connection to the server.
- **WebSocket Server**: Receives audio data from the frontend and orchestrates the workflow between different service modules.
- **Speech-to-Text (STT) Module**: Converts the incoming audio data into text.
- **Large Language Model (LLM) Module**: Processes the text from STT and generates an intelligent response.
- **Text-to-Speech (TTS) Module**: Converts the LLM's text response back into audio data.
- **Storage Service (MinIO)**: Stores the generated audio files and provides accessible URLs for playback.

## **Directory Structure**

```
voiceflow/
├── cmd/
│   └── main.go              # Application entry point
├── configs/
│   ├── config.yaml          # Business logic configuration file
│   └── .env                 # Environment variables file (sensitive keys, etc.)
├── internal/
│   ├── config/              # Configuration loading module
│   ├── server/              # WebSocket server implementation
│   ├── stt/                 # Speech-to-Text module (interfaces, implementations)
│   ├── tts/                 # Text-to-Speech module (interfaces, implementations)
│   ├── llm/                 # LLM interaction module (interfaces, implementations)
│   ├── storage/             # Storage module (interfaces, implementations like MinIO)
│   ├── models/              # Data models/structs used across the application
│   └── utils/               # Utility functions
├── pkg/
│   └── logger/              # Logging module setup
├── scripts/                 # Build and deployment scripts (if any)
├── go.mod                   # Go modules file (dependencies)
├── go.sum                   # Go modules checksum file
└── README.md                # Project description (this file)

```

## **Core Modules**

1. **WebSocket Server**
    - Implemented using `gorilla/websocket`.
    - Handles real-time communication with the frontend, receiving audio data and sending back processing results (like audio URLs).
2. **Speech-to-Text (STT)**
    - **Interface Definition**: `internal/stt/stt.go` defines the standard interface for STT services.
    - **Pluggable Implementations**: Supports various providers like Azure, Google Cloud Speech, and potentially local models. New providers can be added by implementing the interface.
3. **Text-to-Speech (TTS)**
    - **Interface Definition**: `internal/tts/tts.go` defines the standard interface for TTS services.
    - **Pluggable Implementations**: Supports various providers like Azure, Google Cloud Text-to-Speech, and potentially local models.
4. **Large Language Model (LLM)**
    - **Interface Definition**: `internal/llm/llm.go` defines the interface for interacting with LLMs.
    - **Pluggable Implementations**: Supports providers like OpenAI (GPT models) and potentially local LLMs.
5. **Storage Module**
    - **Interface Definition**: `internal/storage/storage.go` defines the interface for storage services.
    - **Implementation**: Defaults to using MinIO for object storage (ideal for audio files) but can be adapted to use local file systems or other cloud storage providers.

## **TODO**

- [ ]  Implement a Message Bus (e.g., Kafka, NATS) for better decoupling between services.
- [ ]  Integrate a Configuration Center (e.g., Consul, etcd) for dynamic configuration management.
- [ ]  Provide Containerized Deployment options (Dockerfile, docker-compose.yaml).
- [ ]  Implement Hooks/Callbacks for extending functionality at various stages of the pipeline.

## **References**

- [OpenAI - Hello GPT-4o](https://openai.com/index/hello-gpt-4o/)
- [Medium - The Differences Between ASR and TTS](https://medium.com/@artificial--intelligence/the-differences-between-asr-and-tts-c85a08269c98#:~:text=We%20are%20familiar%20with%20the,analogous%20to%20the%20human%20mouth.)

## **Contributing**

We welcome contributions of any kind! Please read [CONTRIBUTING.md](https://gemini.google.com/app/CONTRIBUTING.md) (if available, otherwise follow standard GitHub practices) for more information.

- **Reporting Issues**: If you find a bug or have a feature suggestion, please submit an issue on GitHub.
- **Contributing Code**: Fork the repository, make your changes on a separate branch, and submit a Pull Request.

## **License**

`voiceflow` is licensed under the [Apache License 2.0](https://gemini.google.com/app/LICENSE).

## **Acknowledgements**

Thank you to all the developers who have contributed to this project!

<img src="https://contrib.rocks/image?repo=telepace/voiceflow" />
