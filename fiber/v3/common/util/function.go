package util

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/GoFurry/awesome-go-template/fiber/v3/roof/env"
	"github.com/bwmarrin/snowflake"
)

// clusterId 雪花算法的分布式节点编号
var clusterId, _ = snowflake.NewNode(int64(env.GetServerConfig().ClusterId))

// GenerateId 雪花算法生成新 ID
func GenerateId() int64 {
	return clusterId.Generate().Int64()
}

// GenerateRandomCode 生成随机验证码
func GenerateRandomCode(length int) string {
	//rand.Seed(time.Now().UnixNano())
	letters := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	randCode := ""
	for i := 1; i <= length; i++ {
		randCode += string(letters[rand.Intn(len(letters))])
	}
	return randCode
}

// GetDateFormatStr 格式化时间
func GetDateFormatStr(format string, date time.Time) string {
	dataString := date.Format(format)
	return dataString
}

// String2Int 字符串转 int
func String2Int(numString string) (int, error) {
	if strings.TrimSpace(numString) == "" {
		return 0, errors.New("字符串不能为0")
	}
	id, err := strconv.Atoi(numString)
	return id, err
}

// Int642String int64 转字符串
func Int642String(i64 int64) string { return strconv.FormatInt(i64, 10) }

// Int2String int 转字符串
func Int2String(i int) string { return fmt.Sprintf("%d", i) }

// String2Int64 字符串转 int64
func String2Int64(numString string) (int64, error) {
	if strings.TrimSpace(numString) == "" {
		return 0, errors.New("参数不能为空")
	}
	id, parseErr := strconv.ParseInt(strings.TrimSpace(numString), 10, 64)
	return id, parseErr
}
