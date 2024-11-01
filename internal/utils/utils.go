// Copyright © 2023 telepace open source community. All rights reserved.
// Licensed under the MIT License (the "License");
// you may not use this file except in compliance with the License.

package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func GenerateSessionID() string {
	// 根据当前时间戳或其他随机数生成会话 ID
	return "session-id"
}

func HashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
