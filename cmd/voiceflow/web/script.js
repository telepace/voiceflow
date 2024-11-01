// script.js
// 获取 WebSocket URL
const ws = new WebSocket(WEBSOCKET_URL);

ws.onopen = () => {
    console.log('WebSocket 连接已建立');
    // 在聊天窗口添加提示信息
    appendSystemMessage('提示：您可以长按麦克风按钮 & 长按 键盘 V 进行录音');
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.text) {
        appendMessage('助手', data.text);
    } else if (data.audio_url) {
        appendAudioMessage('助手', data.audio_url);
    }
};

ws.onerror = (error) => {
    console.error('WebSocket 错误:', error);
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

function startRecording() {
    if (isRecording) return;

    navigator.mediaDevices.getUserMedia({ audio: true })
        .then(stream => {
            isRecording = true;
            mediaStream = stream;
            recordVoiceBtn.classList.add('recording');

            mediaRecorder = new MediaRecorder(stream);
            mediaRecorder.start();

            mediaRecorder.ondataavailable = e => {
                audioChunks.push(e.data);
            };

            mediaRecorder.onstop = () => {
                const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
                audioChunks = [];
                sendAudioMessage(audioBlob);
                isRecording = false;
                recordVoiceBtn.classList.remove('recording');

                // 停止所有音频轨道，释放麦克风
                mediaStream.getTracks().forEach(track => track.stop());
                mediaStream = null;
            };
        })
        .catch(err => {
            console.error('麦克风访问错误:', err);
        });
}

function stopRecording() {
    if (mediaRecorder && isRecording) {
        mediaRecorder.stop();
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
    // 通过 WebSocket 发送文字消息
    ws.send(JSON.stringify({ text: text }));
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

function appendMessage(user, text) {
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

    messageDiv.appendChild(userSpan);
    messageDiv.appendChild(audio);
    chatWindow.appendChild(messageDiv);
    chatWindow.scrollTop = chatWindow.scrollHeight;
}

function appendSystemMessage(text) {
    const messageDiv = document.createElement('div');
    messageDiv.classList.add('message', 'system-message');

    const textSpan = document.createElement('span');
    textSpan.textContent = text;

    messageDiv.appendChild(textSpan);
    chatWindow.appendChild(messageDiv);
    chatWindow.scrollTop = chatWindow.scrollHeight;
}