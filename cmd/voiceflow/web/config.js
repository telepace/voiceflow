// config.js

// 加载环境变量
function loadEnvironmentVariables() {
    // 从 .env 文件或环境变量中获取端口
    const port = process.env.VOICEFLOW_SERVER_PORT || '18080';
    
    // 将配置暴露给全局 window 对象
    window.VOICEFLOW_SERVER_PORT = port;
}

// 页面加载时执行
document.addEventListener('DOMContentLoaded', loadEnvironmentVariables); 