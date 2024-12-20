// script.js
// 获取 WebSocket URL
const ws = new WebSocket(WEBSOCKET_URL);

ws.onopen = () => {
    console.log('WebSocket 连接已建立');
    // 在聊天窗口添加提示信息
    appendSystemMessage('提示：您可以长按麦克风按钮 & 长按 键盘 V 进行录音');
};

ws.onmessage = function(event) {
    if (typeof event.data === 'string') {
        const response = JSON.parse(event.data);
        console.log('收到 WebSocket 响应:', response);
        
        if (response.type) {
            // 处理带有 type 字段的消息（语音识别等）
            switch(response.type) {
                case 'audio_stored':
                    appendAudioMessage('你', response.audio_url);
                    break;
                    
                case 'recognition_complete':
                    appendMessage('你', response.text);
                    break;
                    
                case 'recognition_error':
                    appendSystemMessage(`识别错误: ${response.error}`);
                    break;
                    
                case 'tts_complete':
                    // 移除"正在生成语音..."的系统消息
                    const systemMessages = document.querySelectorAll('.message.system');
                    systemMessages.forEach(msg => {
                        if (msg.textContent === '正在生成语音...') {
                            msg.remove();
                        }
                    });
                    
                    // 添加 AI 的文本和音频消息
                    appendMessage('AI', response.text);
                    appendAudioMessage('AI', response.audio_url);
                    break;
                    
                default:
                    console.log('Unknown message type:', response.type);
            }
        }
    }
};

ws.onerror = (error) => {
    console.error('WebSocket 错误:', error);
};

// 添加重连逻辑
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;

ws.onclose = (event) => {
    console.log('WebSocket connection closed:', event);
    
    if (reconnectAttempts < maxReconnectAttempts) {
        reconnectAttempts++;
        const timeout = Math.min(1000 * Math.pow(2, reconnectAttempts), 10000);
        
        appendSystemMessage(`连接已断开，${timeout/1000}秒后尝试重新连接...`);
        
        setTimeout(() => {
            ws = new WebSocket(WEBSOCKET_URL);
            // 重新绑定事件处理器
        }, timeout);
    } else {
        appendSystemMessage('连接已断开，请刷新页面重试');
    }
};

const chatWindow = document.getElementById('chat-window');
const textInput = document.getElementById('text-input');
const sendTextBtn = document.getElementById('send-text-btn');
const recordVoiceBtn = document.getElementById('record-voice-btn');
const uploadAudioBtn = document.getElementById('upload-audio-btn');
const uploadAudioInput = document.getElementById('upload-audio-input');

sendTextBtn.addEventListener('click', () => {
    const text = textInput.value.trim();
    if (text) {
        appendMessage('你', text);
        sendTextMessage(text);
        textInput.value = '';
    }
});

textInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        sendTextBtn.click();
    }
});

let mediaRecorder;
let audioChunks = [];
let isRecording = false;
let mediaStream = null;

// 添加会话 ID 生成函数
function generateSessionId() {
    return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
}

let currentSessionId = null;

function startRecording() {
    if (isRecording) return;
    
    // 生成新的会话 ID
    currentSessionId = generateSessionId();
    
    // 发送开始信号
    ws.send(JSON.stringify({
        type: "audio_start",
        session_id: currentSessionId
    }));
    
    navigator.mediaDevices.getUserMedia({ audio: true })
        .then(stream => {
            isRecording = true;
            mediaStream = stream;
            recordVoiceBtn.classList.add('recording');

            mediaRecorder = new MediaRecorder(stream);
            const timeslice = 250;

            mediaRecorder.start(timeslice);

            mediaRecorder.ondataavailable = e => {
                if (e.data && e.data.size > 0) {
                    // 直接发送二进制数据
                    ws.send(e.data);
                }
            };

            mediaRecorder.onstop = () => {
                isRecording = false;
                recordVoiceBtn.classList.remove('recording');

                // 停止所有音频轨道
                mediaStream.getTracks().forEach(track => track.stop());
                mediaStream = null;

                // 发送结束信号
                ws.send(JSON.stringify({
                    type: "audio_end",
                    session_id: currentSessionId
                }));
                
                currentSessionId = null;
            };
        })
        .catch(err => {
            console.error('获取麦克风权限失败:', err);
            appendSystemMessage('错误：无法访问麦克风');
        });
}

function sendAudioChunk(audioBlob, sessionId) {
    const reader = new FileReader();
    reader.onload = () => {
        // 发送二进制数据前，先发送元数据
        ws.send(JSON.stringify({
            type: 'audio_metadata',
            session_id: sessionId,
            is_start: true
        }));
        
        // 然后发送音频数据
        ws.send(reader.result);
    };
    reader.readAsArrayBuffer(audioBlob);
}

function stopRecording() {
    if (mediaRecorder && isRecording) {
        mediaRecorder.stop();
        // 发送结束信号时包含会话 ID
        ws.send(JSON.stringify({ 
            type: 'audio_end',
            session_id: currentSessionId 
        }));
        currentSessionId = null;
    }
}

// 监听录音按钮的长按事件
recordVoiceBtn.addEventListener('mousedown', startRecording);
recordVoiceBtn.addEventListener('mouseup', stopRecording);
recordVoiceBtn.addEventListener('mouseleave', stopRecording);

// 兼容移动设备的触摸事件
recordVoiceBtn.addEventListener('touchstart', (e) => {
    e.preventDefault(); // 防止触发点击事件
    startRecording();
});
recordVoiceBtn.addEventListener('touchend', stopRecording);
recordVoiceBtn.addEventListener('touchcancel', stopRecording);

// 键盘事件监听（可选，如果需要按住 'V' 键录音）
document.addEventListener('keydown', (e) => {
    if (e.key.toLowerCase() === 'v' && !isRecording) {
        startRecording();
    }
});

document.addEventListener('keyup', (e) => {
    if (e.key.toLowerCase() === 'v' && isRecording) {
        stopRecording();
    }
});

uploadAudioBtn.addEventListener('click', () => {
    uploadAudioInput.click();
});

uploadAudioInput.addEventListener('change', () => {
    const file = uploadAudioInput.files[0];
    if (file) {
        sendAudioMessage(file);
    }
});

function sendTextMessage(text) {
    // 显示发送的消息
    appendMessage('你', text);
    
    // ��过 WebSocket 发送文字消息
    ws.send(JSON.stringify({ 
        text: text,
        require_tts: true
    }));
    
    // 可以添加一个加载提示
    appendSystemMessage('正在生成语音...');
}

function sendAudioMessage(audioBlob) {
    // 在聊天窗口添加可播放的语音消息
    appendAudioMessage('你', URL.createObjectURL(audioBlob));

    // 将 audioBlob 转换为 ArrayBuffer 并通过 WebSocket 发送
    const reader = new FileReader();
    reader.onload = () => {
        ws.send(reader.result);
    };
    reader.readAsArrayBuffer(audioBlob);
}

let partialMessageDiv;

function updatePartialMessage(user, text) {
    if (!partialMessageDiv) {
        partialMessageDiv = document.createElement('div');
        partialMessageDiv.classList.add('message');

        const userSpan = document.createElement('span');
        userSpan.classList.add('user');
        userSpan.textContent = `${user}: `;

        const textSpan = document.createElement('span');
        textSpan.classList.add('partial-text');

        partialMessageDiv.appendChild(userSpan);
        partialMessageDiv.appendChild(textSpan);
        chatWindow.appendChild(partialMessageDiv);
    }

    const textSpan = partialMessageDiv.querySelector('.partial-text');
    textSpan.textContent = text;
    chatWindow.scrollTop = chatWindow.scrollHeight;
}

// 当录音结束时，清除部分消息
function clearPartialMessage() {
    if (partialMessageDiv) {
        partialMessageDiv.remove();
        partialMessageDiv = null;
    }
}

// 修改录音停止的函数，添加清除部分消息的逻辑
function stopRecording() {
    if (mediaRecorder && isRecording) {
        mediaRecorder.stop();
        clearPartialMessage();
    }
}


// 当最终文本到达时，替换部分转录文本
function appendMessage(user, text) {
    if (partialMessageDiv) {
        partialMessageDiv.remove();
        partialMessageDiv = null;
    }
    // 继续现有代码，添加消息
    const messageDiv = document.createElement('div');
    messageDiv.classList.add('message');

    const userSpan = document.createElement('span');
    userSpan.classList.add('user');
    userSpan.textContent = `${user}: `;

    const textSpan = document.createElement('span');
    textSpan.textContent = text;

    messageDiv.appendChild(userSpan);
    messageDiv.appendChild(textSpan);
    chatWindow.appendChild(messageDiv);
    chatWindow.scrollTop = chatWindow.scrollHeight;
}

function appendAudioMessage(user, audioUrl) {
    const messageDiv = document.createElement('div');
    messageDiv.classList.add('message');

    const userSpan = document.createElement('span');
    userSpan.classList.add('user');
    userSpan.textContent = `${user}: `;

    const audio = document.createElement('audio');
    audio.src = audioUrl;
    audio.controls = true;
    
    // 添加音频加载错误处理
    audio.onerror = function() {
        console.error('音频加载失败:', audioUrl);
        appendSystemMessage('音频加载失败，请重试');
    };
    
    // 添加音频加载成功处理
    audio.onloadeddata = function() {
        console.log('音频加载成功:', audioUrl);
    };

    messageDiv.appendChild(userSpan);
    messageDiv.appendChild(audio);
    chatWindow.appendChild(messageDiv);
    chatWindow.scrollTop = chatWindow.scrollHeight;
}

function appendSystemMessage(text) {
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message system';
    messageDiv.textContent = text;
    chatWindow.appendChild(messageDiv);
    chatWindow.scrollTop = chatWindow.scrollHeight;
}